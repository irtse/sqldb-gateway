package domain

import (
	"fmt"
	"errors"
	"strings"
	"reflect"
	"encoding/json"
	tool "sqldb-ws/lib"
	domain "sqldb-ws/lib/domain/service"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
/*
	Domain is defined as the DDD patterns will suggest it.
	It's the specialized part of the API, it concive particular behavior on datas (in our cases, particular Root DB declares in entity)
	
	Main Service at a Domain level, it follows the DOMAIN ITF from tool. 
	Domain interact at a "Model" level with generic and abstract infra services. 
	Main service give the main process to interact with Infra. 
*/
type MainService struct {
	name                string
	User				string
	Shallowed			bool
	SuperAdmin			bool
	RawView				bool
	Super				bool
	isGenericService    bool
	Specialization		bool
	PermService			tool.InfraServiceItf
	Db					*conn.Db
}
// generate a new domain controller. 
func Domain(superAdmin bool, user string, isGenericService bool) *MainService {
	return &MainService{ 
		isGenericService: isGenericService, // generic specialized service is CustomService
		SuperAdmin: superAdmin, // carry the security level of the "User" or an inner action
		User : user, // current user... 
		Specialization : true, // define if a specialized treatment is await. 
	}
}
// Main accessor defined by DomainITF interface
func (d *MainService) GetPermission() tool.InfraServiceItf { return d.PermService }
func (d *MainService) SetIsCustom(isCustom bool) { d.isGenericService = isCustom }
func (d *MainService) GetUser() string { return d.User }
func (d *MainService) IsSuperAdmin() bool { return d.SuperAdmin }
func (d *MainService) IsSuperCall() bool { return d.Super && d.SuperAdmin }
func (d *MainService) IsShallowed() bool { return d.Shallowed }
func (d *MainService) IsRawView() bool { return d.RawView }
func (d *MainService) GetDB() tool.DbITF { return d.Db }

// Infra func caller with admin view && superadmin right (not a structured view made around data for view reason)
func (d *MainService) SuperCall(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	params[tool.RootRawView]="enable"
	params[tool.RootSuperCall]="enable"
	return Domain(true, d.User, d.isGenericService).call(false, params, record, method, true, funcName, args...)
}
// Infra func caller with current option view and user rights.
func (d *MainService) Call(params tool.Params, record tool.Record, method tool.Method, auth bool, funcName string, args... interface{}) (tool.Results, error) {
	return d.call(true, params, record, method, auth, funcName, args...)
}
// Main process to call an Infra function
func (d *MainService) call(postTreat bool, params tool.Params, record tool.Record, method tool.Method, auth bool, funcName string, args... interface{}) (tool.Results, error) {
	var service tool.InfraServiceItf // generate an empty var for a casual infra service ITF (interface) to embedded any service.
	res := tool.Results{}
	if adm, ok := params[tool.RootSuperCall]; ok && adm == "enable" { d.Super = true } // set up admin view
	if adm, ok := params[tool.RootRawView]; ok && adm == "enable" { d.RawView = true } // set up admin view
	if shallow, ok := params[tool.RootShallow]; ok && shallow == "enable" { d.Shallowed = true }  // set up shallow option (lighter version of results)
	if tablename, ok := params[tool.RootTableParam]; ok { // retrieve tableName in query (not optionnal)
		var specializedService tool.SpecializedService
		if d.Specialization {
			specializedService = &domain.CustomService{}
			if !d.isGenericService { specializedService = domain.SpecializedService(tablename) }
			specializedService.SetDomain(d)
			for _, exception := range entities.PERMISSIONEXCEPTION {
				if tablename == exception.Name { auth = false; break }
			}
		}
		d.Db = conn.Open() // open base
		defer d.Db.Conn.Close() // close when finished
		// load the highest entity avaiable Table level.
		table := infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), params, record, method)
		delete(params, tool.RootTableParam)
		service=table
		tablename = strings.ToLower(tablename)
		d.PermService = infrastructure.Permission(d.Db, 
			d.SuperAdmin, 
			d.User,
			params, 
			record,
			method)
		if res, err := d.PermService.(*infrastructure.PermissionInfo).Get(); res != nil && err == nil { 
			d.PermService.(*infrastructure.PermissionInfo).GeneratePerms(res) 
		}
		table.PermService=d.PermService.(*infrastructure.PermissionInfo)
		if rowName, ok := params[tool.RootRowsParam]; ok { // rows override columns
			if tablename == tool.ReservedParam { 
				return res, errors.New("can't load table as " + tool.ReservedParam) 
			}
			if auth {
			   	if _, ok := d.PermService.Verify(tablename); !ok { 
					return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
			    }
			}
			params[tool.SpecialIDParam]=strings.ToLower(rowName) 
			delete(params, tool.RootRowsParam)
			if params[tool.SpecialIDParam] == tool.ReservedParam { delete(params, tool.SpecialIDParam) }
			service = table.TableRow(specializedService)
			service.SetAuth(auth)
			res, err := d.invoke(service, funcName, args...)
			if specializedService != nil && postTreat {
				if dest_id, ok := params[tool.RootDestTableIDParam]; ok {
					return specializedService.PostTreatment(res, tablename, dest_id), nil
				}
				return specializedService.PostTreatment(res, tablename), nil
			}
			if (d.PermService.(*infrastructure.PermissionInfo).PartialResults != "") {
				return res, errors.New(d.PermService.(*infrastructure.PermissionInfo).PartialResults)
			}
			return res, err
		}
		if !d.SuperAdmin { 
			return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
		}
		if col, ok := params[tool.RootColumnsParam]; ok { 
			if tablename == tool.ReservedParam { 
				d.Db.Conn.Close()
				return res, errors.New("can't load table as " + tool.ReservedParam) 
			}
			params[tool.RootColumnsParam]=strings.ToLower(col)
			service = table.TableColumn() 
		}
		service.SetAuth(auth)
		return d.invoke(service, funcName, args...)
	}
	return res, errors.New("no service avaiblable")
}
func (d *MainService) invoke(service tool.InfraServiceItf, funcName string, args... interface{}) (tool.Results, error) {
    var err error
	res := tool.Results{}
	clazz := reflect.ValueOf(service).MethodByName(funcName)
	if !clazz.IsValid() { return res, errors.New("not implemented <"+ funcName +"> (invalid)") }
	if clazz.IsZero() { return res, errors.New("not implemented <"+ funcName +"> (zero)") }
	var values []reflect.Value
	if len(args) > 0 {
		vals := []reflect.Value {}
		for _, arg := range args { vals = append(vals, reflect.ValueOf(arg)) }
		values = clazz.Call(vals)
	} else { values = clazz.Call(nil) }
	if len(values) > 0 { res = values[0].Interface().(tool.Results) }
	if len(values) > 1 { 
		if values[1].Interface() == nil { err = nil
		} else { err = values[1].Interface().(error) } 
	}
	return res, err
}

type View struct {
	Name  		 string 					`json:"name"`
	SchemaName   string 					`json:"schema_name"`
	Description  string 					`json:"description"`
	Path		 string 					`json:"link_path"`
	Schema		 tool.Record 				`json:"schema"`
	Items		 []tool.Record 				`json:"items"`
	Actions		 []map[string]interface{} 	`json:"actions"`
}

type ViewItem struct {
	Path 	   string					 	`json:"link_path"`
	Values 	   map[string]interface{} 	    `json:"values"`
	DataPaths  string				        `json:"data_path"`
	ValuePaths map[string]string			`json:"values_path"`
}

func (d *MainService) PostTreat(results tool.Results, tableName string, shallow bool, 
	                            additonnalRestriction ...string) tool.Results {
	cols := map[string]entities.SchemaColumnEntity{}
	sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
	sqlFilter += entities.DBSchema.Name + " WHERE name=" + conn.Quote(tableName) + ")"
	// retrive all fields from schema...
	params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
		                   tool.RootRowsParam: tool.ReservedParam, }
	schemas, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
	if err != nil || len(schemas) == 0 { return tool.Results{} }
	var view View
	if !d.IsShallowed() && !d.IsRawView() {
		schemes := map[string]interface{}{}
		for _, r := range schemas {
			var scheme entities.SchemaColumnEntity
			var shallowField entities.ShallowSchemaColumnEntity
			b, _ := json.Marshal(r)
			json.Unmarshal(b, &scheme)
			cols[scheme.Name]=scheme
			json.Unmarshal(b, &shallowField)
			if scheme.Link != "" {
				shallowField.LinkPath = "/" + scheme.Link + "?rows=all"
				if scheme.LinkView != "" { shallowField.LinkPath += "&" + tool.RootColumnsParam + "=" + scheme.LinkView  }
				if scheme.LinkOrder != "" { shallowField.LinkPath += "&" + tool.RootOrderParam + "=" + scheme.LinkOrder  }
			}
			schemes[scheme.Name]=shallowField
		}
		view = View{ Name : tableName, Description : tableName + " datas", 
	                  Path : "", 
					  Schema : schemes,
					  SchemaName: tableName, 
					  Actions : []map[string]interface{}{},
					  Items : []tool.Record{} }	
		res := tool.Results{} 
		for _, record := range results { 
			rec := d.PostTreatRecord(record, tableName, cols, shallow)
			if rec == nil { continue }
			view.Items = append(view.Items, rec)
			if shallow { break; }
		}
		r := tool.Record{}
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &r)
		res = append(res, r)
		return res
	} else { return results }
}

func (d *MainService) PostTreatRecord(record tool.Record, tableName string, 
									  cols map[string]entities.SchemaColumnEntity, shallow bool) tool.Record {
	if d.IsShallowed() {
		if _, ok := record[entities.NAMEATTR]; ok {
			return tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR] }
		} else { return record }
	} else {
		if d.IsRawView() { return record } // if admin view avoid.
		vals := map[string]interface{}{}
		contentPaths := map[string]string{}
		datapath := ""
		if !shallow { vals[tool.SpecialIDParam]=fmt.Sprintf("%v", record[tool.SpecialIDParam]) }
		for _, field := range cols {
			if d.Db.GetSQLView() != "" && !strings.Contains(d.Db.GetSQLView(), field.Name){ continue }
			if strings.Contains(field.Name, entities.DBSchema.Name) && !shallow { 
				dest, ok := record[entities.RootID("dest_table")]
				id, ok2 := record[field.Name]
				if ok2 && ok && dest != nil && id != nil {
					schemas, err := d.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : id })
					if err != nil || len(schemas) == 0 { continue }
					datapath=d.BuildPath(fmt.Sprintf("%v",schemas[0][entities.NAMEATTR]), fmt.Sprintf("%v", dest))
				}
				continue
			}
			if f, ok:= record[field.Name]; ok && field.Link != "" && f != nil && !shallow { 
				contentPaths[field.Name]=d.BuildPath(field.Link, fmt.Sprintf("%v", f), "shallow=enable")
				continue
			}
			if shallow { vals[field.Name]=nil 
			} else if v, ok:=record[field.Name]; ok { vals[field.Name]=v }
		}
		view := ViewItem{ Values : vals, Path : "", DataPaths :  datapath, ValuePaths : contentPaths, }
		var newRec tool.Record
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &newRec)
		return newRec
	}
}

func (d *MainService) BuildPath(tableName string, rows string, extra... string) string {
		path := "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + rows
		for _, ext := range extra { path += "&" + ext }
		return path
}
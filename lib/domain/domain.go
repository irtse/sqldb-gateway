package domain

import (
	"fmt"
	"slices"
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
	Empty               bool
	LowerRes            bool
	Method				tool.Method
	Params 				tool.Params
	Perms				map[string]map[string]Perms
	notAllowedFields	[]string
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
func (d *MainService) SetIsCustom(isCustom bool) { d.isGenericService = isCustom }
func (d *MainService) SetLowerRes(empty bool) { d.LowerRes = empty }
func (d *MainService) SetEmpty(empty bool) { d.Empty = empty }
func (d *MainService) GetEmpty() bool { return d.Empty }
func (d *MainService) GetUser() string { return d.User }
func (d *MainService) IsSuperAdmin() bool { return d.SuperAdmin }
func (d *MainService) IsSuperCall() bool { return d.Super && d.SuperAdmin }
func (d *MainService) IsShallowed() bool { return d.Shallowed }
func (d *MainService) GetParams() tool.Params { return d.Params }
func (d *MainService) GetDB() tool.DbITF { return d.Db }

// Infra func caller with admin view && superadmin right (not a structured view made around data for view reason)
func (d *MainService) SuperCall(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	params[tool.RootRawView]="enable"
	params[tool.RootSuperCall]="enable"
	return Domain(true, d.User, d.isGenericService).call(params, record, method, funcName, args...)
}
func (d *MainService) PermsSuperCall(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	params[tool.RootRawView]="enable"
	return Domain(true, d.User, d.isGenericService).call(params, record, method, funcName, args...)
}
// Infra func caller with current option view and user rights.
func (d *MainService) Call(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	return d.call(params, record, method, funcName, args...)
}
// Main process to call an Infra function
func (d *MainService) call(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	var service tool.InfraServiceItf // generate an empty var for a casual infra service ITF (interface) to embedded any service.
	res := tool.Results{}
	d.Method = method
	d.Params = params
	d.notAllowedFields = []string{}
	if adm, ok := params[tool.RootSuperCall]; ok && adm == "enable" { d.Super = true } // set up admin view
	if shallow, ok := params[tool.RootShallow]; (ok && shallow == "enable") || slices.Contains(tool.EXCEPTION_FUNC, funcName) { d.Shallowed = true }  // set up shallow option (lighter version of results)
	if tablename, ok := params[tool.RootTableParam]; ok { // retrieve tableName in query (not optionnal)
		var specializedService tool.SpecializedService
		if d.Specialization {
			specializedService = &domain.CustomService{}
			if !d.isGenericService { specializedService = domain.SpecializedService(tablename) }
			specializedService.SetDomain(d)
		}
		d.Db = conn.Open() // open base
		defer d.Db.Conn.Close() // close when finished
		d.PermsBuilder()
		if !d.PermsCheck(tablename, "", "", d.Method) {
			return res, errors.New("not authorized to " + method.String() + " " + tablename + " datas")
		}
		// load the highest entity avaiable Table level.
		table := infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), params, record, method)
		delete(params, tool.RootTableParam)
		service=table
		tablename = strings.ToLower(tablename)
		if rowName, ok := params[tool.RootRowsParam]; ok { // rows override columns
			if tablename == tool.ReservedParam { return res, errors.New("can't load table as " + tool.ReservedParam) }
			params[tool.SpecialIDParam]=strings.ToLower(rowName) 
			delete(params, tool.RootRowsParam)
			if params[tool.SpecialIDParam] == tool.ReservedParam { delete(params, tool.SpecialIDParam) }
			service = table.TableRow(specializedService)
			res, err := d.invoke(service, funcName, args...)
			if err != nil { return res, err }
			if specializedService != nil && params[tool.RootRawView] != "enable" && !d.IsSuperCall() {
				if dest_id, ok := params[tool.RootDestTableIDParam]; ok {
					return specializedService.PostTreatment(res, tablename, dest_id), nil
				}
				return specializedService.PostTreatment(res, tablename), nil
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
	Name  		 	string 						`json:"name"`
	SchemaID		int64						`json:"schema_id"`
	SchemaName   	string 						`json:"schema_name"`
	Description  	string 						`json:"description"`
	Path		 	string 						`json:"link_path"`
	Order		 	[]string 					`json:"order"`
	Schema		 	tool.Record 				`json:"schema"`
	Items		 	[]tool.Record 				`json:"items"`
	Actions		 	[]map[string]interface{} 	`json:"actions"`
	Parameters 		[]string 					`json:"parameters"`
}

type ViewItem struct {
	Path 	   	   string					 	`json:"link_path"`
	Values 	   	   map[string]interface{} 	    `json:"values"`
	DataPaths  	   string				        `json:"data_path"`
	ValueShallow   map[string]interface{}		`json:"values_shallow"`
	ValueMany      map[string]tool.Results		`json:"values_many"`
	ValuePathMany  map[string]string			`json:"values_path_many"`
}

func (d *MainService) PostTreat(results tool.Results, tableName string) tool.Results {
	// retrive all fields from schema...
	var view View
	if !d.IsShallowed() {
		schemes, id, order, cols, addAction := d.GetScheme(tableName, false) 
		view = View{ Name : tableName, Description : tableName + " datas",  Path : "", 
					 Schema : schemes, 
					 Order : order,
					 SchemaID: id,
					 SchemaName: tableName, 
					 Actions : []map[string]interface{}{},  Items : []tool.Record{} }	
		res := tool.Results{} 
		for _, record := range results { 
			rec := d.PostTreatRecord(record, tableName, cols, d.Empty)
			if rec == nil { continue }
			view.Items = append(view.Items, rec)
		}
		r := tool.Record{}
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &r)
		r["action_path"] = "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + tool.ReservedParam
		r["actions"]=[]string{}
		for _, meth := range []tool.Method{ tool.SELECT, tool.CREATE, tool.UPDATE, tool.DELETE } {
			if d.PermsCheck(tableName, "", "", meth) || slices.Contains(addAction, meth.Method()) { 
				r["actions"]=append(r["actions"].([]string), meth.Method())
			} else if meth == tool.UPDATE { r["readonly"] = true }
		} 
		res = append(res, r)
		return res
	} else { 
		res := tool.Results{}
		for _, record := range results {
			if n, ok := record[entities.NAMEATTR]; ok {
				label := fmt.Sprintf("%v", n)
				if l, ok2 := record["label"]; ok2 { label = fmt.Sprintf("%v", l) }
				if record[entities.RootID(entities.DBSchema.Name)] != nil { // SCHEMA ? 
					schemas, err := d.Schema(record, true)
					actionPath := "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + tool.ReservedParam
					actions := []string{}
					readonly := false
					if err == nil || len(schemas) > 0 { 
						schema, id, order,  _, addAction := d.GetScheme(schemas[0].GetString(entities.NAMEATTR), false)
						for _, meth := range []tool.Method{ tool.SELECT, tool.CREATE, tool.UPDATE, tool.DELETE } {
							if d.PermsCheck(schemas[0].GetString(entities.NAMEATTR), "", "", meth) || slices.Contains(addAction, meth.Method()) { 
								actions=append(actions, meth.Method())
							} else if meth == tool.UPDATE { readonly = true 
							} else if meth == tool.CREATE && d.Empty { readonly = true }
						} 
						res = append(res, tool.Record{ 
							tool.SpecialIDParam : record[tool.SpecialIDParam],
							entities.NAMEATTR : n,
							"label": label, 
							"order" : order,
							"schema_id" : id,
							"actions" : actions,
							"action_path" : actionPath,
							"readonly" : readonly,
							"link_path" : "/" + tool.MAIN_PREFIX + "/" + schemas[0].GetString(entities.NAMEATTR) + "?rows=" + tool.ReservedParam,
							"schema_name" : schemas[0].GetString(entities.NAMEATTR),
							"schema" : schema, })	
						continue
					}	
				}
				res = append(res, tool.Record{ tool.SpecialIDParam : record[tool.SpecialIDParam], 
					                           entities.NAMEATTR : n, "label": label,})	
			} else { res = append(res, record) }
		}
		return res
	} 
	return results
}

func (d *MainService) PostTreatRecord(record tool.Record, tableName string,  cols map[string]entities.SchemaColumnEntity, shallow bool) tool.Record {
		vals := map[string]interface{}{}
		shallowVals := map[string]interface{}{}
		manyPathVals := map[string]string{}
		manyVals := map[string]tool.Results{}
		datapath := ""
		if !shallow { vals[tool.SpecialIDParam]=fmt.Sprintf("%v", record[tool.SpecialIDParam]) }
		for _, field := range cols {
			if strings.Contains(field.Name, entities.DBSchema.Name) { 
				dest, ok := record[entities.RootID("dest_table")]
				id, ok2 := record[field.Name]
				if ok2 && ok && dest != nil && id != nil {
					schemas, err := d.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : id }, true)
					if err != nil || len(schemas) == 0 { continue }
					if dest != nil {
						datapath=d.BuildPath(fmt.Sprintf("%v",schemas[0][entities.NAMEATTR]), fmt.Sprintf("%v", dest))			
					}
				}
			}
			if f, ok:= record[field.Name]; ok && field.Link != "" && f != nil && !shallow && !strings.Contains(field.Type, "many") { 
				params := tool.Params{ tool.RootTableParam : field.Link, tool.RootRowsParam: fmt.Sprintf("%v", f), tool.RootShallow : "enable" }
				r, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(r) == 0 { continue }
				shallowVals[field.Name]=r[0]
				continue
			}
			if field.Link != "" && !shallow && !d.LowerRes && strings.Contains(field.Type, "manytomany") { 
				params := tool.Params{ tool.RootTableParam : field.Link, tool.RootRowsParam: tool.ReservedParam, tool.RootShallow : "enable",
									   entities.RootID(tableName) : record.GetString(tool.SpecialIDParam), }
				r, err := d.Call( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(r) == 0 { continue }
				ids := []string{}
				for _, r2 := range r {
					for field2, _ := range r2 {
						if !strings.Contains(field2, tableName) && field2 != "id" && strings.Contains(field2, "_id") {
							if !slices.Contains(ids, strings.Replace(field2, "_id", "", -1)) {
								ids = append(ids, strings.Replace(field2, "_id", "", -1))
							}
						}
					}
				}
				for _, id := range ids {
					params = tool.Params{ tool.RootTableParam : id, tool.RootRowsParam: tool.ReservedParam, 
						                  tool.RootShallow : "enable", tableName + "_id": record.GetString(tool.SpecialIDParam) }
					sqlFilter := "id IN (SELECT " + id + "_id FROM " + field.Link + " WHERE " + tableName + "_id = " + record.GetString(tool.SpecialIDParam) + " )"
					r, err = d.Call( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
					if err != nil || len(r) == 0 { continue }
					if _, ok := manyVals[field.Name]; !ok { manyVals[field.Name] = tool.Results{} }
					manyVals[field.Name]= append(manyVals[field.Name], r...)
				}
				continue
			}
			if field.Link != "" && !shallow && strings.Contains(field.Type, "onetomany") && !d.LowerRes { 
				manyPathVals[field.Name] = "/" + tool.MAIN_PREFIX + "/" + field.Link + "?" + tool.RootRowsParam + "=" + tool.ReservedParam + "&" + tableName + "_id=" + record.GetString(tool.SpecialIDParam)
				continue
			}
			if shallow { vals[field.Name]=nil } else if v, ok:=record[field.Name]; ok { vals[field.Name]=v }
		}
		view := ViewItem{ Values : vals, Path : "", DataPaths :  datapath, ValueShallow : shallowVals, ValueMany: manyVals, ValuePathMany: manyPathVals, }
		var newRec tool.Record
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &newRec)
		return newRec
}



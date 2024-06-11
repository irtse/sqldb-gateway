package domain

import (
	"time"
	"slices"
	"errors"
	"strings"
	"reflect"
	"encoding/json"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	domain "sqldb-ws/lib/domain/service"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
/*
	Domain is defined as the DDD patterns will suggest it.
	It's the specialized part of the API, it concive particular behavior on datas (in our cases, particular Root DB declares in entity)
	Main Service at a Domain level, it follows the DOMAIN ITF from schserv. 
	Domain interact at a "Model" level with generic and abstract infra services. 
	Mai	"fmt"
n service give the main process to interact with Infra. 
*/
var EXCEPTION_FUNC = []string{"Count"}
type MainService struct {
	AutoLoad			bool
	User				string
	Shallowed			bool
	SuperAdmin			bool
	ExternalSuperAdmin	bool
	RawView				bool
	Super				bool
	isGenericService    bool
	Specialization		bool
	Empty               bool
	LowerRes            bool
	Own 				bool
	Method				utils.Method
	Params 				utils.Params
	Perms				map[string]map[string]Perms
	notAllowedFields	[]string
	Service 			infrastructure.InfraServiceItf
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
func (d *MainService) GetAutoload() bool { return d.AutoLoad }
func (d *MainService) SetExternalSuperAdmin(external bool) { d.ExternalSuperAdmin = external }
func (d *MainService) GetMethod() utils.Method { return d.Method }
func (d *MainService) GetEmpty() bool { return d.Empty }
func (d *MainService) GetUser() string { return d.User }
func (d *MainService) IsSuperAdmin() bool { return d.SuperAdmin }
func (d *MainService) IsSuperCall() bool { return d.Super && d.SuperAdmin || d.ExternalSuperAdmin }
func (d *MainService) IsShallowed() bool { return d.Shallowed }
func (d *MainService) IsOwn() bool { return d.Own }
func (d *MainService) SetOwn(b bool) { d.Own = b }
func (d *MainService) GetParams() utils.Params { return d.Params }
func (d *MainService) GetDb() *conn.Db { return d.Db }
// Infra func caller with admin view && superadmin right (not a structured view made around data for view reason)
func (d *MainService) SuperCall(params utils.Params, record utils.Record, method utils.Method, args... interface{}) (utils.Results, error) {
	params[utils.RootRawView]="enable"; params[utils.RootSuperCall]="enable"
	d2 := Domain(true, d.User, d.isGenericService)
	d2.ExternalSuperAdmin = d.ExternalSuperAdmin
	return d2.call(params, record, method, args...)
}
func (d *MainService) PermsSuperCall(params utils.Params, record utils.Record, method utils.Method, args... interface{}) (utils.Results, error) {
	params[utils.RootRawView]="enable"
	d2 := Domain(true, d.User, d.isGenericService)
	d2.ExternalSuperAdmin = d.ExternalSuperAdmin
	return d2.call(params, record, method, args...)
}

func (d *MainService) SpecialSuperCall(params utils.Params, record utils.Record, method utils.Method, args... interface{}) (utils.Results, error) {
	params[utils.RootRawView]="enable"
	d2 := Domain(true, d.User, d.isGenericService)
	d2.Own = d.IsOwn()
	d2.ExternalSuperAdmin = d.ExternalSuperAdmin
	return d2.call(params, record, method, args...)
}
// Infra func caller with current option view and user rights.
func (d *MainService) Call(params utils.Params, record utils.Record, method utils.Method, args... interface{}) (utils.Results, error) {
	return d.call(params, record, method, args...)
}
// Main process to call an Infra function
func (d *MainService) call(params utils.Params, record utils.Record, method utils.Method, args... interface{}) (utils.Results, error) {
	d.Method = method; d.notAllowedFields = []string{}
	if adm, ok := params[utils.RootSuperCall]; ok && adm == "enable" { d.Super = true } // set up admin view
	if shallow, ok := params[utils.RootShallow]; (ok && shallow == "enable") { d.Shallowed = true }  // set up shallow option (lighter version of results)
	if tablename, ok := params[utils.RootTableParam]; ok { // retrieve tableName in query (not optionnal)
		tablename := schserv.GetTablename(tablename)
		if raw, ok := params[utils.RootRawView]; (!ok || raw != "enable") { d.ClearDeprecatedDatas(tablename) }
		var specializedService utils.SpecializedServiceITF
		if d.Specialization {
			specializedService = &domain.CustomService{}
			if !d.isGenericService { specializedService = domain.SpecializedService(tablename) }
			specializedService.SetDomain(d)
		}
		if d.Db != nil { d.Db.Close() } // open base
		d.Db = conn.Open(); 
		defer d.Db.Close() 
		if !d.SuperAdmin && !d.PermsCheck(tablename, "", "", d.Method) && !d.AutoLoad {
			return utils.Results{}, errors.New("not authorized to " + method.String() + " " + tablename + " data")
		}
		// load the highest entity avaiable Table level.
		table := infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), record)
		d.Service=table
		delete(params, utils.RootTableParam)
		tablename = strings.ToLower(tablename)
		if rowName, ok := params[utils.RootRowsParam]; ok { // rows override columns
			if id, ok := params[utils.SpecialIDParam]; ok { params[utils.SpecialSubIDParam]=id }
			if strings.ToLower(rowName) != utils.ReservedParam { params[utils.SpecialIDParam]=strings.ToLower(rowName) }
			delete(params, utils.RootRowsParam)
			if params[utils.SpecialIDParam] == "" || params[utils.SpecialIDParam] == utils.ReservedParam || params[utils.SpecialIDParam] == "<nil>" { delete(params, utils.SpecialIDParam) 
			} else if table.Record != nil { table.Record[utils.SpecialIDParam] = params[utils.SpecialIDParam] }
			d.Service = table.TableRow(specializedService)
			utils.ParamsMutex.Lock()
			d.Params = params
			utils.ParamsMutex.Unlock()
			res, err := d.invoke(method.Calling(), args...)
			if err == nil && specializedService != nil && params[utils.RootRawView] != "enable" && !d.Super && !slices.Contains(EXCEPTION_FUNC, method.Calling()) {
				if dest_id, ok := params[utils.RootDestTableIDParam]; ok {
					return specializedService.PostTreatment(res, tablename, dest_id), nil
				}
				return specializedService.PostTreatment(res, tablename), nil
			}
			return res, err
		}
		if !d.SuperAdmin || method == utils.DELETE { 
			return utils.Results{}, errors.New("not authorized to " + method.String() + " " + table.Name + " data") 
		}
		if col, ok := params[utils.RootColumnsParam]; ok { 
			if tablename == utils.ReservedParam { return utils.Results{}, errors.New("can't load table as " + utils.ReservedParam) }
			params[utils.RootColumnsParam]=strings.ToLower(col)
			d.Service = table.TableColumn(specializedService, params[utils.RootColumnsParam]) 
		}
		return d.invoke(method.Calling(), args...)
	}
	return utils.Results{}, errors.New("no service available")
}
func (d *MainService) invoke(funcName string, args... interface{}) (utils.Results, error) {
    var err error
	res := utils.Results{}	
	if d.Service == nil { return res, errors.New("no service available") }
	clazz := reflect.ValueOf(d.Service).MethodByName(funcName)
	if !clazz.IsValid() { return res, errors.New("not implemented <"+ funcName +"> (invalid)") }
	if clazz.IsZero() { return res, errors.New("not implemented <"+ funcName +"> (zero)") }
	var values []reflect.Value
	if len(args) > 0 {
		vals := []reflect.Value {}
		for _, arg := range args { vals = append(vals, reflect.ValueOf(arg)) }
		values = clazz.Call(vals)
	} else { values = clazz.Call(nil) }
	if len(values) > 0 { 
		data, _:= json.Marshal(values[0].Interface())
		json.Unmarshal(data, &res)
	}
	if len(values) > 1 { 
		if values[1].Interface() == nil { err = nil } else { err = values[1].Interface().(error) } 
	}
	return res, err
}

func (d *MainService) ValidateBySchema(data utils.Record, tableName string) (utils.Record, error) {
	if d.Method == utils.DELETE || d.Method == utils.SELECT { return data, nil }
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return data,  errors.New("no schema corresponding to reference") }
	newData := utils.Record{}
	if d.Method == utils.UPDATE {
		for _, field := range schema.Fields {
			if v, ok := data[field.Name]; ok { newData[field.Name]=v }
		}
		return newData, nil
	}
	for _, field := range schema.Fields {
		if field.Required && field.Default == nil {
			if _, ok := data[field.Name]; ok || field.Name == utils.SpecialIDParam || !d.PermsCheck(tableName, field.Name, field.Level, utils.SELECT) { continue }
			if field.Label != "" { return data, errors.New("Missing a required field " + field.Label + " (can't see it ? you probably missing permissions)")
			} else { return data, errors.New("Missing a required field " + field.Name + " (can't see it ? you probably missing permissions)") }
		}
		if v, ok := data[field.Name]; ok { 
			newData[field.Name]=v 
			if field.Name == schserv.FOREIGNTABLEKEY { 
				schema, err :=schserv.GetSchema(v.(string)) 
				if err != nil { newData[schserv.LINKKEY] = schema.ID}
			}
		}
	}
	return newData, nil
}

func  (d *MainService) ClearDeprecatedDatas(tableName string) {
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return }
	if schema.HasField(schserv.STARTKEY) && schema.HasField(schserv.ENDKEY) {
		currentTime := time.Now()
		sqlFilter := "'" + currentTime.Format("2000-01-01") + "' < start_date OR " 
		sqlFilter += "'" + currentTime.Format("2000-01-01") + "' > end_date"
		p := utils.AllParams(tableName)
		p[utils.RootRawView] = "enable"
		d.Call(p, utils.Record{}, utils.DELETE, sqlFilter)
	}
}
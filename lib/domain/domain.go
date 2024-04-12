package domain

import (
	"slices"
	"errors"
	"strings"
	"reflect"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	domain "sqldb-ws/lib/domain/service"
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
	ExternalSuperAdmin	bool
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
func (d *MainService) SetExternalSuperAdmin(external bool) { d.ExternalSuperAdmin = external }
func (d *MainService) GetEmpty() bool { return d.Empty }
func (d *MainService) GetUser() string { return d.User }
func (d *MainService) IsSuperAdmin() bool { return d.SuperAdmin }
func (d *MainService) IsSuperCall() bool { return d.Super && d.SuperAdmin || d.ExternalSuperAdmin }
func (d *MainService) IsShallowed() bool { return d.Shallowed }
func (d *MainService) SetParams(params tool.Params) { d.Params = params }
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
		if d.Db == nil || d.Db.Conn == nil { 
			d.Db = conn.Open() 
			defer d.Db.Close() // close when finished
		} // open base		
		d.PermsBuilder()
		if !d.SuperAdmin && !d.PermsCheck(tablename, "", "", d.Method) {
			return res, errors.New("not authorized to " + method.String() + " " + tablename + " datas")
		}
		// load the highest entity avaiable Table level.
		table := infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), params, record, method)
		delete(params, tool.RootTableParam)
		service=table
		tablename = strings.ToLower(tablename)
		if rowName, ok := params[tool.RootRowsParam]; ok { // rows override columns
			if tablename == entities.DBView.Name {
				if strings.ToLower(rowName) == tool.ReservedParam {
					table = infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), tool.Params{}, record, method)
				} else {
					table = infrastructure.Table(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), tool.Params{
						tool.SpecialIDParam: strings.ToLower(rowName),
					}, record, method)
				}
			} else {
				delete(params, "new")
				if _, ok := params[tool.SpecialIDParam]; !ok {
					params[tool.SpecialIDParam]=strings.ToLower(rowName) 
					delete(params, tool.RootRowsParam)
					if params[tool.SpecialIDParam] == tool.ReservedParam { delete(params, tool.SpecialIDParam) }
				}
			}
			
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
	return res, errors.New("no service available")
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

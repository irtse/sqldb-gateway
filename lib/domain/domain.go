package domain

import (
	"errors"
	"strings"
	"reflect"
	tool "sqldb-ws/lib"
	domain "sqldb-ws/lib/domain/service"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
/*
	Main Controller at a Domain level, it follows the DOMAIN ITF from tool. 
	Domain interact at a "Model" level with generic and abstract infra services. 
	Main service give the main process to interact with Infra. 
*/
type MainService struct {
	name                string
	User				string
	Shallowed			bool
	SuperAdmin			bool
	AdminView			bool
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
func (d *MainService) IsShallowed() bool { return d.Shallowed }
func (d *MainService) IsAdminView() bool { return d.AdminView }
func (d *MainService) GetDB() tool.DbITF { return d.Db }

// Infra func caller with admin view && superadmin right (not a structured view made around data for view reason)
func (d *MainService) SuperCall(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	params[tool.RootAdminView]="enable" // set admin view params to true
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
	if adm, ok := params[tool.RootAdminView]; ok && adm == "enable" && d.SuperAdmin { d.AdminView = true } // set up admin view
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
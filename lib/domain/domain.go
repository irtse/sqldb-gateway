package domain

import (
	"os"
	"errors"
	"strings"
	"reflect"
	tool "sqldb-ws/lib"
	domain "sqldb-ws/lib/domain/service"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
type MainService struct {
	name                string
	User				string
	SuperAdmin			bool
	isGenericService    bool
	PermService			tool.InfraServiceItf
}
func Domain(superAdmin bool, user string, isGenericService bool) *MainService {
	if os.Getenv("automate") == "false" { isGenericService=true }
	return &MainService{ 
		isGenericService: isGenericService, 
		SuperAdmin: superAdmin, 
		User : user, 
	}
}
func (d *MainService) GetPermission() tool.InfraServiceItf { return d.PermService }
func (d *MainService) SetIsCustom(isCustom bool) { d.isGenericService = isCustom }
func (d *MainService) GetUser() string { return d.User }
func (d *MainService) IsSuperAdmin() bool { return d.SuperAdmin }

func (d *MainService) SuperCall(params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	return Domain(true, "", d.isGenericService).call(false, params, record, method, true, funcName, args...)
}

func (d *MainService) Call(params tool.Params, record tool.Record, method tool.Method, auth bool, funcName string, args... interface{}) (tool.Results, error) {
	return d.call(true, params, record, method, auth, funcName, args...)
}

func (d *MainService) call(postTreat bool, params tool.Params, record tool.Record, method tool.Method, auth bool, funcName string, args... interface{}) (tool.Results, error) {
	var service tool.InfraServiceItf
	res := tool.Results{}
	if tablename, ok := params[tool.RootTableParam]; ok {
		var specializedService tool.SpecializedService
		specializedService = &domain.CustomService{}
		if !d.isGenericService { specializedService = domain.SpecializedService(tablename) }
		specializedService.SetDomain(d)
		for _, exception := range entities.PERMISSIONEXCEPTION {
			if tablename == exception.Name { auth = false; break }
		}
		database := conn.Open()
		defer database.Conn.Close()
		table := infrastructure.Table(database, d.SuperAdmin, d.User, strings.ToLower(tablename), params, record, method)
		delete(params, tool.RootTableParam)
		service=table
		tablename = strings.ToLower(tablename)
		if auth {
			d.PermService = infrastructure.Permission(database, 
				d.SuperAdmin, 
				tablename, 
				params, 
				record,
				method)
			if res, err := d.PermService.(*infrastructure.PermissionInfo).Row.Get(); res != nil && err == nil { 
				d.PermService.(*infrastructure.PermissionInfo).GeneratePerms(res) 
			}
		}
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
			if adminView, valid := params[tool.RootAdminView]; valid && adminView == "enable" { service = table.TableRow(specializedService, true)
			} else { service = table.TableRow(specializedService, false) }
			service.SetAuth(auth)
			service.SetPostTreatment(postTreat)
			return d.invoke(service, funcName, args...)
		}
		if !d.SuperAdmin { 
			return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
		}
		if col, ok := params[tool.RootColumnsParam]; ok { 
			if tablename == tool.ReservedParam { 
				database.Conn.Close()
				return res, errors.New("can't load table as " + tool.ReservedParam) 
			}
			params[tool.RootColumnsParam]=strings.ToLower(col)
			service = table.TableColumn() 
		}
		service.SetAuth(auth)
		service.SetPostTreatment(postTreat)
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
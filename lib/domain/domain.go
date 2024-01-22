package domain

import (
	"errors"
	"strings"
	"reflect"
	tool "sqldb-ws/lib"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
type MainService struct {
	name                string
	User				string
	isGenericService    bool
}
func Domain(isGenericService bool) *MainService {
	return &MainService{ isGenericService: isGenericService }
}

func (d *MainService) SafeCall(superAdmin bool, user string, params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	return d.call(superAdmin, user, params, record, method, true, funcName, args...)
}

func (d *MainService) UnSafeCall(user string, params tool.Params, record tool.Record, method tool.Method, funcName string, args... interface{}) (tool.Results, error) {
	return d.call(false, user, params, record, method, false, funcName, args...)
}

func (d *MainService) call(superAdmin bool, user string, params tool.Params, record tool.Record, method tool.Method, auth bool, funcName string, args... interface{}) (tool.Results, error) {
	var service infrastructure.InfraServiceItf
	d.User = user
	res := tool.Results{}
	if tablename, ok := params[tool.RootTableParam]; ok {
		var specializedService tool.SpecializedService
		specializedService = &tool.CustomService{}
		if !d.isGenericService { specializedService = SpecializedService(tablename) }
		specializedService.SetDomain(d)
		database := conn.Open()
		defer database.Conn.Close()
		table := infrastructure.Table(database, superAdmin, user, strings.ToLower(tablename), params, record, method)
		delete(params, tool.RootTableParam)
		service=table
		tablename = strings.ToLower(tablename)
		isRestricted := len(tablename) > 1 && tablename[0:2] == "db" 
		if isRestricted && !superAdmin && method != tool.SELECT { 
			return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
		}
		if rowName, ok := params[tool.RootRowsParam]; ok { // rows override columns
			if tablename == tool.ReservedParam { 
				return res, errors.New("can't load table as " + tool.ReservedParam) 
			}
			perms := infrastructure.Permission(database, 
				                               superAdmin, 
											   tablename, 
											   params, 
											   record,
											   method)
			if _, ok := perms.Verify(tablename); !ok && auth { 
				return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
			}
			params[tool.SpecialIDParam]=strings.ToLower(rowName) 
			delete(params, tool.RootRowsParam)
			if params[tool.SpecialIDParam] == tool.ReservedParam { delete(params, tool.SpecialIDParam) }
			service = table.TableRow(specializedService)
			return d.invoke(service, funcName, args...)
		}
		if auth && !superAdmin { 
			return res, errors.New("not authorized to " + method.String() + " " + table.Name + " datas") 
		}
		if col, ok := params[tool.RootColumnsParam]; ok { 
			if tablename == tool.ReservedParam { 
				return res, errors.New("can't load table as " + tool.ReservedParam) 
			}
			params[tool.RootColumnsParam]=strings.ToLower(col)
			service = table.TableColumn() 
		}
		return d.invoke(service, funcName, args...)
	}
	return res, errors.New("no service avaiblable")
}
func (d *MainService) invoke(service infrastructure.InfraServiceItf, funcName string, args... interface{}) (tool.Results, error) {
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

func SpecializedService(name string) tool.SpecializedService {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &tool.CustomService{}
}

var SERVICES = []tool.SpecializedService{&SchemaService{}, &SchemaFields{}, &TaskAssigneeService{}, &TaskVerifyerService{} }

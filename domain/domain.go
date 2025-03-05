package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	permissions "sqldb-ws/domain/permission"
	schserv "sqldb-ws/domain/schema"
	sm "sqldb-ws/domain/schema/models"
	domain "sqldb-ws/domain/service"
	"sqldb-ws/domain/utils"
	conn "sqldb-ws/infrastructure/connector"
	infrastructure "sqldb-ws/infrastructure/service"
	"strings"
	"time"
)

/*
		Domain is defined as the DDD patterns will suggest it.
		It's the specialized part of the API, it concive particular behavior on datas (in our cases, particular Root DB declares in entity)
		Main Service at a Domain level, it follows the DOMAIN ITF from schserv.
		Domain interact at a "Model" level with generic and abstract infra services.
		Mai	"fmt"

	  service give the main process to interact with Infra.
*/
var EXCEPTION_FUNC = []string{"Count"}

type SpecializedDomain struct {
	utils.AbstractDomain
	isGenericService bool
	PermsService     *permissions.PermDomainService
	Service          infrastructure.InfraServiceItf
	Db               *conn.Database
}

// generate a new domain controller.
func Domain(superAdmin bool, user string, isGenericService bool, permsService *permissions.PermDomainService) *SpecializedDomain {
	if permsService == nil {
		permsService = permissions.NewPermDomainService(conn.Open(nil), user, superAdmin, false)
	}
	return &SpecializedDomain{
		isGenericService: isGenericService, // generic specialized service is CustomService
		AbstractDomain: utils.AbstractDomain{
			SuperAdmin: superAdmin, // carry the security level of the "User" or an inner action
			User:       user,       // current user...
		},
		PermsService: permsService, // carry the permissions service
	}
}

func (d *SpecializedDomain) VerifyAuth(tableName string, colName string, level string, method utils.Method, args ...string) bool {
	if len(args) > 0 {
		return d.PermsService.LocalPermsCheck(tableName, colName, level, method, args[0])
	} else {
		return d.PermsService.PermsCheck(tableName, colName, level, method)
	}
}

func (s *SpecializedDomain) HandleRecordAttributes(record utils.Record) {
	s.isGenericService = record["is_custom"] != nil && record["is_custom"].(bool)
	s.Empty = record["is_empty"] != nil && record["is_empty"].(bool)
	s.LowerRes = record["is_list"] != nil && record["is_list"].(bool)
	s.Own = record["own_view"] != nil && record["own_view"].(bool)
}
func (d *SpecializedDomain) IsOwn(checkPerm bool, force bool, method utils.Method) bool {
	if checkPerm {
		return d.PermsService.IsOwnPermission(d.TableName, force, method) && d.Own
	}
	return d.Own
}
func (d *SpecializedDomain) GetDb() *conn.Database { return d.Db }

func (d *SpecializedDomain) CreateSuperCall(params utils.Params, record utils.Record, args ...interface{}) (utils.Results, error) {
	return d.SuperCall(params, utils.Record{}, utils.CREATE, false, args...) // how to...
}

func (d *SpecializedDomain) UpdateSuperCall(params utils.Params, record utils.Record, args ...interface{}) (utils.Results, error) {
	return d.SuperCall(params, utils.Record{}, utils.UPDATE, false, args...) // how to...
}

func (d *SpecializedDomain) DeleteSuperCall(params utils.Params, args ...interface{}) (utils.Results, error) {
	return d.SuperCall(params, utils.Record{}, utils.DELETE, false, args...) // how to...
}

// Infra func caller with admin view && superadmin right (not a structured view made around data for view reason)
func (d *SpecializedDomain) SuperCall(params utils.Params, record utils.Record, method utils.Method, isOwn bool, args ...interface{}) (utils.Results, error) {
	params[utils.RootRawView] = "enable"
	d2 := Domain(true, d.User, d.isGenericService, d.PermsService)
	if isOwn {
		d2.Own = d.IsOwn(false, false, method)
	}
	d2.ExternalSuperAdmin = d.ExternalSuperAdmin
	return d2.call(params, record, method, args...)
}

// Infra func caller with current option view and user rights.
func (d *SpecializedDomain) Call(params utils.Params, record utils.Record, method utils.Method, args ...interface{}) (utils.Results, error) {
	return d.call(params, record, method, args...)
}

func (d *SpecializedDomain) onBooleanValue(key string, sup func(bool)) {
	if t, ok := d.Params[key]; ok && t == "enable" {
		sup(ok)
	}
}

// Main process to call an Infra function
func (d *SpecializedDomain) call(params utils.Params, record utils.Record, method utils.Method, args ...interface{}) (utils.Results, error) {
	d.Method = method
	d.Params = params
	d.onBooleanValue(utils.RootSuperCall, func(b bool) { d.Super = b })
	d.onBooleanValue(utils.RootShallow, func(b bool) { d.Shallowed = b })
	if tablename, ok := params[utils.RootTableParam]; ok { // retrieve tableName in query (not optionnal)
		d.TableName = schserv.GetTablename(tablename)
		d.onBooleanValue(utils.RootRawView, func(b bool) {
			if !b {
				d.ClearDeprecatedDatas(tablename)
			}
		})
		var specializedService utils.SpecializedServiceITF = &domain.CustomService{}
		if !d.isGenericService {
			specializedService = domain.SpecializedService(tablename)
		}
		specializedService.SetDomain(d)
		d.Db = conn.Open(d.Db)
		defer d.Db.Close()
		if d.Method.IsMath() {
			d.Method = utils.SELECT
		}
		if !d.SuperAdmin && !d.PermsService.PermsCheck(tablename, "", "", d.Method) && !d.AutoLoad {
			return utils.Results{}, errors.New("not authorized to " + method.String() + " " + tablename + " data")
		}
		// load the highest entity avaiable Table level.
		d.Service = infrastructure.NewTableService(d.Db, d.SuperAdmin, d.User, strings.ToLower(tablename), record)
		delete(d.Params, utils.RootTableParam)
		tablename = strings.ToLower(tablename)
		if rowName, ok := params[utils.RootRowsParam]; ok { // rows override columns
			return d.GetRowResults(rowName, specializedService, args...)
		}
		if !d.SuperAdmin || method == utils.DELETE {
			return utils.Results{}, errors.New(
				"not authorized to " + method.String() + " " + d.Service.GetName() + " data")
		}
		if col, ok := params[utils.RootColumnsParam]; ok && tablename != utils.ReservedParam {
			d.Service = d.Service.(*infrastructure.TableService).NewTableColumnService(specializedService, strings.ToLower(col))
		} else if tablename == utils.ReservedParam {
			return utils.Results{}, errors.New("can't load table as " + utils.ReservedParam)
		}
		return d.invoke(method, args...)
	}
	return utils.Results{}, errors.New("no service available")
}

func (d *SpecializedDomain) GetRowResults(rowName string, specializedService utils.SpecializedServiceITF, args ...interface{}) (utils.Results, error) {
	if id, ok := d.Params[utils.SpecialIDParam]; ok {
		d.Params[utils.SpecialSubIDParam] = id
	}
	d.Params.Add(utils.SpecialIDParam, strings.ToLower(rowName), func(_ string) bool {
		return strings.ToLower(rowName) != utils.ReservedParam
	})
	delete(d.Params, utils.RootRowsParam)
	if d.Params[utils.SpecialIDParam] == "" || d.Params[utils.SpecialIDParam] == utils.ReservedParam || d.Params[utils.SpecialIDParam] == "<nil>" {
		delete(d.Params, utils.SpecialIDParam)
	} else if d.Service.(*infrastructure.TableService).Record != nil {
		d.Service.(*infrastructure.TableService).Record[utils.SpecialIDParam] = d.Params[utils.SpecialIDParam]
	}
	d.Service = d.Service.(*infrastructure.TableService).NewTableRowService(specializedService)
	res, err := d.invoke(d.Method, args...)
	if err == nil && d.Params[utils.RootRawView] != "enable" && !d.IsSuperCall() && !slices.Contains(EXCEPTION_FUNC, d.Method.Calling()) {
		return specializedService.TransformToGenericView(res, d.TableName, d.Params.GetAsArgs(utils.RootDestTableIDParam)...), nil
	}
	return res, err
}

func (d *SpecializedDomain) invoke(method utils.Method, args ...interface{}) (utils.Results, error) {
	res := utils.Results{}
	if d.Service == nil {
		return res, errors.New("no service available")
	}
	fmt.Println("invoke", method)
	clazz := reflect.ValueOf(d.Service).MethodByName(method.Calling())
	if !clazz.IsValid() || clazz.IsZero() {
		return res, errors.New("not implemented <" + method.Calling() + ">")
	}
	vals := []reflect.Value{}
	if method.IsMath() {
		vals = append(vals, reflect.ValueOf(method.String()))
	}
	for _, arg := range args {
		vals = append(vals, reflect.ValueOf(arg))
	}
	values := clazz.Call(vals)
	if len(values) > 0 {
		data, _ := json.Marshal(values[0].Interface())
		json.Unmarshal(data, &res)
	} else if len(values) > 1 {
		return res, values[1].Interface().(error)
	}
	return res, nil
}

func (d *SpecializedDomain) ClearDeprecatedDatas(tableName string) {
	if schema, err := schserv.GetSchema(tableName); err == nil && schema.HasField(sm.STARTKEY) && schema.HasField(sm.ENDKEY) {
		currentTime := time.Now()
		sqlFilter := "'" + currentTime.Format("2000-01-01") + "' < start_date OR "
		sqlFilter += "'" + currentTime.Format("2000-01-01") + "' > end_date"
		p := utils.AllParams(tableName).RootRaw()
		d.Call(p, utils.Record{}, utils.DELETE, sqlFilter)
	}
}

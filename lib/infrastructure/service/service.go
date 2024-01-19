package service

import (
	"os"
	"strings"
	"encoding/json"
	"html/template"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
	entities "sqldb-ws/lib/infrastructure/entities"
	
)

type InfraServiceItf interface {
	Verify(string)              (string, bool)
	Save() 			        	(error)
	Get()                   	(tool.Results, error)
	CreateOrUpdate()        	(tool.Results, error)
	Delete()                	(tool.Results, error)
	Import(string)          	(tool.Results, error)
	Template()               	(interface{}, error) 
	GenerateFromTemplate(string) error
	Close()
}

type InfraService struct {  
	Name                string       				`json:"name"`
	User                string       				`json:"-"`
	Params          	tool.Params       			`json:"-"`
	Record          	tool.Record       			`json:"-"`
	Results         	tool.Results      			`json:"-"`
	Method  	    	tool.Method     			`json:"-"`
	SuperAdmin 	    	bool		 				`json:"-"`
	PermService         *PermissionInfo             `json:"-"`
	db                  *conn.Db
	InfraServiceItf
}

func (service *InfraService) Close() { service.db.Conn.Close() }

func (service *InfraService) SpecializedFill(params tool.Params, record tool.Record, method tool.Method) {
	service.Record = record
	service.Method = method
	service.Params = params
}

func (service *InfraService) Fill(name string, admin bool, user string, params tool.Params, record tool.Record, method tool.Method) {
	service.Name = name
	service.Record = record
	service.Method = method
	service.Params = params
	service.User = user
	service.SuperAdmin = admin
}

func (service *InfraService) Save() error {
	res, err := service.Get()
	if err != nil { return err  }
	file, err := json.MarshalIndent(res, "", " ")
	if err != nil { return err }
	return os.WriteFile(service.Name, file, 0644)
}

func (service *InfraService) GenerateFromTemplate(templateName string) error {
	data, err := service.Template()
	t, err := template.ParseFiles(templateName)
	if err != nil { return err  }
	f, err := os.Create(service.Name)
	if err != nil { return err  }
	if t.Execute(f, data) != nil { return err  }
	return nil
}

func (service *InfraService) restrition() {
	for key, element := range service.Params {
		col := &TableColumnInfo{ }
		col.db = service.db
		col.Name = service.Name
		typ, ok := col.Verify(key)
		if ok { 
			if strings.Contains(element, ",") { 
				els := ""
				for _, el := range strings.Split(element, ",") { els += conn.FormatForSQL(typ, el) + "," }
				service.db.SQLRestriction += key + " IN (" + conn.RemoveLastChar(els) + ") " 
			} else { service.db.SQLRestriction += key + "=" + conn.FormatForSQL(typ, element) + " " }
		}
	}
}

func EmptyTable(name string) *TableInfo {
	database := conn.Open()
	if database == nil { return nil }
    table := &TableInfo{ } 
	table.db = database 
	table.Name = name 
	return table 
}

func Table(admin bool, user string, name string, params tool.Params, record tool.Record, method tool.Method) *TableInfo {
	database := conn.Open()
	if database == nil { return nil }
	table := &TableInfo{ } 
	table.db = database 
    table.Fill(name, admin, user, params, record, method)
	table.PermService = Permission(admin, user, tool.Params{}, tool.Record{}, method)
	return table
}

func Permission(admin bool, user string, params tool.Params, record tool.Record, method tool.Method) *PermissionInfo {
	database := conn.Open()
	if database == nil { return nil }
	perms :=  &PermissionInfo { }
	perms.db = database
	perms.Perms = map[string]tool.Record{}
	perms.WarningUpdateField = []string{}
	perms.Fill(entities.DBPermission.Name, admin, user, params, record, method)
	// HEAVY SQL PERMISSIONS
	paramsNew := tool.Params{ tool.RootSQLFilterParam : tool.ReservedParam, }
	paramsNew[tool.RootSQLFilterParam] += entities.DBPermission.Name + ".id IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT " + entities.DBPermission.Name + "_id FROM " 
	paramsNew[tool.RootSQLFilterParam] += entities.DBRolePermission.Name + " WHERE " + entities.DBRole.Name + "_id IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT " + entities.DBRole.Name + "_id FROM " 
	paramsNew[tool.RootSQLFilterParam] += entities.DBRoleAttribution.Name + " WHERE " + entities.DBUser.Name + "_id IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE " 
	paramsNew[tool.RootSQLFilterParam] += entities.DBUser.Name + ".login = " + conn.Quote(perms.User) + ") OR " + entities.DBEntity.Name + "_id IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT " + entities.DBEntity.Name + "_id FROM "
	paramsNew[tool.RootSQLFilterParam] += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
	paramsNew[tool.RootSQLFilterParam] += entities.DBUser.Name + ".login = " + conn.Quote(perms.User) + "))))"
    perms.Row = &TableRowInfo{ } 
	perms.Row.Table = EmptyTable(entities.DBPermission.Name)
	perms.Row.SpecializedService = &tool.CustomService{}
	perms.Row.db = database
	perms.Row.Fill(entities.DBPermission.Name, admin, user, paramsNew, tool.Record{}, tool.SELECT,)
	perms.Row.PermService=nil
	if res, err := perms.Row.Get(); res != nil && err == nil { perms.generatePerms(res) }
	return perms
}

func Load() {
	database := conn.Open()
	for _, table := range entities.ROOTTABLES {
		rec := tool.Record{}
		data, _:= json.Marshal(table)
		json.Unmarshal(data, &rec)
		service := &TableInfo{ }
		service.db = database
		service.Name = table.Name
		service.SpecializedFill(tool.Params{}, rec, tool.CREATE)
		if _,ok := service.Verify(table.Name); !ok { service.Create() }
	}
}

func DBError(res tool.Results, err error) (tool.Results, error) {
	log.Error().Msg(err.Error())
	return res, err
}
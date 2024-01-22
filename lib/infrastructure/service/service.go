package service

import (
	"os"
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
	Link()        				(tool.Results, error)
	UnLink()                	(tool.Results, error)
	Import(string)          	(tool.Results, error)
	Template()               	(interface{}, error) 
	GenerateFromTemplate(string) error
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

func EmptyTable(database *conn.Db, name string) *TableInfo {
    table := &TableInfo{ } 
	table.db = database 
	table.Name = name 
	return table 
}
func TableNoPerm(database *conn.Db, admin bool, user string, name string, params tool.Params, record tool.Record, method tool.Method) *TableInfo {
	table := &TableInfo{ } 
	table.db = database 
    table.Fill(name, admin, user, params, record, method)
	return table
}

func Table(database *conn.Db, admin bool, user string, name string, params tool.Params, record tool.Record, method tool.Method) *TableInfo {
	table := TableNoPerm(database, admin, user, name, params, record, method)
	table.PermService = Permission(database, admin, user, tool.Params{}, tool.Record{}, method)
	return table
}

func Permission(database *conn.Db, admin bool, user string, params tool.Params, record tool.Record, method tool.Method) *PermissionInfo {
	perms :=  &PermissionInfo { }
	perms.db = database
	perms.Perms = map[string]tool.Record{}
	perms.WarningUpdateField = []string{}
	perms.Fill(entities.DBPermission.Name, admin, user, params, record, method)
	// HEAVY SQL PERMISSIONS
	paramsNew := tool.Params{ }
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
	perms.Row.Table = EmptyTable(database, entities.DBPermission.Name)
	perms.Row.SpecializedService = &tool.CustomService{}
	perms.Row.db = database
	perms.Row.Fill(entities.DBPermission.Name, admin, user, paramsNew, tool.Record{}, tool.SELECT,)
	perms.Row.PermService=nil
	if res, err := perms.Row.Get(); res != nil && err == nil { perms.generatePerms(res) }
	return perms
}

func Load() {
	database := conn.Open()
	defer database.Conn.Close()
	for _, table := range entities.ROOTTABLES {
		rec := tool.Record{}
		data, _:= json.Marshal(table)
		json.Unmarshal(data, &rec)
		service := Table(database, true, "", table.Name, tool.Params{}, rec, tool.CREATE)
		if _,ok := service.Verify(table.Name); !ok { service.Create() }
	}
}

func DBError(res tool.Results, err error) (tool.Results, error) {
	log.Error().Msg(err.Error())
	return res, err
}
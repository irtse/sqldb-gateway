package service

import (
	"os"
	"encoding/json"
	"html/template"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
	entities "sqldb-ws/lib/entities"
	
)

type InfraService struct {  
	Name                string       				`json:"name"`
	User                string       				`json:"-"`
	Params          	tool.Params       			`json:"-"`
	Record          	tool.Record       			`json:"-"`
	Results         	tool.Results      			`json:"-"`
	Method  	    	tool.Method     			`json:"-"`
	SuperAdmin 	    	bool		 				`json:"-"`
	PermService         *PermissionInfo             `json:"-"`
	NoLog				bool						`json:"-"`
	db                  *conn.Db
	tool.InfraServiceItf
}

func (service *InfraService) SetAuth(auth bool) {
	if !auth { service.PermService= nil }
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

func (service *InfraService) DBError(res tool.Results, err error) (tool.Results, error) {
	if !service.NoLog && os.Getenv("log") == "enable" { log.Error().Msg(err.Error()) }
	return res, err
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
	for _, restricted := range entities.DBRESTRICTED {
		if table.Name == restricted.Name { table.PermService = nil; break }
	}
	return table
}

func Permission(database *conn.Db, admin bool, user string, params tool.Params, record tool.Record, method tool.Method) *PermissionInfo {
	perms :=  &PermissionInfo { }
	perms.db = database
	perms.Perms = map[string]tool.Record{}
	perms.WarningUpdateField = []string{}
	perms.Fill(entities.DBPermission.Name, admin, user, params, record, method)
    perms.Row = &TableRowInfo{ } 
	perms.Row.Table = EmptyTable(database, entities.DBPermission.Name)
	perms.Row.SpecializedService = nil
	perms.Row.db = database
	perms.Row.Fill(entities.DBPermission.Name, admin, user, tool.Params{}, tool.Record{}, tool.SELECT,)
	perms.Row.PermService=nil
	return perms
}
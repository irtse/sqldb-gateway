package service

import (
	"os"
	"html/template"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
)
/*
	Infrastructure is meant as DDD pattern, as a generic accessor to database and distant services. 
	Main Procedure of services at Infrastructure level.
*/
type InfraService struct {  
	Name                string       				`json:"name"`
	User                string       				`json:"-"`
	Params          	map[string]string       	`json:"-"`
	Record          	tool.Record       			`json:"-"`
	Results         	tool.Results      			`json:"-"`
	Method  	    	tool.Method     			`json:"-"`
	SuperAdmin 	    	bool		 				`json:"-"`
	NoLog				bool						`json:"-"`
	db                  *conn.Db
	tool.InfraServiceItf
}
// Service Builder for Specialized purpose
func (service *InfraService) SpecializedFill(params tool.Params, record tool.Record, method tool.Method) {
	service.Record = record
	service.Method = method
	service.Params = params
}
// Main Service Builder 
func (service *InfraService) Fill(name string, admin bool, user string, params tool.Params, record tool.Record, method tool.Method) {
	service.Name = name
	service.Record = record
	service.Method = method
	service.Params = params
	service.User = user
	service.SuperAdmin = admin
}
// Common Service action of generation by template (TO USE)
func (service *InfraService) GenerateFromTemplate(templateName string) error {
	data, err := service.Template()
	t, err := template.ParseFiles(templateName)
	if err != nil { return err  }
	f, err := os.Create(service.Name)
	if err != nil { return err  }
	if t.Execute(f, data) != nil { return err  }
	return nil
}
// Common service error 
func (service *InfraService) DBError(res tool.Results, err error) (tool.Results, error) {
	if !service.NoLog && os.Getenv("log") == "enable" { log.Error().Msg(err.Error()) }
	return res, err
}
// Generate an Empty TableInfo Service (no perm, no specialization)
func EmptyTable(database *conn.Db, name string) *TableInfo {
    table := &TableInfo{ } 
	table.db = database 
	table.Name = name 
	return table 
}
// Generate an Empty TableInfo Service
func Table(database *conn.Db, admin bool, user string, name string, params tool.Params, record tool.Record, method tool.Method) *TableInfo {
	table := &TableInfo{ } 
	table.db = database 
    table.Fill(name, admin, user, params, record, method)
	return table
}
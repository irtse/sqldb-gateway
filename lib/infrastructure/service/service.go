package service

import (
	"os"
	"html/template"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
)
/*
	Infrastructure is meant as DDD pattern, as a generic accessor to database and distant services. 
	Main Procedure of services at Infrastructure level.
*/
type InfraSpecializedServiceItf interface {
	ConfigureFilter(tableName string, innerestr... string) (string, string, string, string)
	WriteRowAutomation(record map[string]interface{}, tableName string)
	UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) 
	DeleteRowAutomation(results []map[string]interface{}, tableName string)
	VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool)
}

type InfraSpecializedService struct { }
func (s *InfraSpecializedService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { }
func (s *InfraSpecializedService) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {}
func (s *InfraSpecializedService) WriteRowAutomation(record map[string]interface{}, tableName string) { }

type InfraServiceItf interface {
	Verify(string)              					(string, bool)
	Count(restriction... string)  					([]map[string]interface{}, error)
	Get(restriction... string)  					([]map[string]interface{}, error)
	CreateOrUpdate(restriction... string)        	([]map[string]interface{}, error)
	Delete(restriction... string)                	([]map[string]interface{}, error)
	Template(restriction... string)               	(interface{}, error) 
	GenerateFromTemplate(string) error
}
type InfraService struct {  
	Name                string       				`json:"name"`
	User                string       				`json:"-"`
	Record          	map[string]interface{}      `json:"-"`
	Results         	[]map[string]interface{}    `json:"-"`
	SuperAdmin 	    	bool		 				`json:"-"`
	NoLog				bool						`json:"-"`
	SpecializedService  InfraSpecializedServiceItf 	`json:"-"`
	db                  *conn.Db
	InfraServiceItf
}
// Service Builder for Specialized purpose
func (service *InfraService) SpecializedFill(record map[string]interface{}) {
	service.Record = record
}
// Main Service Builder 
func (service *InfraService) Fill(name string, admin bool, user string, record map[string]interface{}) {
	service.Name = name
	service.Record = record
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
func (service *InfraService) DBError(res []map[string]interface{}, err error) ([]map[string]interface{}, error) {
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
func Table(database *conn.Db, admin bool, user string, name string, record map[string]interface{}) *TableInfo {
	table := &TableInfo{ } 
	table.db = database 
    table.Fill(name, admin, user, record)
	return table
}
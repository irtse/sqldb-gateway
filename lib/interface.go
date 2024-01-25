package lib

import (
	"fmt"
	"strings"
	"sqldb-ws/lib/infrastructure/entities"
)

type InfraServiceItf interface {
	SetPostTreatment(bool)
	Verify(string)              (string, bool)
	Save() 			        	(error)
	Get()                   	(Results, error)
	CreateOrUpdate()        	(Results, error)
	Delete()                	(Results, error)
	Link()        				(Results, error)
	UnLink()                	(Results, error)
	Import(string)          	(Results, error)
	Template()               	(interface{}, error) 
	GenerateFromTemplate(string) error
}

type DomainITF interface {
	SuperCall(params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	Call(params Params, rec Record, m Method, auth bool, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	GetUser() string
	IsSuperAdmin() bool
	GetPermission() InfraServiceItf
}
type SpecializedServiceInfo interface { GetName() string }
type SpecializedService interface {
	Entity() SpecializedServiceInfo
	SetDomain(d DomainITF)
	WriteRowAutomation(record Record)
	VerifyRowAutomation(record Record, create bool) (Record, bool)
	DeleteRowAutomation(results Results)
	UpdateRowAutomation(results Results, record Record) 
	PostTreatment(results Results) Results
	ConfigureFilter(tableName string, params Params) (string, string)
}

type AbstractSpecializedService struct { Domain DomainITF }
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }

func GeneratePermissionCommand(tableName string, throughTableName string, userName string) string {
	cmd := "id IN (SELECT " + entities.RootID(tableName) + " FROM " + throughTableName + " " 
	cmd += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	cmd += "SELECT id FROM " + entities.DBUser.Name + "WHERE login=" + userName +") OR "
	cmd += entities.RootID(entities.DBEntity.Name) + " IN ("
	cmd += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	cmd += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	cmd += "SELECT id FROM " + entities.DBUser.Name + "WHERE login=" + userName + "))) "
	return cmd
}

func ViewDefinition(domain DomainITF, tableName string, params Params) (string, string) {
	SQLview := ""
	SQLrestriction := ""
	p := Params{ RootTableParam : entities.DBSchema.Name, 
		         RootRowsParam : ReservedParam,
		         entities.NAMEATTR : tableName, 
	}
	schemas, err := domain.SuperCall( p, Record{}, SELECT, "Get")
	if err != nil || len(schemas) == 0 { return SQLrestriction, SQLview }
	p = Params{ RootTableParam : entities.DBSchemaField.Name, 
		        RootRowsParam : ReservedParam,
		        entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", schemas[0][SpecialIDParam]), }
	fields, err := domain.SuperCall( p, Record{}, SELECT, "Get")
	if err == nil {
		for _, field := range fields {
			if  hide, ok := field["hidden"]; ok && !hide.(bool) {
				if columns, ok3:= params["columns"]; ok3 && columns != "" {
					if strings.Contains(columns, field["name"].(string)) { SQLview += field["name"].(string) + "," } 
				} else { SQLview += field["name"].(string) + "," }
			}
		}
	}
	p = Params{ RootTableParam : entities.DBView.Name, 
		        RootRowsParam : ReservedParam,
		        entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", schemas[0][SpecialIDParam]), }
	views, err := domain.SuperCall( p, Record{}, SELECT, "Get")
	if err == nil {
		for _, view := range views {
			if through, ok := view["through_perms"]; ok {
				p = Params{ RootTableParam : entities.DBSchema.Name, 
					         RootRowsParam : fmt.Sprintf("%v", through), }          
				throughs, err := domain.SuperCall( p, Record{}, SELECT, "Get")
				if err != nil || len(throughs) == 0 { continue }
				if len(SQLrestriction) > 0 {
					SQLrestriction += " AND " + GeneratePermissionCommand(schemas[0][entities.NAMEATTR].(string),
				                                                          throughs[0][entities.NAMEATTR].(string),
															              domain.GetUser(),
														                 )
				} else {
					SQLrestriction +=  GeneratePermissionCommand(schemas[0][entities.NAMEATTR].(string),
				                                                 throughs[0][entities.NAMEATTR].(string),
															     domain.GetUser(),
														        )
				}
			}
		}
	}
	if len(SQLview) > 0 {SQLview = SQLview[:len(SQLview) - 1] }
	return SQLrestriction, SQLview
}
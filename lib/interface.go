package lib

import (
	"fmt"
	"strings"
	"sqldb-ws/lib/infrastructure/entities"
)

type DomainITF interface {
	SafeCall(admin bool, user string, params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	UnSafeCall(user string, params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	GetUser() string
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

func (s *AbstractSpecializedService) ConfigureFilter(tableName string, params Params) (string, string) {
	p := Params{ RootTableParam : entities.DBSchema.Name, 
		              RootRowsParam : ReservedParam,
			          entities.NAMEATTR : tableName, 
	}
	schemas, err := s.Domain.SafeCall(true, "", p, Record{}, SELECT, "Get")
	if err != nil || len(schemas) == 0 { return "", "" }
	p = Params{ RootTableParam : entities.DBView.Name, 
				RootRowsParam : ReservedParam,
				entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%d", schemas[0][SpecialIDParam].(int64)), }
	views, err := s.Domain.SafeCall(true, "", p, Record{}, SELECT, "Get")
	if err != nil || len(views) == 0 { return "", "" }
	SQLrestriction := ""
	SQLview := ""
	view := views[0]
	if filter, valid := view["sqlfilter"]; valid && filter != nil && filter != "" {
		SQLrestriction += filter.(string) + " "
	}												
	p = Params{ RootTableParam : entities.DBSchemaField.Name, 
		        RootRowsParam : ReservedParam,
		        entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%d", schemas[0][SpecialIDParam].(int64)), }
	fields, err := s.Domain.SafeCall(true, "", p, Record{}, SELECT, "Get")
	if err != nil || len(views) == 0 { return "", "" }
	for _, field := range fields {
		if  hide, ok := field["hidden"]; ok && !hide.(bool) {
			if columns, ok3:= params["columns"]; ok3 && columns != "" {
				if strings.Contains(columns, field["name"].(string)) { SQLview += field["name"].(string) + "," } 
			} else { SQLview += field["name"].(string) + "," }
		}
	}
	return SQLrestriction[:len(SQLrestriction) - 1], SQLview[:len(SQLview) - 1] 
}	
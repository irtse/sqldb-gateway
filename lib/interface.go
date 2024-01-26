package lib

import (
	"fmt"
	"errors"
	"strings"
	"encoding/json"
	"sqldb-ws/lib/infrastructure/entities"
)

type InfraServiceItf interface {
	SetAuth(bool)
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
	IsShallowed() bool 
	IsAdminView() bool
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
	PostTreatment(results Results, tableName string) Results
	ConfigureFilter(tableName string, params Params) (string, string)
}

type AbstractSpecializedService struct { Domain DomainITF }
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }

func PostTreat(domain DomainITF, results Results, tableName string) Results {
	res := Results{}
	var cols map[string]entities.TableColumnEntity
	if !domain.IsShallowed() {
		params := Params{ RootTableParam : tableName, }
			schemas, err := domain.SuperCall( params, Record{}, SELECT, "Get")
			if err != nil || len(schemas) == 0 { return Results{} }
			if _, ok := schemas[0]["columns"]; !ok { return Results{} }
			cols = schemas[0]["columns"].(map[string]entities.TableColumnEntity)
	}
	for _, record := range results { 
		rec := PostTreatRecord(domain, record, tableName, cols) 
		if rec != nil { res = append(res, rec) }
	}
	return res
}

func PostTreatRecord(domain DomainITF, record Record, tableName string, cols map[string]entities.TableColumnEntity, additonnalRestriction ...string) Record {
	newRec := Record{}
	if domain.IsShallowed() {
		if _, ok := record[entities.NAMEATTR]; ok {
			return Record{ entities.NAMEATTR : record[entities.NAMEATTR] }
		} else { return record }
	} else {
		if domain.IsAdminView() { return record } // if admin view avoid.
		contents := map[string]interface{}{}
		vals := map[string]interface{}{}
		for _, field := range cols {
			var fieldInfo entities.TableColumnEntity
			b, _:= json.Marshal(field)
			json.Unmarshal(b, &fieldInfo)
			vals[fieldInfo.Name]=record[fieldInfo.Name]
			if _, ok := record[fieldInfo.Name]; ok && strings.Contains(fieldInfo.Name, "_" + SpecialIDParam){ 
				tableName := fieldInfo.Name[:(len(fieldInfo.Name) - len(SpecialIDParam) - 1)]
				params := Params{ RootTableParam : tableName, 
						            RootRowsParam : fmt.Sprintf("%v", record[fieldInfo.Name]), }
				fmt.Printf("ADD %v \n", additonnalRestriction)
				params[RootSQLFilterParam]=""
				for _, add := range additonnalRestriction { 
					if add != "" { params[RootSQLFilterParam] += add + " AND " }
				}
				if len(params[RootSQLFilterParam]) >= 4 { 
					params[RootSQLFilterParam] = params[RootSQLFilterParam][:len(params[RootSQLFilterParam]) - 4]
				}
				datas, err := domain.SuperCall( params, Record{}, SELECT, "Get")
				if err != nil || len(datas) == 0 { continue }
				contents[tableName] = PostTreat(domain, datas, tableName)
			}
		}
		newRec["values"]=vals
		newRec["schema"]=cols
		newRec["name"]=tableName
		newRec["contents"]=contents
	}
	return newRec 
}

func Schema(domain DomainITF, record Record) (Results, error) {
	if schemaID, ok := record[entities.RootID(entities.DBSchema.Name)]; ok {
		params := Params{ RootTableParam : entities.DBSchema.Name, 
			              RootRowsParam : fmt.Sprintf("%v", schemaID), 
		                }
		schemas, err := domain.SuperCall( params, Record{}, SELECT, "Get")
		if err != nil || len(schemas) == 0 { return nil, err }
		if _, ok := domain.GetPermission().Verify(schemas[0][entities.NAMEATTR].(string)); !ok { 
			return nil, errors.New("not authorized ") 
		}
		return schemas, nil
	}
	return nil, errors.New("no schemaID refered...")
}

func ViewDefinition(domain DomainITF, tableName string, params Params) (string, string) {
	SQLview := ""
	SQLrestriction := ""
	auth := true
	if admin, ok := params[RootAdminView]; ok && admin == "enable" && domain.IsSuperAdmin() { return SQLrestriction, SQLview }
	for _, exception := range entities.PERMISSIONEXCEPTION {
		if tableName == exception.Name { auth = false; break }
	}
	if auth { SQLrestriction, SQLview = byFields(domain, tableName) }
	if filter, ok := params[RootSQLFilterParam]; ok {
		if len(SQLrestriction) > 0 { SQLrestriction += " AND " + filter 
	    } else { SQLrestriction = filter  }
	}
	return SQLrestriction, SQLview
}

func byFields(domain DomainITF, tableName string) (string, string) {
	SQLview := ""
	for _, restricted := range entities.DBRESTRICTED {
		if restricted.Name == tableName { return "id=-1", "" }
	}
	p := Params{ RootTableParam : entities.DBSchema.Name,
	                  RootRowsParam : ReservedParam,
					  entities.NAMEATTR : tableName }
	schemas, err := domain.SuperCall( p, Record{}, SELECT, "Get")
	if err != nil || len(schemas) == 0 { return "id=-1", "" }
	domain.SuperCall( p, Record{}, SELECT, "Get")
	p = Params{ RootTableParam : entities.DBSchemaField.Name,
	                  RootRowsParam : ReservedParam,
					  entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", schemas[0][SpecialIDParam]) }
	fields, err := domain.SuperCall( p, Record{}, SELECT, "Get")
	if err != nil || len(fields) == 0 { return "id=-1", "" }
	for _, field := range fields {
		if name, okName := field["name"]; !okName {
			n := fmt.Sprintf("%v", name)
			if hide, ok := field["hidden"]; (!ok || !hide.(bool)) && n != "id" {
				SQLview += n + ","
			}
		}
	}
	return "", SQLview
}
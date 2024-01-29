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
	GetDB() DbITF
} 

type DbITF interface {
	GetSQLView()        string 
	GetSQLOrder()       string 
	GetSQLRestriction() string 	
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

func PostTreat(domain DomainITF, results Results, tableName string, shallow bool, additonnalRestriction ...string) Results {
	res := Results{}
	cols := map[string]entities.SchemaColumnEntity{}
	sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM " + entities.DBSchema.Name + " WHERE name='" + tableName + "')"
	params := Params{ RootTableParam : entities.DBSchemaField.Name, RootRowsParam: ReservedParam, RootSQLFilterParam: sqlFilter }
	schemas, err := domain.SuperCall( params, Record{}, SELECT, "Get")
	if err != nil || len(schemas) == 0 { return Results{} }
	if !domain.IsShallowed() {
		for _, r := range schemas {
			var scheme entities.SchemaColumnEntity
			b, _ := json.Marshal(r)
			json.Unmarshal(b, &scheme)
			cols[scheme.Name]=scheme
		}
	}
	for _, record := range results { 
		rec := PostTreatRecord(domain, record, fmt.Sprintf("%v", schemas[0][SpecialIDParam]), tableName, cols, shallow, additonnalRestriction...) 
		if rec != nil { res = append(res, rec) }
	}
	return res
}

func PostTreatRecord(domain DomainITF, record Record, tableID string, tableName string, cols map[string]entities.SchemaColumnEntity, shallow bool, additonnalRestriction ...string) Record {
	newRec := Record{}
	if domain.IsShallowed() {
		if _, ok := record[entities.NAMEATTR]; ok {
			return Record{ entities.NAMEATTR : record[entities.NAMEATTR] }
		} else { return record }
	} else {
		if domain.IsAdminView() { return record } // if admin view avoid.
		readonly := false 
		if r, ok := record["readonly"]; ok && r.(bool) { readonly = true }
		schemass := map[string]interface{}{}
		schemes := map[string]interface{}{}
		contents := map[string]interface{}{}
		vals := map[string]interface{}{}
		path := "/" + tableName + "?rows=all"
		newRec["name"]=tableName
		newRec["description"]=tableName + " datas"
		if domain.GetDB().GetSQLView() != "" {
			path += "&" + RootColumnsParam + "=" + strings.TrimSpace(domain.GetDB().GetSQLView())
		}
		if strings.TrimSpace(domain.GetDB().GetSQLRestriction()) != "" || len(additonnalRestriction) > 0 {
			base := "&" + RootSQLFilterParam + "="
			ext := ""
			if domain.GetDB().GetSQLRestriction() != "" {
				ext += strings.Replace(strings.TrimSpace(domain.GetDB().GetSQLRestriction()), " ", "+", -1) + "+"
			}
			for _, add := range additonnalRestriction { 
				if add != "" { ext += add + "+AND+" }
			}
			if len(ext) >= 4 && ext[:len(ext) - 4] == "AND+" { ext = ext[:len(ext) - 4] }
			if ext != "" { path += base + ext }
		}
		if strings.TrimSpace(domain.GetDB().GetSQLOrder()) != "" {
			path += "&" + RootOrderParam + "=" + strings.Replace(strings.TrimSpace(domain.GetDB().GetSQLOrder()), " ", "+", -1)
		}
		if !shallow {
			for _, field := range cols {
				if readonly { field.Readonly = true }
				if domain.GetDB().GetSQLView() != "" && !strings.Contains(domain.GetDB().GetSQLView(), field.Name){ continue }
				if field.Link != 0 {
					schemas, err := Schema(domain, Record{entities.RootID(entities.DBSchema.Name) : field.Link})
					if err != nil || len(schemas) == 0 { continue }
					link_path := "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]) +  "?rows=all"
					if field.LinkRestriction != "" { 
						link_path += "&" + RootSQLFilterParam + "=" + strings.Replace(field.LinkRestriction, " ", "+", -1) }
					if field.LinkOrder != "" { 
						link_path += "&" + RootOrderParam + "=" + strings.Replace(field.LinkOrder, " ", "+", -1) 
					}
					if field.LinkDir != "" { 
						link_path += "&" + RootDirParam + "=" + strings.Replace(field.LinkDir, " ", "+", -1) 
					}
					if field.LinkColumns != "" { 
						link_path += "&" + RootColumnsParam + "=" + field.LinkColumns
					}
					field.LinkPath=link_path
				}
				schemes[field.Name]=field
				vals[field.Name]=record[field.Name]
				if _, ok := record[field.Name]; ok && strings.Contains(field.Name, "_" + SpecialIDParam) { 
					tableName := field.Name[:(len(field.Name) - len(SpecialIDParam) - 1)]
					params := Params{ RootTableParam : tableName, 
									  RootRowsParam : fmt.Sprintf("%v", record[field.Name]), }
					datas, err := domain.SuperCall( params, Record{}, SELECT, "Get")
					if err != nil || len(datas) == 0 { continue }
					tbname := datas[0][entities.NAMEATTR]
					ad := Results{}
					ress := PostTreat(domain, Results{Record{}}, tbname.(string), true, additonnalRestriction...)
					for _, recs := range ress {
						schemass[tbname.(string)]=recs["schema"]
						delete(recs, "schema")
						ad = append(ad, recs)
					}
					contents[tbname.(string)]=ad
				}
			}
		}
		// TODO IF GET IMPORT ACTION OF 
		newRec["path"]=path
		if !shallow {
			newRec["values"]=vals
			newRec["schema"]=schemes 
			if len(schemass) > 0 { newRec["schemas"]=schemass }
			if len(contents) > 0 { newRec["contents"]=contents }
		}
		//newRec["actions"]=actions
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
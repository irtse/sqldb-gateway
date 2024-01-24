package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type SchemaFields struct { tool.AbstractSpecializedService }

func (s *SchemaFields) Entity() tool.SpecializedServiceInfo {return entities.DBSchemaField }
func (s *SchemaFields) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) {
	schemas, err := Schema(s.Domain, record)
	newRecord := tool.Record{}
	if !create {
		for k, v := range record {
			if k == "name" { newRecord["label"] = v 
			} else if k != "type" { newRecord[k] = v }
		}
	}
	return newRecord, err == nil && schemas != nil && len(schemas) > 0
}
func (s *SchemaFields) WriteRowAutomation(record tool.Record) { 
	res, err := s.Domain.SuperCall(
		tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			         tool.RootRowsParam: fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) }, 
		tool.Record{}, 
		tool.SELECT, 
		"Get",
	)
	if err != nil { fmt.Printf("ERROR %s", err.Error()) }
	data := tool.Record{ 
		entities.NAMEATTR : record[entities.NAMEATTR],
		entities.TYPEATTR : record[entities.TYPEATTR],
	}
	if _, ok := record["default_value"]; ok { data["default_value"] = record["default_value"] }
	if _, ok := record["description"]; ok { data["comment"] = record["description"] }
	if len(res) > 0 {
		_, err := s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				         tool.RootColumnsParam: tool.ReservedParam }, 
			data, 
			tool.CREATE, 
			"CreateOrUpdate")
		if err != nil { fmt.Printf("error %s", err.Error()) }
	}
}
func (s *SchemaFields) UpdateRowAutomation(results tool.Results, record tool.Record) {
	for _, r := range results {
		res, err := s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
				    tool.RootRowsParam: fmt.Sprintf("%s", r[entities.RootID(entities.DBSchema.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get",
		)
		if err != nil || res == nil || len(res) == 0 { return }
		newRecord := tool.Record{}
		for k, v := range record {
			if k == "default_value" { newRecord[k] = v 
			} else if k == "description"{ newRecord["comment"] = v }
		}
		newRecord[entities.TYPEATTR] = r[entities.TYPEATTR]
		newRecord[entities.NAMEATTR] = r[entities.NAMEATTR]
		_, err = s.Domain.SuperCall(
			tool.Params{ 
				tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				tool.RootColumnsParam: r[entities.NAMEATTR].(string) }, 
			newRecord, 
			tool.UPDATE, 
			"CreateOrUpdate",
		)
	}
}
func (s *SchemaFields) DeleteRowAutomation(results tool.Results) { 
	for _, record := range results { 
		res, err := s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
				    tool.RootRowsParam: fmt.Sprintf("%d", record[entities.RootID(entities.DBSchema.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get",
		)
		if err != nil || res == nil || len(res) == 0 { continue }
	    _, err = s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				    tool.RootColumnsParam: record[entities.NAMEATTR].(string) }, 
			tool.Record{}, 
			tool.DELETE, 
			"Delete",
		)
		if err != nil { fmt.Printf("error %s", err.Error()) }
	}
}
func (s *SchemaFields) PostTreatment(results tool.Results) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		schemas, err := Schema(s.Domain, tool.Record{
			entities.RootID(entities.DBSchema.Name) : record[entities.RootID(entities.DBSchema.Name)].(int64)})
		if err != nil || len(schemas) == 0 { continue }
		res = append(res, record)
	}
	return res 
}

func (s *SchemaFields) ConfigureFilter(tableName string, params tool.Params) (string, string) {
	return "", ""
}	
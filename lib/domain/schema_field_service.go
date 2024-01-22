package domain

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type SchemaFields struct {
	Domain tool.DomainITF
}

func (s *SchemaFields) SetDomain(d tool.DomainITF) { s.Domain = d }
func (s *SchemaFields) Entity() tool.SpecializedServiceInfo {return entities.DBSchemaField }
func (s *SchemaFields) VerifyRowWorkflow(record tool.Record, create bool) (tool.Record, bool) {
	rows := "all"
	found := false
	if _, ok := record[entities.RootID(entities.DBSchema.Name)]; !ok {
		if _, ok2 := record[entities.TABLENAMEATTR]; ok2 { found=true
		} else { return record, false }
	} else { rows = fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) }
	params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
						   tool.RootRowsParam: rows, }
	if found { params[entities.NAMEATTR]=record[entities.TABLENAMEATTR].(string)  }
	res, err := s.Domain.SafeCall(
		true, 
		"",
		params, 
		tool.Record{}, 
		tool.SELECT, 
		"Get",
	)
	newRecord := tool.Record{}
	if !create {
		for k, v := range record {
			if k == "name" { newRecord["label"] = v 
			} else if k != "type" { newRecord[k] = v }
		}
	}
	return newRecord, err == nil && res != nil && len(res) > 0
}
func (s *SchemaFields) WriteRowWorkflow(record tool.Record) { 
	res, err := s.Domain.SafeCall(true, "",
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
		_, err := s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				         tool.RootColumnsParam: tool.ReservedParam }, 
			data, 
			tool.CREATE, 
			"CreateOrUpdate")
		if err != nil { fmt.Printf("error %s", err.Error()) }
	}
}
func (s *SchemaFields) UpdateRowWorkflow(results tool.Results, record tool.Record) {
	for _, r := range results {
		res, err := s.Domain.SafeCall(true, "",
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
		_, err = s.Domain.SafeCall(true, "",
			tool.Params{ 
				tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				tool.RootColumnsParam: r[entities.NAMEATTR].(string) }, 
			newRecord, 
			tool.UPDATE, 
			"CreateOrUpdate",
		)
	}
}
func (s *SchemaFields) DeleteRowWorkflow(results tool.Results) { 
	for _, record := range results { 
		res, err := s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
				    tool.RootRowsParam: fmt.Sprintf("%d", record[entities.RootID(entities.DBSchema.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get",
		)
		if err != nil || res == nil || len(res) == 0 { continue }
	    _, err = s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				    tool.RootColumnsParam: record[entities.NAMEATTR].(string) }, 
			tool.Record{}, 
			tool.DELETE, 
			"Delete",
		)
		if err != nil { fmt.Printf("error %s", err.Error()) }
	}
}
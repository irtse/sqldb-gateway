package schema_service

import (
	"fmt"
	"strings"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type SchemaFields struct { tool.AbstractSpecializedService }

func (s *SchemaFields) Entity() tool.SpecializedServiceInfo {return entities.DBSchemaField }
func (s *SchemaFields) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) {
	newRecord := tool.Record{}
	if !create {
		for k, v := range record {
			if k == "name" { newRecord["label"] = v 
			} else if k != "type" { newRecord[k] = v }
		}
	} else {
		for k, v := range record {
			if k != "type" { newRecord[k] = v 
			} else { 
				if strings.Contains(fmt.Sprintf("%v", v), "enum") {
					typ := fmt.Sprintf("%v", v)
					typ = strings.Replace(typ, " ", "", -1)
					typ = strings.Replace(typ, "'", "", -1)
					typ = strings.Replace(typ, "(", ":", -1)
					typ = strings.Replace(typ, ")", "", -1)
					found := false
					for _, verifiedType := range tool.DATATYPE {
						if strings.Contains(strings.ToUpper(typ), verifiedType) {
							found = true; break;
						}
					}
					if found { newRecord[k] = strings.ToLower(typ) }
				} else { newRecord[k] = v  }
			}
		}
		if label, ok := newRecord["label"]; !ok || label == "" {
			newRecord["label"] = strings.Replace(fmt.Sprintf("%v", newRecord["name"]), "_", " ", -1)
		}
		if nullable, ok := newRecord["nullable"]; !ok || nullable == nil {
			newRecord["nullable"] = true
		}
	}
	return newRecord, true, true
}
func (s *SchemaFields) WriteRowAutomation(record tool.Record, tableName string) { 
	res, err := s.Domain.SuperCall(
		tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			         tool.RootRowsParam: fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) }, 
		tool.Record{}, 
		tool.SELECT, 
		"Get",
	)
	if err != nil || len(res) == 0 { return }
	for role, mainPerms := range tool.MAIN_PERMS {
			read_levels := []string{entities.LEVELNORMAL}
			if level, ok := record["read_level"]; ok && level != "" {
				read_levels = append(read_levels, strings.Replace(fmt.Sprintf("%v", level), "'", "", -1))
			}
			for _, l := range read_levels {
				rec := tool.Record{ 
					entities.NAMEATTR : fmt.Sprintf("%v", res[0][entities.NAMEATTR]) + ":" + fmt.Sprintf("%v", record[entities.NAMEATTR]) + ":" + l + ":" + role, 
				}
				for perms, value := range mainPerms { rec[perms]=value }
				rec[tool.SELECT.String()]=l
				s.Domain.SuperCall(
					tool.Params{ tool.RootTableParam : entities.DBPermission.Name, tool.RootRowsParam : tool.ReservedParam }, 
					rec, tool.CREATE, "CreateOrUpdate",)
			}
	}
	s.Domain.SuperCall(tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				       tool.RootColumnsParam: tool.ReservedParam }, 
						record, 
						tool.CREATE, 
						"CreateOrUpdate")
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
		if err != nil || len(res) == 0 { return }
		for role, mainPerms := range tool.MAIN_PERMS {
			read_levels := []string{entities.LEVELNORMAL}
			if level, ok := record["read_level"]; ok && level != "" {
				read_levels = append(read_levels,strings.Replace( fmt.Sprintf("%v", level), "'", "", -1))
			}
			for _, l := range read_levels {
				rec := tool.Record{ 
					entities.NAMEATTR : fmt.Sprintf("%v", res[0][entities.NAMEATTR]) + ":" + fmt.Sprintf("%v", record[entities.NAMEATTR]) + ":" + l + ":" + role, 
				}
				for perms, value := range mainPerms { rec[perms]=value }
				rec[tool.SELECT.String()]=l
				s.Domain.SuperCall(
					tool.Params{ tool.RootTableParam : entities.DBPermission.Name, tool.RootRowsParam : tool.ReservedParam }, 
					rec, tool.CREATE, "CreateOrUpdate",)
			}
		}
		newRecord := tool.Record{}
		for k, v := range record {
			newRecord[k] = v 
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
func (s *SchemaFields) DeleteRowAutomation(results tool.Results, tableName string) { 
	for _, record := range results { 
		res, err := s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
				    tool.RootRowsParam: fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get",
		)
		if err != nil || res == nil || len(res) == 0 { continue }
	    s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : res[0][entities.NAMEATTR].(string), 
				    tool.RootColumnsParam: record[entities.NAMEATTR].(string) }, 
			tool.Record{}, 
			tool.DELETE, 
			"Delete",
		)
		
		s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBPermission.Name, 
				         tool.RootRowsParam : tool.ReservedParam, 
						 entities.NAMEATTR : "%" + tableName + ":" + fmt.Sprintf("%v", record[entities.NAMEATTR]) + "%" }, 
						 tool.Record{ },  tool.DELETE, "Delete", )
	}
}
func (s *SchemaFields) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	res := tool.Results{}
	for _, rec := range results {
		schemas, err := s.Domain.Schema(rec, true)
		if err != nil && len(schemas) == 0 { continue }
		if s.Domain.PermsCheck(fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]), "", "", tool.SELECT) {
			res = append(res, rec)
		}
	}
	return s.Domain.PostTreat( res, tableName) 
}
func (s *SchemaFields) ConfigureFilter(tableName string) (string, string) {
	return s.Domain.ViewDefinition(tableName)
}	
package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type SchemaService struct { tool.AbstractSpecializedService }

func (s *SchemaService) Entity() tool.SpecializedServiceInfo { return entities.DBSchema }
func (s *SchemaService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { return record, true, false }
func (s *SchemaService) DeleteRowAutomation(results tool.Results, tableName string) { 
	for _, record := range results { 
		s.Domain.SetIsCustom(true)
		s.Domain.SuperCall( 
		                	tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
			                                tool.RootRowsParam: fmt.Sprintf("%v", record["id"]), }, 
							tool.Record{ }, 
							tool.DELETE, 
							"Delete",
						)
		s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBPermission.Name, 
				         tool.RootRowsParam : tool.ReservedParam,
						entities.NAMEATTR : "%" + tableName + "%" }, 
						tool.Record{ },  tool.DELETE, "Delete",)
	}
}
func (s *SchemaService) UpdateRowAutomation(_ tool.Results, record tool.Record) {
	for role, mainPerms := range tool.MAIN_PERMS {
		rec := tool.Record{ entities.NAMEATTR : fmt.Sprintf("%v", record[entities.NAMEATTR]) + ":" + role, }
		for perms, value := range mainPerms { rec[perms]=value }
		rec[tool.SELECT.String()]=entities.LEVELNORMAL
		s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBPermission.Name, tool.RootRowsParam : tool.ReservedParam }, 
			rec, tool.CREATE, "CreateOrUpdate",)
	}
}
func (s *SchemaService) WriteRowAutomation(record tool.Record, tableName string) { 
	s.Domain.SuperCall(
		tool.Params{ tool.RootTableParam : record[entities.NAMEATTR].(string), }, 
		tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR], 
			         "columns": map[string]interface{}{} }, 
				     tool.CREATE, "CreateOrUpdate",)
	s.UpdateRowAutomation(nil, record)
}
func (s *SchemaService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	res := tool.Results{}
	for _, rec := range results {
		if s.Domain.PermsCheck(fmt.Sprintf("%v", rec[entities.NAMEATTR]), "", "", tool.SELECT) {
			res = append(res, rec)
		}
	}
	return s.Domain.PostTreat( res, tableName) 
}
func (s *SchemaService) ConfigureFilter(tableName string) (string, string) {
	return s.Domain.ViewDefinition(tableName)
}	
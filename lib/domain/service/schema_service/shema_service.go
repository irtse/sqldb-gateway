package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type SchemaService struct { tool.AbstractSpecializedService }

func (s *SchemaService) Entity() tool.SpecializedServiceInfo { return entities.DBSchema }
func (s *SchemaService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	res, _ := s.Domain.SafeCall(true, "",
		tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
					 tool.RootRowsParam : tool.ReservedParam, 
				     entities.NAMEATTR: fmt.Sprintf("%v", record[entities.NAMEATTR]) }, 
		tool.Record{}, 
		tool.SELECT, 
		"Get")
	return record, res == nil || len(res) == 0
}
func (s *SchemaService) DeleteRowAutomation(results tool.Results) { 
	for _, record := range results { 
		s.Domain.SetIsCustom(true)
		s.Domain.SafeCall(true, "", 
		                	tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
			                                tool.RootRowsParam: tool.ReservedParam,
			 	                            entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%d", record["id"].(int64)) }, 
							tool.Record{ }, 
							tool.DELETE, 
							"Delete",
						)
	}
}
func (s *SchemaService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *SchemaService) WriteRowAutomation(record tool.Record) { 
	s.Domain.SafeCall(true, "",
		tool.Params{ tool.RootTableParam : record[entities.NAMEATTR].(string), }, 
		tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR], 
			    "columns": map[string]interface{}{} }, 
				tool.CREATE, 
				"CreateOrUpdate",)
}
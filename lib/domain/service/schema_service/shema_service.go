package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type SchemaService struct { tool.AbstractSpecializedService }

func (s *SchemaService) Entity() tool.SpecializedServiceInfo { return entities.DBSchema }
func (s *SchemaService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *SchemaService) DeleteRowAutomation(results tool.Results) { 
	for _, record := range results { 
		s.Domain.SetIsCustom(true)
		s.Domain.SuperCall( 
		                	tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
			                                tool.RootRowsParam: tool.ReservedParam,
			 	                            entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", record["id"]) }, 
							tool.Record{ }, 
							tool.DELETE, 
							"Delete",
						)
	}
}
func (s *SchemaService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *SchemaService) WriteRowAutomation(record tool.Record) { 
	s.Domain.SuperCall(
		tool.Params{ tool.RootTableParam : record[entities.NAMEATTR].(string), }, 
		tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR], 
			    "columns": map[string]interface{}{} }, 
				tool.CREATE, "CreateOrUpdate",)
}
func (s *SchemaService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName) 
}
func (s *SchemaService) ConfigureFilter(tableName string, params tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
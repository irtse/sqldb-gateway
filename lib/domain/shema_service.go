package domain

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type SchemaService struct {
	Domain tool.DomainITF
}

func (s *SchemaService) SetDomain(d tool.DomainITF) { s.Domain = d }
func (s *SchemaService) Entity() tool.SpecializedServiceInfo { return entities.DBSchema }
func (s *SchemaService) VerifyRowWorkflow(record tool.Record, create bool) (tool.Record, bool) { 
	res, _ := s.Domain.SafeCall(true, "",
		tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
					 tool.RootRowsParam : tool.ReservedParam, 
				     entities.NAMEATTR: fmt.Sprintf("%v", record[entities.NAMEATTR]) }, 
		tool.Record{}, 
		tool.SELECT, 
		"Get")
	return record, res == nil || len(res) == 0
}
func (s *SchemaService) DeleteRowWorkflow(results tool.Results) { 
	for _, record := range results { 
		s.Domain.(*MainService).isGenericService=true
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
func (s *SchemaService) UpdateRowWorkflow(results tool.Results, record tool.Record) {}
func (s *SchemaService) WriteRowWorkflow(record tool.Record) { 
	s.Domain.SafeCall(true, "",
		tool.Params{ tool.RootTableParam : record[entities.NAMEATTR].(string), }, 
		tool.Record{ entities.NAMEATTR : record[entities.NAMEATTR], 
			    "columns": map[string]interface{}{} }, 
				tool.CREATE, 
				"CreateOrUpdate",)
}
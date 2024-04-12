package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)
type WorkflowService struct { tool.AbstractSpecializedService }

func (s *WorkflowService) Entity() tool.SpecializedServiceInfo { return entities.DBWorkflow }
func (s *WorkflowService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *WorkflowService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *WorkflowService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *WorkflowService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *WorkflowService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 
	res := tool.Results{}	
	for _, rec := range results {
		params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
							   tool.RootRowsParam : tool.ReservedParam,
							   tool.SpecialIDParam : rec.GetString(entities.RootID(entities.DBSchema.Name)), }
		schemas, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(schemas) == 0 || !s.Domain.PermsCheck(fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]), "", "", tool.CREATE) { 
			continue 
		}
		res = append(res, rec)
	}
	return s.Domain.PostTreat( res, tableName) 
}
func (s *WorkflowService) ConfigureFilter(tableName string) (string, string) {
	restr := "is_meta=false"
	return s.Domain.ViewDefinition(tableName, restr)
}
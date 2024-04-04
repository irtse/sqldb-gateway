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
	return s.Domain.PostTreat( results, tableName) 
}
func (s *WorkflowService) ConfigureFilter(tableName string) (string, string) {
	rows, ok := s.Domain.GetParams()[tool.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[tool.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != tool.ReservedParam) || (ok2 && ids != "") {
		return s.Domain.ViewDefinition(tableName)
	}
	restr := "is_meta=false"
	return s.Domain.ViewDefinition(tableName, restr)
}
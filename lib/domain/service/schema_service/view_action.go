package schema_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type ViewActionService struct { tool.AbstractSpecializedService }

func (s *ViewActionService) Entity() tool.SpecializedServiceInfo { return entities.DBViewAction }
func (s *ViewActionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *ViewActionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *ViewActionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ViewActionService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *ViewActionService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *ViewActionService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}
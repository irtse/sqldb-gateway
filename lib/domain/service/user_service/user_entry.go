package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type UserEntryService struct { tool.AbstractSpecializedService }

func (s *UserEntryService) Entity() tool.SpecializedServiceInfo { return entities.DBUserEntry }
func (s *UserEntryService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *UserEntryService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *UserEntryService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *UserEntryService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *UserEntryService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *UserEntryService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}
package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type UserService struct { tool.AbstractSpecializedService }

func (s *UserService) Entity() tool.SpecializedServiceInfo { return entities.DBUser }
func (s *UserService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *UserService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *UserService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *UserService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *UserService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *UserService) ConfigureFilter(tableName string) (string, string) {
	restr := "id IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ") "
	return s.Domain.ViewDefinition(tableName, restr)
}
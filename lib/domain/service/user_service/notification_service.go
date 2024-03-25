package user_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type NotificationService struct { tool.AbstractSpecializedService }

func (s *NotificationService) Entity() tool.SpecializedServiceInfo { return entities.DBNotification }
func (s *NotificationService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *NotificationService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *NotificationService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *NotificationService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *NotificationService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName) 
}
func (s *NotificationService) ConfigureFilter(tableName string) (string, string) {
	rows, ok := s.Domain.GetParams()[tool.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[tool.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != tool.ReservedParam) || (ok2 && ids != "") {
		return s.Domain.ViewDefinition(tableName)
	}
	restr := entities.RootID(entities.DBUser.Name)  + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ") OR " 
	restr += entities.RootID(entities.DBEntity.Name) + " IN (SELECT id FROM " + entities.DBEntityUser.Name + " WHERE id IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + "))" 
	return s.Domain.ViewDefinition(tableName, restr)
}
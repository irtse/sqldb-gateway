package task_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskWatcherService struct { tool.AbstractSpecializedService }

func (s *TaskWatcherService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskWatcher }
func (s *TaskWatcherService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *TaskWatcherService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *TaskWatcherService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *TaskWatcherService) WriteRowAutomation(record tool.Record, tableName string) {}
func (s *TaskWatcherService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *TaskWatcherService) ConfigureFilter(tableName string) (string, string) {
	restr := entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")" 
	restr += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	restr += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	restr += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")"
	return s.Domain.ViewDefinition(tableName, restr)
}	
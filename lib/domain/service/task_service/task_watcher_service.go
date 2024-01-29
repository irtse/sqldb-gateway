package task_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type TaskWatcherService struct { tool.AbstractSpecializedService }

func (s *TaskWatcherService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskWatcher }
func (s *TaskWatcherService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *TaskWatcherService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *TaskWatcherService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *TaskWatcherService) WriteRowAutomation(record tool.Record, tableName string) {}
func (s *TaskWatcherService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *TaskWatcherService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	params[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')"
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
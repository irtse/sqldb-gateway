package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskAssigneeService struct { tool.AbstractSpecializedService }

func (s *TaskAssigneeService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskAssignee }
func (s *TaskAssigneeService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	var res tool.Results
	if taskID, ok := record[entities.RootID(entities.DBTask.Name)]; ok && taskID != nil {
		if userID, ok := record[entities.RootID(entities.DBUser.Name)]; ok && userID != nil {
			res, _ = s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBTaskAssignee.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBTask.Name)]),
						 entities.RootID(entities.DBUser.Name): fmt.Sprintf("%v", record[entities.RootID(entities.DBUser.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		} else if entityID, ok := record[entities.RootID(entities.DBEntity.Name)]; ok && entityID != nil {
			res, _ = s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBTaskAssignee.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBEntity.Name): fmt.Sprintf("%v", record[entities.RootID(entities.DBEntity.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		}
	}
	return record, res == nil || len(res) == 0
}
func (s *TaskAssigneeService) DeleteRowAutomation(results tool.Results) { }
func (s *TaskAssigneeService) UpdateRowAutomation(results tool.Results, record tool.Record) {
	for _, res := range results { s.WriteRowAutomation(res) }
}
func (s *TaskAssigneeService) WriteRowAutomation(record tool.Record) { 
	paramsNew := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
							  tool.RootRowsParam: tool.ReservedParam, }
	paramsNew[tool.RootSQLFilterParam] += "id IN (SELECT id FROM " + entities.DBHierarchy.Name + " WHERE "
	paramsNew[tool.RootSQLFilterParam] += entities.RootID(entities.DBUser.Name) + " IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")"
	paramsNew[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name
	paramsNew[tool.RootSQLFilterParam] += " WHERE " + entities.RootID(entities.DBUser.Name) + "=" + conn.Quote(s.Domain.GetUser()) + "))"
	hierarchy, err := s.Domain.SuperCall( 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
					)
	if err == nil {
		for _, upper := range hierarchy {
			s.Domain.SuperCall(
							tool.Params{ 
								tool.RootTableParam : entities.DBTaskWatcher.Name,
								tool.RootRowsParam : tool.ReservedParam,
								entities.RootID(entities.DBUser.Name) : fmt.Sprintf("%v", upper[tool.SpecialIDParam]),
								entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBTask.Name)]),
							},
							tool.Record{ entities.RootID(entities.DBUser.Name) : upper[tool.SpecialIDParam],
										 entities.RootID(entities.DBTask.Name) : record[entities.RootID(entities.DBTask.Name)],
									   },
								  tool.CREATE,
								  "CreateOrUpdate",
							)
		}
	}
	
}
func (s *TaskAssigneeService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *TaskAssigneeService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	params[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')"
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
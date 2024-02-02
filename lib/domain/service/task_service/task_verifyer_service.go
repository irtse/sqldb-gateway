package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskVerifyerService struct { tool.AbstractSpecializedService }

func (s *TaskVerifyerService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskVerifyer }
func (s *TaskVerifyerService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	var res tool.Results
	if taskID, ok := record[entities.RootID(entities.DBTask.Name)]; ok && taskID != nil {
		if userID, ok := record[entities.RootID(entities.DBUser.Name)]; ok && userID != nil {
			res, _ = s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBTask.Name)]),
						 entities.RootID(entities.DBUser.Name): fmt.Sprintf("%v", record[entities.RootID(entities.DBUser.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		} else if entityID, ok := record[entities.RootID(entities.DBEntity.Name)]; ok && entityID != nil {
			res, _ = s.Domain.SuperCall(
			tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBEntity.Name): fmt.Sprintf("%v", record[entities.RootID(entities.DBEntity.Name)]) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		}
	}
	return record, res == nil || len(res) == 0, false
}
func (s *TaskVerifyerService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *TaskVerifyerService) UpdateRowAutomation(results tool.Results, record tool.Record) {
	if state, ok := record["state"]; ok && state != "completed" { return }
	for _, rec := range results {
		if id, ok2 := rec[entities.RootID(entities.DBTask.Name)]; ok2 {
			if state, ok3 := rec["state"]; ok3 && state == "rejected" {
				params := tool.Params{ tool.RootTableParam : entities.DBTaskAssignee.Name, 
					                   tool.RootRowsParam: tool.ReservedParam,
									   entities.RootID(entities.DBTask.Name): fmt.Sprintf("%v", id), }
				s.Domain.SuperCall( params, tool.Record{ "state": "pending" }, tool.UPDATE, "CreateOrUpdate", )
			}
			paramsNew := tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
				                      tool.RootRowsParam: tool.ReservedParam, }
			paramsNew[entities.RootID(entities.DBTask.Name)] = fmt.Sprintf("%v", id)
			sqlFilter := "state != 'completed'"
			unfinished, err := s.Domain.SuperCall( 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
						sqlFilter,
					)
			if len(unfinished) > 0 || err != nil  { continue }
			paramsNew = tool.Params{ tool.RootTableParam : entities.DBTaskAssignee.Name, 
				                      tool.RootRowsParam: tool.ReservedParam, }
			paramsNew[entities.RootID(entities.DBTask.Name)] = fmt.Sprintf("%v", id)
			sqlFilter = "state != 'completed'"
			unfinishedAssign, err := s.Domain.SuperCall( 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
						sqlFilter,
					)
			if len(unfinishedAssign) > 0  || err != nil { continue }
			// TODO when all is verified
			s.Domain.SuperCall(
							tool.Params{ 
								tool.RootTableParam : entities.DBTask.Name,
								tool.RootRowsParam : tool.ReservedParam,
							},
							tool.Record{ tool.SpecialIDParam : id, "state" : "close" },
							tool.UPDATE,
							"CreateOrUpdate",
						)
		}
		// TODO IF VERIFYER DISMISS THE TASK 
	}
}
func (s *TaskVerifyerService) WriteRowAutomation(record tool.Record, tableName string) {}
func (s *TaskVerifyerService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *TaskVerifyerService) ConfigureFilter(tableName string) (string, string) {
	restr := entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")" 
	return s.Domain.ViewDefinition(tableName, restr)
}	
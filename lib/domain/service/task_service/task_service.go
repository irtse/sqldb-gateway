package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type TaskService struct { tool.AbstractSpecializedService }

func (s *TaskService) Entity() tool.SpecializedServiceInfo { return entities.DBTask }
func (s *TaskService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	if state, ok := record["state"]; ok && (state == "open" || state == "close" || create) { 
		params := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
			                   tool.RootRowsParam : tool.ReservedParam, 
			                   "login" : s.Domain.GetUser(),
		}
		user, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(user) == 0 { return record, false }
		if state == "open" { 
			record["opened_by"]=user[0][tool.SpecialIDParam] 
			record["opened_date"]="CURRENT_TIMESTAMP"
		}
		if create { record["created_by"]=user[0][tool.SpecialIDParam] }
	}
	return record, true 
}
func (s *TaskService) DeleteRowAutomation(results tool.Results) { 
	for _, res := range results {
		res["state"]="close"
	}
	s.UpdateRowAutomation(results, tool.Record{})
}
func (s *TaskService) UpdateRowAutomation(results tool.Results, record tool.Record) {
	for _, res := range results {
		// TODO in completion : seek
		if state, ok := res["state"]; !ok || state != "close" { continue }
		if workflowID, ok2 := res[entities.RootID(entities.DBWorkflow.Name)]; ok2{ 
			params := tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name, 
				                   tool.RootRowsParam : tool.ReservedParam, 
								   entities.RootID(entities.DBSchema.Name) : fmt.Sprintf(
										"%d", res[entities.RootID(entities.DBSchema.Name)].(int64)),
				                   entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%d", workflowID.(int64)),
			                     }
			schemas, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
			if err != nil || len(schemas) == 0 { continue }
			if order, ok3 := schemas[0]["order"]; ok3 {
				params := tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name, 
					tool.RootRowsParam : tool.ReservedParam, 
					entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%d", workflowID.(int64)),
					"order": fmt.Sprintf("%d", order.(int64) + 1,),
				}
				uppers, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(uppers) == 0 { continue }
				params = tool.Params{ tool.RootTableParam : entities.DBWorkflowTask.Name, 
					                   tool.RootRowsParam : tool.ReservedParam, 
									   entities.RootID(entities.DBWorkflowSchema.Name) : fmt.Sprintf(
										"%d", uppers[0][tool.ReservedParam].(int64)),
									 }
				wbTasks, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
				if err != nil { continue }
				for _, wbTask := range wbTasks {
					dbs := []string{entities.DBTaskAssignee.Name, entities.DBTaskVerifyer.Name,entities.DBTaskWatcher.Name}
					for _, dbName := range dbs {
						s.Domain.SafeCall(true, "", 
					                  tool.Params{ 
										tool.RootTableParam : dbName, 
					                    tool.RootRowsParam : tool.ReservedParam,
										entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%d", wbTask[entities.RootID(entities.DBTask.Name)].(int64)),
									  }, 
									  tool.Record{ "hidden": false, }, 
									  tool.UPDATE, "CreateOrUpdate")
					}
				}
			}
	    }
	}
}
func (s *TaskService) WriteRowAutomation(record tool.Record) {}
package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type TaskService struct { tool.AbstractSpecializedService }

func (s *TaskService) Entity() tool.SpecializedServiceInfo { return entities.DBTask }
func (s *TaskService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	schemas, err := s.Domain.Schema(record)
	if err != nil && len(schemas) == 0 { return record, false, false }
	id := int64(-1)
	if idFromRec, ok := record[tool.SpecialIDParam]; ok { id = idFromRec.(int64) }
	if idFromTask, ok := record[entities.RootID("dest_table")]; ok { id = idFromTask.(int64) }
	if id == -1 { return record, false, false }
	params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			                   tool.RootRowsParam : fmt.Sprintf("%v", id), } // empty record
	s.Domain.SuperCall( params, record, tool.UPDATE, "CreateOrUpdate")
	if _, ok := record[entities.RootID("dest_table")]; ok && !create { // TODO if not superadmin PROTECTED
		delete(record, entities.RootID("dest_table"))
	}
	if state, ok := record["state"]; ok && (state == "in progress" || create) { 
		params := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
			                   tool.RootRowsParam : tool.ReservedParam, 
			                   "login" : s.Domain.GetUser(),
		}
		user, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(user) == 0 { return record, false, false }
		if user[0]["state"] != "in progress"  { 
			record[entities.RootID("opened_by")]=user[0][tool.SpecialIDParam] 
			record["opened_date"]="CURRENT_TIMESTAMP"
		}
		if create { record[entities.RootID("created_by")]=user[0][tool.SpecialIDParam] }
	}
	return record, true , false
}
func (s *TaskService) DeleteRowAutomation(results tool.Results, tableName string) { 
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
										"%v", res[entities.RootID(entities.DBSchema.Name)]),
				                   entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", workflowID),
			                     }
			schemas, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
			if err != nil || len(schemas) == 0 { continue }
			if order, ok3 := schemas[0]["index"]; ok3 {
				params := tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name, 
					tool.RootRowsParam : tool.ReservedParam, 
					entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", workflowID),
					"index": fmt.Sprintf("%v", order.(int64) + 1,),
				}
				afterScheme, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(afterScheme) == 0 { continue }
				params = tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name, 
					tool.RootRowsParam : tool.ReservedParam, 
					entities.RootID(entities.DBSchema.Name) : fmt.Sprintf("%v", afterScheme[0][entities.RootID(entities.DBSchema.Name)]),
					entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", afterScheme[0][entities.RootID(entities.DBWorkflow.Name)]),
				}
				tasks, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(tasks) == 0 { continue }
				dbs := []string{entities.DBTaskAssignee.Name, entities.DBTaskVerifyer.Name,entities.DBTaskWatcher.Name}
				for _, dbName := range dbs {
					s.Domain.SuperCall( 
					            tool.Params{ 
									tool.RootTableParam : dbName, 
					                tool.RootRowsParam : tool.ReservedParam,
									entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%v", tasks[0][tool.SpecialIDParam]),
								}, 
								tool.Record{ "hidden": false, }, 
								tool.UPDATE, "CreateOrUpdate")
				}
			}
	    }
	}
}
func (s *TaskService) WriteRowAutomation(record tool.Record, tableName string) {
	// task creation automation.
	schemas, err := s.Domain.Schema(record)
	if err != nil && len(schemas) == 0 { return }
	params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			              tool.RootRowsParam : tool.ReservedParam, } // empty record
	created, err := s.Domain.SuperCall( params, tool.Record{}, tool.CREATE, "CreateOrUpdate")
	if err != nil && len(created) == 0 { return }
	newRec := tool.Record{ entities.RootID("dest_table"): created[0][tool.SpecialIDParam] }
	params = tool.Params{ tool.RootTableParam : s.Entity().GetName(), 
							  tool.RootRowsParam : tool.ReservedParam, } 
	s.Domain.SuperCall( params, newRec, tool.UPDATE, "CreateOrUpdate")
	s.Domain.WriteRow(tableName, record)
}

func (s *TaskService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}

func (s *TaskService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT " + entities.RootID(entities.DBTask.Name) + " FROM " + entities.DBTaskWatcher.Name + " WHERE "
	params[tool.RootSQLFilterParam] += entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	params[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')))"
	params[tool.RootSQLFilterParam] += " OR id IN (SELECT " + entities.RootID(entities.DBTask.Name) + " FROM " + entities.DBTaskAssignee.Name + " WHERE "
	params[tool.RootSQLFilterParam] += entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	params[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')))"
	params[tool.RootSQLFilterParam] += " OR id IN (SELECT " + entities.RootID(entities.DBTask.Name) + " FROM " + entities.DBTaskVerifyer.Name + " WHERE "
	params[tool.RootSQLFilterParam] += entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "'))" 
	return s.Domain.ViewDefinition(tableName, params)
}	
package task_service

import (
	"fmt"
	"errors"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type TaskService struct { tool.AbstractSpecializedService }

func (s *TaskService) Entity() tool.SpecializedServiceInfo { return entities.DBTask }
func (s *TaskService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	// TODO if "form" presence THEN -> update row and kick form out of record
	if form, ok := record["form"]; ok {
		rec := tool.Record{}
		for k, v := range form.(map[string]tool.Record) { rec[k]=v["value"] }
		delete(record, "form")
		schemas, err := s.schema(record)
		if err != nil && len(schemas) == 0 { return record, false }
		id := int64(-1)
		if idFromRec, ok := rec[tool.SpecialIDParam]; ok { id = idFromRec.(int64) }
		if idFromTask, ok := record[entities.RootID("dest_table")]; ok { id = idFromTask.(int64) }
		if id == -1 { return record, false }
		params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			                   tool.RootRowsParam : fmt.Sprintf("%d", id), } // empty record
		s.Domain.SafeCall(true, "", params, rec, tool.UPDATE, "CreateOrUpdate")
	}
	if _, ok := record[entities.RootID("dest_table")]; ok && !create { // TODO if not superadmin PROTECTED
		delete(record, entities.RootID("dest_table"))
	}
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
func (s *TaskService) WriteRowAutomation(record tool.Record) {
	// task creation automation.
	schemas, err := s.schema(record)
	if err != nil && len(schemas) == 0 { return }
	params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			              tool.RootRowsParam : tool.ReservedParam,
	} // empty record
	created, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.CREATE, "CreateOrUpdate")
	if err != nil && len(created) == 0 { return }
	newRec := tool.Record{ entities.RootID("dest_table"): created[0][tool.SpecialIDParam] }
	params = tool.Params{ tool.RootTableParam : s.Entity().GetName(), 
							  tool.RootRowsParam : tool.ReservedParam,
						} 
	s.Domain.SafeCall(true, "", params, newRec, tool.UPDATE, "CreateOrUpdate")
}

func (s *TaskService) PostTreatment(results tool.Results) tool.Results { 
	for _, record := range results {
		if dest_id, ok:= record[entities.RootID("dest_table")]; !ok || dest_id == nil { continue }
		schemas, err := s.schema(record)
		if err != nil && len(schemas) == 0 { return results }
		params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			                   tool.RootRowsParam : fmt.Sprintf("%d", record[entities.RootID("dest_table")].(int64)),}
		rows, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
		if err != nil && len(rows) == 0 { return results }
		record["form"] = map[string]tool.Record{}
		form := map[string]tool.Record{}
		params = tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
			tool.RootRowsParam : tool.ReservedParam, 
			entities.RootID(entities.DBSchema.Name): fmt.Sprintf("%d", schemas[0][tool.SpecialIDParam].(int64)),
		}
		fields, err := s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
		for _, row := range rows {
			for k, v := range row {
				if k == tool.SpecialIDParam { continue }
				for _, field := range fields {
					if field[entities.NAMEATTR].(string) == k && !field["hidden"].(bool) {
						form[k]=field
						form[k]["value"]=v
					}
				}
			}
		}
		record["form"] = form
	}
	return results 
}

func (s *TaskService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}	

func (s *TaskService) schema(record tool.Record) (tool.Results, error) {
	if schemaID, ok := record[entities.RootID(entities.DBSchema.Name)]; ok {
		params := tool.Params{ tool.RootTableParam : entities.DBSchema.Name, 
			tool.RootRowsParam : tool.ReservedParam, 
			tool.SpecialIDParam : fmt.Sprintf("%d", schemaID.(int64)),
		}
		return s.Domain.SafeCall(true, "", params, tool.Record{}, tool.SELECT, "Get")
	}
	return nil, errors.New("no schemaID refered...")
}
package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskService struct { tool.AbstractSpecializedService }

func (s *TaskService) Entity() tool.SpecializedServiceInfo { return entities.DBTask }
func (s *TaskService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	if create { 
		sqlFilter := "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser())
		params := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
			                   tool.RootRowsParam : tool.ReservedParam, }
		user, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		if err != nil || len(user) == 0 { return record, false, false }
		record[entities.RootID("created_by")]=user[0][tool.SpecialIDParam]  // affected create_by
	} else {
		if record["state"] == "completed" || record["state"] == "dismiss" { record["is_close"] = true 
		} else { record["state"] = "progressing"  }
	}
	return record, true, true
}
func (s *TaskService) DeleteRowAutomation(results tool.Results, tableName string) { 
	for _, res := range results {
		res["state"]="completed"
		res["is_close"]=true
	}
	s.UpdateRowAutomation(results, tool.Record{})
}
func (s *TaskService) UpdateRowAutomation(results tool.Results, record tool.Record) {
	for _, res := range results {
		if res["state"] != "completed" && res["state"] != "dismiss" { continue }
		// retrieve binded demand
		paramsReq := tool.Params{ tool.RootTableParam : entities.DBRequest.Name, tool.RootRowsParam : tool.ReservedParam, }
		requests, err := s.Domain.SuperCall( paramsReq, tool.Record{}, tool.SELECT, "Get", "id=" + res.GetString(entities.RootID(entities.DBRequest.Name)))
		if err != nil || len(requests) == 0 { continue }
		if order, ok3 := requests[0]["current_index"]; ok3 {
			if order.(int64) > 0 {
				params := tool.Params{ tool.RootTableParam : entities.DBTask.Name, tool.RootRowsParam : tool.ReservedParam, 
					entities.RootID(entities.DBRequest.Name) : fmt.Sprintf("%v", res[entities.RootID(entities.DBRequest.Name)]),
				}
				otherPendingTasks, _ := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", "state IN ('pending', 'progressing', 'dismiss') AND (is_close=false)")
				if len(otherPendingTasks) > 0 { continue }
			}
			current_index := order.(int64)
			params := tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name, 
								   tool.RootRowsParam : tool.ReservedParam,
								   entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", requests[0][entities.RootID(entities.DBWorkflow.Name)])  }
			if res["state"] == "completed" { current_index++ }	
			if res["state"] == "dismiss" {
				if order.(int64) > 0 { current_index--
				} else { 
					params = tool.Params{ tool.RootTableParam : entities.DBRequest.Name, tool.RootRowsParam: requests[0].GetString(tool.SpecialIDParam) }
					s.Domain.PermsSuperCall( params, tool.Record{ "state" : "dismiss", "is_close": true }, tool.UPDATE, "CreateOrUpdate")
				} // no before task close request and task
			}
			scheme, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", "index=" + fmt.Sprintf("%v", current_index))
			newRecRequest := requests[0]
			if err != nil || len(scheme) == 0 { 
				newRecRequest["state"] = "completed"
				newRecRequest["is_close"] = true
			} else {
				newRecRequest["current_index"]=current_index
				newRecRequest["state"] = "progressing"
				newRecRequest["is_close"] = false
			} // no new task in workflow
			_,err = s.Domain.SuperCall( tool.Params{ tool.RootTableParam : entities.DBRequest.Name, tool.RootRowsParam : tool.ReservedParam, }, 
				                             newRecRequest, tool.UPDATE, "CreateOrUpdate")
			if err != nil || len(scheme) == 0 { continue }
			params = tool.Params{ tool.RootTableParam : entities.DBTask.Name, 
				                   tool.RootRowsParam: tool.ReservedParam,
								   entities.RootID(entities.DBWorkflowSchema.Name) : scheme[0].GetString(tool.SpecialIDParam),
								   entities.RootID(entities.DBRequest.Name) : requests[0].GetString(tool.SpecialIDParam), }
			beforeTask, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", "is_close=false")
			if err == nil && len(beforeTask) > 0 { continue }
			newTask := tool.Record{
				entities.RootID(entities.DBSchema.Name) : scheme[0][entities.RootID(entities.DBSchema.Name)],
				entities.RootID(entities.DBWorkflowSchema.Name) : scheme[0][tool.SpecialIDParam],
				entities.RootID(entities.DBRequest.Name) : requests[0][tool.SpecialIDParam],
				entities.RootID("created_by") : res[entities.RootID("created_by")],
				"description" : scheme[0]["description"],
				"urgency" : scheme[0]["urgency"],
				"priority" : scheme[0]["priority"],
				entities.NAMEATTR : scheme[0][entities.NAMEATTR],
			}
			if fmt.Sprintf("%v", scheme[0][entities.RootID(entities.DBSchema.Name)]) == fmt.Sprintf("%v", res[entities.RootID(entities.DBSchema.Name)]) {
				newTask[entities.RootID(entities.DBSchema.Name)] = res[entities.RootID(entities.DBSchema.Name)]
				newTask[entities.RootID("dest_table")] = res[entities.RootID("dest_table")]
			} else {
				schemas, err := s.Domain.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : scheme[0][entities.RootID(entities.DBSchema.Name)] }, false)
				if err == nil && len(schemas) > 0 {
					vals, err := s.Domain.SuperCall( tool.Params{ tool.RootTableParam: schemas[0].GetString(entities.NAMEATTR),
						tool.RootRowsParam: tool.ReservedParam, }, tool.Record{}, tool.CREATE, "CreateOrUpdate")
					if err == nil && len(vals) > 0 {
						newTask[entities.RootID(entities.DBSchema.Name)] = scheme[0][entities.RootID(entities.DBSchema.Name)]
						newTask[entities.RootID("dest_table")] = vals[0][tool.ReservedParam]
					}
				}
			}
			params = tool.Params{ tool.RootTableParam : entities.DBTask.Name, tool.RootRowsParam : tool.ReservedParam, }
			s.Domain.SuperCall( params, newTask, tool.CREATE, "CreateOrUpdate")
	    }
	}
}
func (s *TaskService) WriteRowAutomation(record tool.Record, tableName string) {
	// task creation automation.
	schemas, err := s.Domain.Schema(record, true)
	if err != nil && len(schemas) == 0 { return }
	params := tool.Params{ tool.RootTableParam : schemas[0][entities.NAMEATTR].(string), 
			              tool.RootRowsParam : tool.ReservedParam, } // empty record
	created, err := s.Domain.SuperCall( params, tool.Record{}, tool.CREATE, "CreateOrUpdate")
	if err != nil && len(created) == 0 { return }
	newRec := tool.Record{ entities.RootID("dest_table"): created[0][tool.SpecialIDParam] }
	params = tool.Params{ tool.RootTableParam : s.Entity().GetName(), 
							  tool.RootRowsParam : tool.ReservedParam, } 
	s.Domain.SuperCall( params, newRec, tool.UPDATE, "CreateOrUpdate")
}

func (s *TaskService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 
	return s.Domain.PostTreat(results, tableName) 
}

func (s *TaskService) ConfigureFilter(tableName string) (string, string) {
	rows, ok := s.Domain.GetParams()[tool.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[tool.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != tool.ReservedParam) || (ok2 && ids != "") {
		return s.Domain.ViewDefinition(tableName)
	}
	restr := entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ") OR "
	restr += entities.RootID(entities.DBEntity.Name) + " IN (SELECT id FROM " + entities.DBEntityUser.Name + " WHERE " + entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")) OR "
	restr += entities.RootID(entities.DBRequest.Name) + " IN (SELECT id FROM " + entities.DBRequest.Name + " WHERE "
	restr += entities.RootID(entities.DBWorkflowSchema.Name) + " IN (SELECT id FROM " + entities.DBWorkflowSchema.Name + " WHERE "
	restr += entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")" 
	restr += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	restr += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	restr += " WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + "))))"
	return s.Domain.ViewDefinition(tableName, restr)
}	
package task_service

import (
	"fmt"
	"time"
	"strings"
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskService struct { utils.AbstractSpecializedService }

func (s *TaskService) Entity() utils.SpecializedServiceInfo { return schserv.DBTask }
func (s *TaskService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results { return s.Domain.PostTreat(results, tableName, true) }
func (s *TaskService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, bool, bool) { 
	if s.Domain.GetMethod() == utils.CREATE { 
		sqlFilter := "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser())
		user, err := s.Domain.SuperCall(utils.AllParams(schserv.DBUser.Name), utils.Record{}, utils.SELECT, sqlFilter)
		if err != nil || len(user) == 0 { return record, false, false }
		record[schserv.DBUser.Name]=user[0][utils.SpecialIDParam]  // affected create_by
	} else if s.Domain.GetMethod() == utils.UPDATE {
		if record["state"] == "completed" || record["state"] == "dismiss" { 
			record["is_close"] = true 
			record["closing_date"] = time.Now().Format(time.RFC3339)
		} else { record["state"] = "progressing"  }
	}
	if s.Domain.GetMethod() != utils.DELETE {
		rec, err := s.Domain.ValidateBySchema(record, tablename)
		if err != nil && !s.Domain.GetAutoload() { return rec, false, false } else { rec = record }
		return rec, true, true
	}
	return record, true, true
}
func (s *TaskService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { 
	for _, res := range results {
		res["state"]="completed"
		res["is_close"]=true
		res["closing_date"] = time.Now().Format(time.RFC3339)
	}
	s.UpdateRowAutomation(results, map[string]interface{}{})
}
func (s *TaskService) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {
	for _, res := range results {
		if res["state"] != "completed" && res["state"] != "dismiss" { continue }
		paramsReq := utils.Params{ utils.RootTableParam : schserv.DBRequest.Name, 
								  utils.RootRowsParam : utils.GetString(res, schserv.RootID(schserv.DBRequest.Name)), }
		requests, err := s.Domain.SuperCall( paramsReq, utils.Record{}, utils.SELECT)
		if err != nil || len(requests) == 0 { continue }
		if order, ok3 := requests[0]["current_index"]; ok3 {
			if order.(int64) > 0 {
				params := utils.Params{ utils.RootTableParam : schserv.DBTask.Name, utils.RootRowsParam : utils.ReservedParam, 
					schserv.RootID(schserv.DBRequest.Name) : fmt.Sprintf("%v", res[schserv.RootID(schserv.DBRequest.Name)]), }
				otherPendingTasks, _ := s.Domain.SuperCall( params, utils.Record{}, utils.SELECT,
					"state IN ('pending', 'progressing', 'dismiss') AND (is_close=false)")
				if len(otherPendingTasks) > 0 { continue }
			}
			current_index := order.(int64)
			if res["state"] == "completed" { current_index++ }	
			if res["state"] == "dismiss" {
				if order.(int64) > 0 { current_index-- } else {  // Dismiss will close requests.
					s.Domain.Call( paramsReq, utils.Record{ "state" : "dismiss",  "is_close": true, 
						"closing_date" : time.Now().Format(time.RFC3339) }, utils.UPDATE)
				} // no before task close request and task
			}
			params := utils.Params{ utils.RootTableParam : schserv.DBWorkflowSchema.Name, utils.RootRowsParam : utils.ReservedParam, 
				schserv.RootID(schserv.DBWorkflow.Name) : fmt.Sprintf("%v", requests[0][schserv.RootID(schserv.DBWorkflow.Name)])  }
			schemes, err := s.Domain.SuperCall( params, utils.Record{}, utils.SELECT, "index=" + fmt.Sprintf("%v", current_index))
			newRecRequest := utils.Record{ utils.SpecialIDParam : requests[0][utils.SpecialIDParam]}
			if err != nil || len(schemes) == 0 { // no new task in workflow
				newRecRequest["state"] = "completed"
				newRecRequest["is_close"] = true
				newRecRequest["closing_date"] = time.Now().Format(time.RFC3339)
			} else {
				newRecRequest["current_index"]=current_index
				newRecRequest["state"] = "progressing"
				newRecRequest["is_close"] = false
			} 
			s.Domain.PermsSuperCall(utils.AllParams(schserv.DBRequest.Name), newRecRequest, utils.UPDATE)
			if err != nil || len(schemes) == 0 { continue }
			for _, scheme := range schemes {
				params = utils.Params{ utils.RootTableParam : schserv.DBTask.Name, 
									  utils.RootRowsParam: utils.ReservedParam,
									  schserv.RootID(schserv.DBWorkflowSchema.Name) : scheme.GetString(utils.SpecialIDParam),
									  schserv.RootID(schserv.DBRequest.Name) : requests[0].GetString(utils.SpecialIDParam), }
				beforeTask, err := s.Domain.SuperCall( params, utils.Record{}, utils.SELECT, "is_close=false")
				if err == nil && len(beforeTask) > 0 { continue }
				newTask := utils.Record{
					schserv.RootID(schserv.DBSchema.Name) : scheme[schserv.RootID(schserv.DBSchema.Name)],
					schserv.RootID(schserv.DBWorkflowSchema.Name) : scheme[utils.SpecialIDParam],
					schserv.RootID(schserv.DBRequest.Name) : requests[0][utils.SpecialIDParam],
					schserv.RootID(schserv.DBUser.Name) : res[schserv.RootID(schserv.DBUser.Name)],
					"description" : scheme["description"],
					"urgency" : scheme["urgency"],
					"priority" : scheme["priority"],
					schserv.NAMEKEY : scheme[schserv.NAMEKEY],
				}
				if fmt.Sprintf("%v", scheme[schserv.RootID(schserv.DBSchema.Name)]) == fmt.Sprintf("%v", res[schserv.RootID(schserv.DBSchema.Name)]) {
					newTask[schserv.RootID(schserv.DBSchema.Name)] = res[schserv.RootID(schserv.DBSchema.Name)]
					newTask[schserv.RootID("dest_table")] = res[schserv.RootID("dest_table")]
				} else {
					schema, err := schserv.GetSchemaByID(scheme.GetInt(schserv.RootID(schserv.DBSchema.Name)))
					if err == nil {
						vals, err := s.Domain.SuperCall(utils.AllParams(schema.Name), utils.Record{}, utils.CREATE)
						if err == nil && len(vals) > 0 {
							newTask[schserv.RootID(schserv.DBSchema.Name)] = scheme[schserv.RootID(schserv.DBSchema.Name)]
							newTask[schserv.RootID("dest_table")] = vals[0][utils.ReservedParam]
						}
					}
				}
				if strings.Contains(utils.GetString(res, "nexts"), scheme.GetString("wrapped_" + schserv.RootID(schserv.DBWorkflow.Name))) {
					newMetaRequest := utils.Record{ 
						schserv.RootID(schserv.DBWorkflow.Name) : scheme["wrapped_" + schserv.RootID(schserv.DBWorkflow.Name)], 
						schserv.NAMEKEY : "Meta request for " + newTask.GetString(schserv.NAMEKEY) + " task",
						"current_index" : 1,
						"is_meta": true,
						schserv.RootID(schserv.DBSchema.Name) : newTask[schserv.RootID(schserv.DBSchema.Name)],
						schserv.RootID("dest_table") : newTask[schserv.RootID("dest_table")],
						schserv.RootID(schserv.DBUser.Name) : newTask[schserv.RootID(schserv.DBUser.Name)],
					}
					requests, err := s.Domain.Call(utils.AllParams(schserv.DBRequest.Name), newMetaRequest, utils.CREATE)
					if err == nil && len(requests) > 0 {
						newTask["meta_" + schserv.RootID(schserv.DBRequest.Name)]= requests[0][utils.SpecialIDParam]
					}		
				}
				if utils.GetString(res, "nexts") == "all" || strings.Contains(utils.GetString(res, "nexts"), scheme.GetString("wrapped_" + schserv.RootID(schserv.DBWorkflow.Name))) {
					tasks, err := s.Domain.SuperCall(utils.AllParams(schserv.DBTask.Name), newTask, utils.CREATE)
					if err != nil || len(tasks) == 0 { continue }
					schema, err := schserv.GetSchema(schserv.DBTask.Name)
					if err == nil {
						s.Domain.SuperCall( utils.AllParams(schserv.DBNotification.Name), utils.Record{ "link_id" : schema.ID,
							schserv.NAMEKEY : "Task affected : " + tasks[0].GetString(schserv.NAMEKEY), 
							"description" : "Task is affected : " + tasks[0].GetString(schserv.NAMEKEY),
							schserv.RootID(schserv.DBUser.Name) : tasks[0][schserv.RootID(schserv.DBUser.Name)],
							schserv.RootID(schserv.DBEntity.Name) : scheme[schserv.RootID(schserv.DBEntity.Name)],
							schserv.RootID(schserv.DBUser.Name) : scheme[schserv.RootID(schserv.DBUser.Name)],					
							schserv.RootID("dest_table") : tasks[0][utils.SpecialIDParam], }, utils.CREATE)
					}
				}
			}
	    }
	}
}

func (s *TaskService) WriteRowAutomation(record map[string]interface{}, tableName string) {
	// task creation automation.
	schema, err := schserv.GetSchemaByID(utils.GetInt(record, schserv.RootID(schserv.DBSchema.Name)))
	if err != nil { return }
	created, err := s.Domain.SuperCall(utils.AllParams(schema.Name), utils.Record{}, utils.CREATE)
	if err != nil && len(created) == 0 { return }
	newRec := utils.Record{ schserv.RootID("dest_table"): created[0][utils.SpecialIDParam] }
	s.Domain.SuperCall(utils.AllParams(s.Entity().GetName()), newRec, utils.UPDATE)
}

func (s *TaskService) ConfigureFilter(tableName string) (string, string, string, string) {
	rows, ok := s.Domain.GetParams()[utils.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[utils.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != utils.ReservedParam) || (ok2 && ids != "") {
		return s.Domain.ViewDefinition(tableName)
	}
	restr := schserv.RootID(schserv.DBUser.Name) + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ") OR "
	restr += schserv.RootID(schserv.DBRequest.Name) + " IN (SELECT id FROM " + schserv.DBRequest.Name + " WHERE "
	restr += schserv.RootID(schserv.DBWorkflowSchema.Name) + " IN (SELECT id FROM " + schserv.DBWorkflowSchema.Name + " WHERE "
	restr += schserv.RootID(schserv.DBUser.Name) + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")" 
	restr += " OR " + schserv.RootID(schserv.DBEntity.Name) + " IN ("
	restr += "SELECT " + schserv.RootID(schserv.DBEntity.Name) + " FROM " + schserv.DBEntityUser.Name + " "
	restr += " WHERE " + schserv.RootID(schserv.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + "))))"
	restr += " AND meta_" + schserv.RootID(schserv.DBRequest.Name) + " IS NULL"
	return s.Domain.ViewDefinition(tableName, restr)
}	
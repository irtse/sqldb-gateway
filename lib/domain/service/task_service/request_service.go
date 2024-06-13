package task_service

import (
	"fmt"
	"time"
	"errors"
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type RequestService struct { 
	utils.AbstractSpecializedService
}

func (s *RequestService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results { 
	return s.Domain.PostTreat(results, tableName, true) 
}
func (s *RequestService) ConfigureFilter(tableName string, innerestr... string) (string, string, string, string) { 
	restr := ""
	if s.Domain.IsSuperAdmin() { return s.Domain.ViewDefinition(tableName, restr) }
	restr += schserv.RootID(schserv.DBUser.Name) + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")"
	restr += " OR " + schserv.RootID(schserv.DBUser.Name) + " IN (SELECT " + schserv.RootID(schserv.DBUser.Name) + " FROM " + schserv.DBHierarchy.Name + " WHERE parent_" + schserv.RootID(schserv.DBUser.Name) + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + "))"
	innerestr = append(innerestr, restr)
	return s.Domain.ViewDefinition(tableName, innerestr...) 
}

func (s *RequestService) GetHierarchical() (utils.Results, error) {
	paramsNew := utils.Params{ utils.RootTableParam : schserv.DBHierarchy.Name, utils.RootRowsParam: utils.ReservedParam }
	sqlFilter := "id IN (SELECT id FROM " + schserv.DBHierarchy.Name + " WHERE "
	sqlFilter += schserv.RootID(schserv.DBUser.Name) + " IN ("
	sqlFilter += "SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")"
	sqlFilter += " OR " + schserv.DBEntity.Name + "_id IN ("
	sqlFilter += "SELECT " + schserv.DBEntity.Name + "_id FROM "
	sqlFilter += schserv.DBEntityUser.Name + " WHERE " + schserv.DBUser.Name +"_id IN ("
	sqlFilter += "SELECT id FROM " + schserv.DBUser.Name + " WHERE "
	sqlFilter += "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")))"
	return s.Domain.SuperCall( paramsNew, utils.Record{}, utils.SELECT, sqlFilter )
}

func (s *RequestService) Entity() utils.SpecializedServiceInfo { return schserv.DBRequest }
func (s *RequestService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { }
func (s *RequestService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) { 
	if s.Domain.GetMethod() == utils.CREATE { 
		// set up name
		if _, ok := record[utils.RootDestTableIDParam]; !ok { return record, errors.New("Missing related data"), false }
		sqlFilter := "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser())
		user, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBUser.Name + " WHERE " + sqlFilter)
		if err != nil || len(user) == 0 { return record, errors.New("User not found"), true }
		record[schserv.RootID(schserv.DBUser.Name)]=user[0][utils.SpecialIDParam] 
		hierarchy, err := s.GetHierarchical()
		if err != nil || len(hierarchy) > 0 { record["current_index"]=0 } else { record["current_index"]=1 }
		wf, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBWorkflow.Name + " WHERE id=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBWorkflow.Name)]))
		if err != nil || len(wf) == 0 { return record, errors.New("Workflow not found"), false }
		record["name"]=wf[0][schserv.NAMEKEY]
		record["created_date"] = time.Now().Format(time.RFC3339)
		record[schserv.RootID(schserv.DBSchema.Name)]=wf[0][schserv.RootID(schserv.DBSchema.Name)]
	} else if s.Domain.GetMethod() == utils.UPDATE {
		if record["state"] == "completed" || record["state"] == "dismiss" { 
			record["is_close"] = true 
			record["closing_date"] = time.Now().Format(time.RFC3339)
		} else if !s.Domain.IsSuperCall() { return record, errors.New("Can't set any state lower"), false }
	}
	if s.Domain.GetMethod() != utils.DELETE {
		rec, err := s.Domain.ValidateBySchema(record, tablename)
		if err != nil && !s.Domain.GetAutoload() { return rec, err, false } else { rec = record }
		return rec, nil, true
	}
	return record, nil, true
}
func (s *RequestService) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {
	for _, rec := range results {
		if rec["state"] == "dismiss" { 
			schema, err := schserv.GetSchema(schserv.DBRequest.Name)
			if err == nil && !rec["is_meta"].(bool) { 
				p := utils.AllParams(schserv.DBNotification.Name)
				p[schserv.NAMEKEY] = "Rejected " + utils.GetString(rec, schserv.NAMEKEY)
				p[schserv.RootID(schserv.DBUser.Name)] = fmt.Sprintf("%v", rec[schserv.RootID(schserv.DBUser.Name)])
				p[schserv.RootID("dest_table")] = fmt.Sprintf("%v", rec[utils.SpecialIDParam])
				t, _ := s.Domain.SuperCall( p, utils.Record{}, utils.SELECT)
				if len(t) > 0 { return }
				s.Domain.SuperCall( utils.AllParams(schserv.DBNotification.Name), utils.Record{ "link_id" : schema.ID,
					schserv.NAMEKEY : "Rejected " + utils.GetString(rec, schserv.NAMEKEY), 
					"description" : utils.GetString(rec, schserv.NAMEKEY) + " is rejected and closed.",
					schserv.RootID(schserv.DBUser.Name) : rec[schserv.RootID(schserv.DBUser.Name)],
					schserv.RootID("dest_table") : rec[utils.SpecialIDParam], }, utils.CREATE)
			}
		}
		if rec["state"] == "completed" {
			schema, err := schserv.GetSchema(schserv.DBRequest.Name)
			if err == nil && !rec["is_meta"].(bool) {
				p := utils.AllParams(schserv.DBNotification.Name)
				p[schserv.NAMEKEY] = "Validated " + utils.GetString(rec, schserv.NAMEKEY)
				p[schserv.RootID(schserv.DBUser.Name)] = fmt.Sprintf("%v", rec[schserv.RootID(schserv.DBUser.Name)])
				p[schserv.RootID("dest_table")] = fmt.Sprintf("%v", rec[utils.SpecialIDParam])
				t, _ := s.Domain.SuperCall( p, utils.Record{}, utils.SELECT)
				if len(t) > 0 { return }
				s.Domain.SuperCall( utils.AllParams(schserv.DBNotification.Name), utils.Record{ "link_id" : schema.ID,
					schserv.NAMEKEY : "Validated " + utils.GetString(rec, schserv.NAMEKEY), 
					"description" : utils.GetString(rec, schserv.NAMEKEY) + " is approved and closed.",
					schserv.RootID(schserv.DBUser.Name) : rec[schserv.RootID(schserv.DBUser.Name)],
					schserv.RootID("dest_table") : rec[utils.SpecialIDParam], }, utils.CREATE)
			}
		}
		if rec["is_close"].(bool) {
			res, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBTask.Name + " WHERE meta_" + schserv.RootID(schserv.DBRequest.Name) + "=" + utils.GetString(rec, utils.SpecialIDParam) + " AND is_close=false")
			if err == nil && len(res) > 0 {
				for _, task := range res {
					task["is_close"] = true 
					task["state"] = rec["state"] 
					task["closing_date"] = time.Now().Format(time.RFC3339)
					s.Domain.SuperCall( utils.AllParams(schserv.DBTask.Name), task, utils.UPDATE)
				}
			}
		}
	}
}

func (s *RequestService) WriteRowAutomation(record map[string]interface{}, tableName string) {
	if record["current_index"].(int64) == 1 {
		wfs, err := s.Domain.SuperCall( utils.AllParams(schserv.DBWorkflowSchema.Name), utils.Record{}, 
			utils.SELECT, "index=1 AND " + schserv.RootID(schserv.DBWorkflow.Name) + "=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBWorkflow.Name)]))
		if err != nil || len(wfs) == 0 { 
			params := utils.Params{ utils.RootTableParam : schserv.DBRequest.Name, utils.RootRowsParam : utils.GetString(record, utils.SpecialIDParam), }
			s.Domain.SuperCall( params, utils.Record{ }, utils.DELETE); return 
		}
		for _, newTask := range wfs {
			newTask[schserv.RootID(schserv.DBWorkflowSchema.Name)] = newTask[utils.SpecialIDParam]
			delete(newTask, utils.SpecialIDParam)
			newTask[schserv.RootID(schserv.DBRequest.Name)] = record[utils.SpecialIDParam]
			if newTask.GetString(schserv.RootID(schserv.DBSchema.Name)) == utils.GetString(record, schserv.RootID(schserv.DBSchema.Name)) {
				newTask[schserv.RootID(schserv.DBSchema.Name)] = record[schserv.RootID(schserv.DBSchema.Name)]
				newTask[schserv.RootID("dest_table")] = record[schserv.RootID("dest_table")]
			} else {
				schema, err := schserv.GetSchemaByID(newTask.GetInt(schserv.RootID(schserv.DBSchema.Name)))
				if err == nil {
					vals, err := s.Domain.SuperCall( utils.Params{ utils.RootTableParam: schema.Name, utils.RootRowsParam: utils.ReservedParam, }, 
													 utils.Record{}, utils.CREATE)
					if err == nil && len(vals) > 0 { newTask[schserv.RootID("dest_table")] = vals[0][utils.ReservedParam] }
				}
			}
			tasks, err := s.Domain.SuperCall( utils.AllParams(schserv.DBTask.Name), newTask, utils.CREATE)
			if err == nil {
				task := utils.Record{ 
					schserv.NAMEKEY : "Task affected : " + newTask.GetString(schserv.NAMEKEY), 
					"description" : "Task is affected : " + newTask.GetString(schserv.NAMEKEY),
					schserv.RootID(schserv.DBUser.Name) : newTask.GetString(schserv.RootID(schserv.DBUser.Name)),
					schserv.RootID(schserv.DBSchema.Name) : newTask.GetString(schserv.RootID(schserv.DBSchema.Name)),
					schserv.RootID("dest_table") : record[schserv.RootID("dest_table")], }
				if _, ok := newTask["wrapped_" + schserv.RootID(schserv.DBWorkflow.Name)]; ok { task["is_meta"]= true }
				s.Domain.SuperCall( utils.AllParams(schserv.DBTask.Name), task, utils.CREATE)
				if id, ok := newTask["wrapped_" + schserv.RootID(schserv.DBWorkflow.Name)]; ok {
					newMetaRequest := utils.Record{ 
						schserv.RootID(schserv.DBWorkflow.Name) : id, 
						schserv.NAMEKEY : "Meta request for " + task.GetString(schserv.NAMEKEY) + " task.",
						"current_index" : 1, "is_meta": true,
						schserv.RootID(schserv.DBSchema.Name) : task[schserv.RootID(schserv.DBSchema.Name)],
						schserv.RootID("dest_table") : task[schserv.RootID("dest_table")],
						schserv.RootID(schserv.DBUser.Name) : task[schserv.RootID(schserv.DBUser.Name)],
					}
					s.Domain.SuperCall( utils.AllParams(schserv.DBRequest.Name), newMetaRequest, utils.CREATE)
				}
			}
			if err == nil && len(tasks) > 0 {
				schema, err := schserv.GetSchema(schserv.DBTask.Name)
				if err == nil {
					s.Domain.SuperCall( utils.AllParams(schserv.DBNotification.Name), utils.Record{ "link_id" : schema.ID,
						schserv.NAMEKEY : "Task is affected : " + tasks[0].GetString(schserv.NAMEKEY), 
						"description" : "Task is affected : " + tasks[0].GetString(schserv.NAMEKEY),
						schserv.RootID(schserv.DBEntity.Name) : newTask[schserv.RootID(schserv.DBEntity.Name)],
						schserv.RootID(schserv.DBUser.Name) : newTask[schserv.RootID(schserv.DBUser.Name)],
						schserv.RootID("dest_table") : tasks[0][utils.SpecialIDParam], }, utils.CREATE)
				}
			}
		}
	} else {
		hierarchy, _ := s.GetHierarchical()
		user, err := s.Domain.SuperCall( utils.AllParams(schserv.DBUser.Name), utils.Record{}, utils.SELECT,
			"name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()))
		if err == nil && len(user) > 0 {
			for _, hierarch := range hierarchy {
				newTask := utils.Record{
					schserv.RootID(schserv.DBSchema.Name) : record[schserv.RootID(schserv.DBSchema.Name)],
					schserv.RootID("dest_table") : record[schserv.RootID("dest_table")],
					schserv.RootID(schserv.DBRequest.Name) : record[utils.SpecialIDParam],
					schserv.RootID(schserv.DBUser.Name) : user[0][utils.SpecialIDParam],
					schserv.RootID(schserv.DBUser.Name) : hierarch["parent_" + schserv.RootID(schserv.DBUser.Name)],
					"description" : "hierarchical verification expected by the system.",
					"urgency" : "normal",
					"priority" : "normal",
					schserv.NAMEKEY : "hierarchical verification",
				}
				res, err := s.Domain.PermsSuperCall( utils.AllParams(schserv.DBTask.Name), newTask, utils.CREATE)
				if err == nil && len(res) > 0 {
					schema, err := schserv.GetSchema(schserv.DBTask.Name)
					if err == nil {
						s.Domain.SuperCall( utils.AllParams(schserv.DBNotification.Name), utils.Record{ 
							schserv.NAMEKEY : "Hierarchical verification on " + utils.GetString(record, schserv.NAMEKEY) + " request", 
							"description" : utils.GetString(record, schserv.NAMEKEY) + " request need a hierarchical verification.",
							schserv.RootID(schserv.DBUser.Name) : hierarch["parent_" + schserv.RootID(schserv.DBUser.Name)],
							"link_id" : schema.ID, schserv.RootID("dest_table") : res[0][utils.SpecialIDParam], }, utils.CREATE)
					}
				}
			}	
		}
	}
}

package task_service

import (
	"errors"
	"sqldb-ws/domain/filter"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	"sqldb-ws/infrastructure/connector"
	conn "sqldb-ws/infrastructure/connector"
	"time"
)

type RequestService struct {
	servutils.AbstractSpecializedService
}

func (s *RequestService) ShouldVerify() bool { return true }
func (s *RequestService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true)
}
func (s *RequestService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	n := []string{}
	f := filter.NewFilterService(s.Domain)
	if !s.Domain.IsSuperCall() {
		n = append(n, "("+connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			ds.UserDBField: f.GetUserFilterQuery("id"),
			ds.UserDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
				"parent_" + ds.UserDBField: f.GetUserFilterQuery("id"),
			}, false),
		}, true)+")")
	}
	n = append(n, innerestr...)
	return f.GetQueryFilter(tableName, s.Domain.GetParams().Copy(), n...)
}

func (s *RequestService) GetHierarchical() ([]map[string]interface{}, error) {
	f := filter.NewFilterService(s.Domain)
	return s.Domain.GetDb().SelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
		ds.UserDBField:   f.GetUserFilterQuery("id"),
		ds.EntityDBField: f.GetEntityFilterQuery("id"),
	}, false)
}

func (s *RequestService) Entity() utils.SpecializedServiceInfo                                    { return ds.DBRequest }
func (s *RequestService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {}
func (s *RequestService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.CREATE {
		if _, ok := record[utils.RootDestTableIDParam]; !ok {
			return record, errors.New("missing related data"), false
		}
		f := filter.NewFilterService(s.Domain)
		if user, err := s.Domain.GetDb().QueryAssociativeArray(f.GetUserFilterQuery("*")); err != nil || len(user) == 0 {
			return record, errors.New("user not found"), true
		} else {
			record[ds.UserDBField] = user[0][utils.SpecialIDParam]
			if hierarchy, err := s.GetHierarchical(); err != nil || len(hierarchy) > 0 {
				record["current_index"] = 0
			} else {
				record["current_index"] = 1
			}
			if wf, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
				utils.SpecialIDParam: record[ds.WorkflowDBField],
			}, false); err != nil || len(wf) == 0 {
				return record, errors.New("workflow not found"), false
			} else {
				record["name"] = wf[0][sm.NAMEKEY]
				record["created_date"] = time.Now().Format(time.RFC3339)
				record[ds.SchemaDBField] = wf[0][ds.SchemaDBField]
			}
		}

	} else if s.Domain.GetMethod() == utils.UPDATE {
		SetClosureStatus(record)
	}
	if s.Domain.GetMethod() != utils.DELETE {
		servutils.CheckAutoLoad(tablename, record, s.Domain)
	}
	return record, nil, true
}
func (s *RequestService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	for _, rec := range results {
		p := utils.AllParams(ds.DBNotification.Name)
		p[ds.UserDBField] = utils.ToString(rec[ds.UserDBField])
		p[ds.DestTableDBField] = utils.ToString(rec[utils.SpecialIDParam])
		switch rec["state"] {
		case "dismiss":
			p[sm.NAMEKEY] = "Rejected " + utils.GetString(rec, sm.NAMEKEY)
			p["description"] = utils.GetString(rec, sm.NAMEKEY) + " is accepted and closed."
		case "completed":
			p[sm.NAMEKEY] = "Validated " + utils.GetString(rec, sm.NAMEKEY)
			p["description"] = utils.GetString(rec, sm.NAMEKEY) + " is accepted and closed."
		}
		schema, err := schserv.GetSchema(ds.DBRequest.Name)
		if err == nil && !utils.Compare(rec["is_meta"], true) && CheckStateIsEnded(rec["state"]) {
			if t, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT, false); err == nil && len(t) > 0 {
				return
			}
			delete(p, utils.RootTableParam)
			delete(p, utils.RootRowsParam)
			rec := p.Anonymized()
			rec["link_id"] = schema.ID
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), rec)
		}
		if utils.Compare(rec["is_close"], true) {
			p := utils.AllParams(ds.DBTask.Name)
			p["meta_"+ds.RequestDBField] = utils.ToString(rec[utils.SpecialIDParam])
			res, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT, false)
			if err == nil && len(res) > 0 {
				for _, task := range res {
					task := SetClosureStatus(task)
					s.Domain.UpdateSuperCall(utils.AllParams(ds.DBTask.Name), task)
				}
			}
		}
	}
}

func (s *RequestService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if utils.GetInt(record, "current_index") == 1 {
		s.handleInitialWorkflow(record)
	} else {
		s.handleHierarchicalVerification(record)
	}
}

func (s *RequestService) handleInitialWorkflow(record map[string]interface{}) {
	wfs, err := s.Domain.SuperCall(utils.AllParams(ds.DBWorkflowSchema.Name), utils.Record{},
		utils.SELECT, false, "index=1 AND "+ds.WorkflowDBField+"="+utils.ToString(record[ds.WorkflowDBField]))
	if err != nil || len(wfs) == 0 {
		params := utils.Params{utils.RootTableParam: ds.DBRequest.Name, utils.RootRowsParam: utils.GetString(record, utils.SpecialIDParam)}
		s.Domain.DeleteSuperCall(params)
		return
	}

	for _, newTask := range wfs {
		s.prepareAndCreateTask(newTask, record)
	}
}

func (s *RequestService) prepareAndCreateTask(newTask utils.Record, record map[string]interface{}) {
	newTask[ds.WorkflowSchemaDBField] = newTask[utils.SpecialIDParam]
	delete(newTask, utils.SpecialIDParam)
	newTask[ds.RequestDBField] = record[utils.SpecialIDParam]

	if newTask.GetString(ds.SchemaDBField) == utils.GetString(record, ds.SchemaDBField) {
		newTask[ds.SchemaDBField] = record[ds.SchemaDBField]
		newTask[ds.DestTableDBField] = record[ds.DestTableDBField]
	} else if schema, err := schserv.GetSchemaByID(newTask.GetInt(ds.SchemaDBField)); err == nil {
		if vals, err := s.Domain.CreateSuperCall(utils.AllParams(schema.Name), utils.Record{}); err == nil && len(vals) > 0 {
			newTask[ds.DestTableDBField] = vals[0][utils.ReservedParam]
		}
	}

	s.createTaskAndNotify(newTask, record)
}

func (s *RequestService) createTaskAndNotify(newTask, record map[string]interface{}) {
	tasks, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name), newTask)
	if err != nil || len(tasks) == 0 {
		return
	}
	task := s.constructNotificationTask(newTask, record)
	s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name), task)

	if id, ok := newTask["wrapped_"+ds.WorkflowDBField]; ok {
		s.createMetaRequest(task, id)
	}

	if schema, err := schserv.GetSchema(ds.DBTask.Name); len(tasks) > 0 && err == nil {
		task[ds.DestTableDBField] = tasks[0][utils.SpecialIDParam]
		task["link_id"] = schema.ID
		s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), task)
	}
}

func (s *RequestService) constructNotificationTask(newTask utils.Record, record map[string]interface{}) map[string]interface{} {
	task := map[string]interface{}{
		sm.NAMEKEY:          "Task affected : " + newTask.GetString(sm.NAMEKEY),
		"description":       "Task is affected : " + newTask.GetString(sm.NAMEKEY),
		ds.UserDBField:      newTask.GetString(ds.UserDBField),
		ds.SchemaDBField:    newTask.GetString(ds.SchemaDBField),
		ds.DestTableDBField: record[ds.DestTableDBField],
	}
	if _, ok := newTask["wrapped_"+ds.WorkflowDBField]; ok {
		task["is_meta"] = true
	}
	return task
}

func (s *RequestService) createMetaRequest(task map[string]interface{}, id interface{}) {
	s.Domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name), utils.Record{
		ds.WorkflowDBField:  id,
		sm.NAMEKEY:          "Meta request for " + utils.GetString(task, sm.NAMEKEY) + " task.",
		"current_index":     1,
		"is_meta":           true,
		ds.SchemaDBField:    task[ds.SchemaDBField],
		ds.DestTableDBField: task[ds.DestTableDBField],
		ds.UserDBField:      task[ds.UserDBField],
	})
}

func (s *RequestService) handleHierarchicalVerification(record map[string]interface{}) {
	user, err2 := s.Domain.SuperCall(utils.AllParams(ds.DBUser.Name), utils.Record{},
		utils.SELECT, false, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			sm.NAMEKEY: conn.Quote(s.Domain.GetUser()),
			"email":    conn.Quote(s.Domain.GetUser()),
		}, true))

	if hierarchy, err := s.GetHierarchical(); err == nil && err2 == nil && len(user) > 0 {
		for _, hierarch := range hierarchy {
			s.createHierarchicalTask(record, user[0], hierarch)
		}
	}
}

func (s *RequestService) createHierarchicalTask(record, user, hierarch map[string]interface{}) {
	newTask := utils.Record{
		ds.SchemaDBField:           record[ds.SchemaDBField],
		ds.DestTableDBField:        record[ds.DestTableDBField],
		ds.RequestDBField:          record[utils.SpecialIDParam],
		ds.UserDBField:             user[utils.SpecialIDParam],
		"parent_" + ds.UserDBField: hierarch["parent_"+ds.UserDBField],
		"description":              "hierarchical verification expected by the system.",
		"urgency":                  "normal",
		"priority":                 "normal",
		sm.NAMEKEY:                 "hierarchical verification",
	}
	if res, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name).RootRaw(), newTask); err == nil && len(res) > 0 {
		if schema, err := schserv.GetSchema(ds.DBTask.Name); err == nil {
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), utils.Record{
				sm.NAMEKEY:          "Hierarchical verification on " + utils.GetString(record, sm.NAMEKEY) + " request",
				"description":       utils.GetString(record, sm.NAMEKEY) + " request needs a hierarchical verification.",
				ds.UserDBField:      hierarch["parent_"+ds.UserDBField],
				"link_id":           schema.ID,
				ds.DestTableDBField: res[0][utils.SpecialIDParam],
			})
		}
	}
}

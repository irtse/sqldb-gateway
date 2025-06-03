package task_service

import (
	"errors"
	"fmt"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"time"
)

type RequestService struct {
	servutils.AbstractSpecializedService
}

func (s *RequestService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	// TODO: here send back my passive task...
	res := view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
	if len(results) == 1 && s.Domain.GetMethod() == utils.CREATE {
		fmt.Println("THERE", results)
		// retrieve... tasks affected to you
		if r, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.RequestDBField: results[0][utils.SpecialIDParam],
			utils.SpecialIDParam: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
				ds.EntityDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBEntityUser.Name, map[string]interface{}{
					ds.UserDBField: s.Domain.GetUserID(),
				}, false, ds.EntityDBField),
			}, true, utils.SpecialIDParam),
		}, false); err == nil && len(r) > 0 {
			if sch, err := schema.GetSchema(ds.DBTask.Name); err == nil {
				res[0]["inner_redirection"] = utils.BuildPath(sch.ID, utils.GetString(r[0], utils.SpecialIDParam))
			}
		} else if sch, err := schema.GetSchemaByID(utils.GetInt(results[0], ds.SchemaDBField)); err == nil {
			res[0]["inner_redirection"] = utils.BuildPath(sch.ID, utils.GetString(results[0], ds.DestTableDBField))
		}
	}
	return res
}
func (s *RequestService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	n := []string{}
	f := filter.NewFilterService(s.Domain)
	if !s.Domain.IsSuperCall() {
		n = append(n, "("+connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			ds.UserDBField: s.Domain.GetUserID(),
			ds.UserDBField + "_1": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
				"parent_" + ds.UserDBField: s.Domain.GetUserID(),
			}, true, ds.UserDBField),
		}, true)+")")
	}
	n = append(n, innerestr...)
	return f.GetQueryFilter(tableName, s.Domain.GetParams().Copy(), n...)
}

func GetHierarchical(domain utils.DomainITF) ([]map[string]interface{}, error) {
	f := filter.NewFilterService(domain)
	return domain.GetDb().SelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
		ds.UserDBField:   domain.GetUserID(),
		ds.EntityDBField: f.GetEntityFilterQuery(),
	}, true)
}

func (s *RequestService) Entity() utils.SpecializedServiceInfo                                    { return ds.DBRequest }
func (s *RequestService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {}
func (s *RequestService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.CREATE {
		if _, ok := record[utils.RootDestTableIDParam]; !ok {
			return record, errors.New("missing related data"), false
		}
		record[ds.UserDBField] = s.Domain.GetUserID()
		if hierarchy, err := GetHierarchical(s.Domain); err != nil || len(hierarchy) > 0 {
			record["current_index"] = 0
			if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
				"index":                          1,
				"before_hierarchical_validation": true,
				ds.WorkflowDBField:               record[ds.WorkflowDBField],
			}, false); err == nil && len(res) == 0 {
				record["current_index"] = 1
			}
		} else {
			record["current_index"] = 1
		}
		if wf, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
			utils.SpecialIDParam: record[ds.WorkflowDBField],
		}, false); err != nil || len(wf) == 0 {
			return record, nil, true
		} else {
			record["name"] = wf[0][sm.NAMEKEY]
			record[ds.SchemaDBField] = wf[0][ds.SchemaDBField]
			record["send_mail_to"] = wf[0]["send_mail_to"]
		}

	} else if s.Domain.GetMethod() == utils.UPDATE {
		record = SetClosureStatus(record)
	}
	if s.Domain.GetMethod() != utils.DELETE {
		if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
			return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
		} else {
			return record, err, false
		}
	}
	return record, nil, true
}
func (s *RequestService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	if _, ok := record["is_draft"]; ok && utils.GetBool(record, "is_draft") {
		return
	}
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
	for _, rec := range results {
		p := utils.AllParams(ds.DBNotification.Name)
		p.Set(ds.UserDBField, utils.ToString(rec[ds.UserDBField]))
		p.Set(ds.DestTableDBField, utils.ToString(rec[utils.SpecialIDParam]))
		switch rec["state"] {
		case "dismiss":
		case "refused":
			rec["state"] = "refused"
			p.Set(sm.NAMEKEY, "Rejected "+utils.GetString(rec, sm.NAMEKEY))
			p.Set("description", utils.GetString(rec, sm.NAMEKEY)+" is rejected and closed.")
			task.SetEndedRequest(utils.GetString(rec, ds.SchemaDBField), utils.GetString(rec, ds.DestTableDBField),
				utils.GetString(rec, utils.SpecialIDParam), s.Domain.GetDb())
		case "completed":
			p.Set(sm.NAMEKEY, "Validated "+utils.GetString(rec, sm.NAMEKEY))
			p.Set("description", utils.GetString(rec, sm.NAMEKEY)+" is accepted and closed.")
			task.SetEndedRequest(utils.GetString(rec, ds.SchemaDBField), utils.GetString(rec, ds.DestTableDBField),
				utils.GetString(rec, utils.SpecialIDParam), s.Domain.GetDb())
		}
		schema, err := schserv.GetSchema(ds.DBRequest.Name)
		if err == nil && !utils.Compare(rec["is_meta"], true) && CheckStateIsEnded(rec["state"]) {
			if t, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT, false); err == nil && len(t) > 0 {
				return
			}
			p.SimpleDelete(utils.RootTableParam)
			p.SimpleDelete(utils.RootRowsParam)
			rec := p.Anonymized()
			rec["link_id"] = schema.ID
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), rec)
		}
		if utils.Compare(rec["is_close"], true) {
			p := utils.AllParams(ds.DBTask.Name)
			p.Set("meta_"+ds.RequestDBField, utils.ToString(rec[utils.SpecialIDParam]))
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

// vérifier qu'il n'existe pas déjà une request méta en cour... si oui... faire une tache méta dans la nouvelle request

func (s *RequestService) Write(record utils.Record, tableName string) {
	if _, ok := record["is_draft"]; ok && utils.GetBool(record, "is_draft") {
		return
	}
	if utils.GetInt(record, "current_index") == 0 {
		found := false
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
			"index":            1,
			ds.WorkflowDBField: record[ds.WorkflowDBField],
		}, false); err == nil {
			for _, rec := range res {
				if utils.GetBool(rec, "before_hierarchical_validation") {
					found = true
					break
				}
			}
		}
		if found {
			record["current_index"] = 0.9
			record = HandleHierarchicalVerification(s.Domain, utils.GetInt(record, utils.SpecialIDParam), record)
		} else {
			record["current_index"] = 1
		}
	}
	if utils.GetInt(record, "current_index") == 1 {
		s.handleInitialWorkflow(record)
	}
}

func (s *RequestService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.Write(record, tableName)
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}

func (s *RequestService) handleInitialWorkflow(record map[string]interface{}) {
	wfs, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
		"index":            1,
		ds.WorkflowDBField: record[ds.WorkflowDBField],
	}, false)
	if err != nil || len(wfs) == 0 {
		s.Domain.GetDb().DeleteQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
			utils.SpecialIDParam: utils.GetString(record, utils.SpecialIDParam),
		}, false)
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
		// THERE o_o
		r := utils.Record{"is_draft": true}
		if schema.HasField("name") {
			if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err == nil {
				if res, err := s.Domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
					utils.SpecialIDParam: record[ds.DestTableDBField],
				}, false); err == nil && len(res) > 0 {
					fmt.Println(utils.GetString(res[0], "name"))
					r[sm.NAMEKEY] = "<" + utils.GetString(res[0], "name") + "> " + utils.GetString(newTask, sm.NAMEKEY)
				}
			} else {
				r["name"] = utils.GetString(newTask, "name")
			}
		}
		if schema.HasField(ds.DestTableDBField) && schema.HasField(ds.SchemaDBField) {
			// get workflow source schema + dest ID
			r[ds.DestTableDBField] = record[ds.DestTableDBField]
			r[ds.SchemaDBField] = record[ds.SchemaDBField]
		}
		if schema.HasField(ds.UserDBField) {
			r[ds.UserDBField] = record[ds.UserDBField]
		}
		if schema.HasField(ds.EntityDBField) {
			r[ds.EntityDBField] = record[ds.EntityDBField]
		}
		for _, f := range schema.Fields {
			if f.GetLink() == record[ds.SchemaDBField] {
				r[f.Name] = record[ds.DestTableDBField]
			}
		}

		if i, err := s.Domain.GetDb().CreateQuery(schema.Name, r, func(s string) (string, bool) {
			return "", true
		}); err == nil {
			newTask[ds.DestTableDBField] = i
			s.Domain.GetDb().CreateQuery(ds.DBDataAccess.Name, map[string]interface{}{
				ds.SchemaDBField:    schema.ID,
				ds.DestTableDBField: i,
				ds.UserDBField:      s.Domain.GetUserID(),
				"write":             true,
				"update":            false,
			}, func(s string) (string, bool) {
				return "", true
			})
		}
	}
	if utils.GetBool(newTask, "assign_to_creator") {
		newTask[ds.UserDBField] = s.Domain.GetUserID()
	}
	s.createTaskAndNotify(newTask, record)
}

func (s *RequestService) createTaskAndNotify(newTask map[string]interface{}, request utils.Record) {
	task := s.constructNotificationTask(newTask, request)
	i, err := s.Domain.GetDb().CreateQuery(ds.DBTask.Name, task, func(s string) (string, bool) {
		return "", true
	})
	if err != nil {
		fmt.Println(i, err)
		return
	}
	currentTime := time.Now()
	sqlFilter := []string{
		"('" + currentTime.Format("2000-01-01") + "' < start_date OR '" + currentTime.Format("2000-01-01") + "' > end_date)",
	}
	sqlFilter = append(sqlFilter, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"all_tasks":    true,
		ds.UserDBField: s.Domain.GetUserID(),
	}, false))
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDelegation.Name, utils.ToListAnonymized(sqlFilter), false); err == nil && len(res) > 0 {
		tmpUser := utils.GetInt(task, ds.UserDBField)
		for _, delegated := range res {
			task["binded_dbtask"] = i
			task[ds.UserDBField] = delegated["delegated_"+ds.UserDBField]
			if y, err := s.Domain.GetDb().CreateQuery(ds.DBTask.Name, task, func(s string) (string, bool) {
				return "", true
			}); err == nil {
				notify(task, y, s.Domain)
			}
		}
		delete(task, "binded_dbtask")
		task[ds.UserDBField] = tmpUser
	}
	if id, ok := newTask["wrapped_"+ds.WorkflowDBField]; ok && id != nil {
		s.createMetaRequest(task, id)
	}
	notify(task, i, s.Domain)
}

func notify(task utils.Record, i int64, domain utils.DomainITF) {
	if schema, err := schserv.GetSchema(ds.DBTask.Name); err == nil {
		notif := utils.Record{
			"name":              utils.GetString(task, "name"),
			"description":       utils.GetString(task, "description"),
			ds.UserDBField:      task[ds.UserDBField],
			ds.EntityDBField:    task[ds.EntityDBField],
			ds.DestTableDBField: i,
		}
		notif["link_id"] = schema.ID
		domain.GetDb().CreateQuery(ds.DBNotification.Name, notif, func(s string) (string, bool) {
			return "", true
		})
	}
}

func (s *RequestService) constructNotificationTask(newTask utils.Record, request utils.Record) map[string]interface{} {
	task := map[string]interface{}{
		sm.NAMEKEY:               newTask.GetString(sm.NAMEKEY),
		"description":            newTask.GetString(sm.NAMEKEY),
		"urgency":                newTask["urgency"],
		"priority":               newTask["priority"],
		ds.WorkflowSchemaDBField: newTask[ds.WorkflowSchemaDBField],
		ds.UserDBField:           newTask[ds.UserDBField],
		ds.EntityDBField:         newTask[ds.EntityDBField],
		ds.SchemaDBField:         newTask[ds.SchemaDBField],
		ds.DestTableDBField:      newTask[ds.DestTableDBField],
		ds.RequestDBField:        newTask[ds.RequestDBField],
		"send_mail_to":           newTask["send_mail_to"],
	}
	if schema, err := schserv.GetSchemaByID(request.GetInt(ds.SchemaDBField)); err == nil {
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
			utils.SpecialIDParam: request[ds.DestTableDBField],
		}, false); err == nil && len(res) > 0 {
			fmt.Println(utils.GetString(res[0], "name"))
			task[sm.NAMEKEY] = "<" + utils.GetString(res[0], "name") + "> " + utils.GetString(task, sm.NAMEKEY)
		}
	}
	return task
}

func (s *RequestService) createMetaRequest(task map[string]interface{}, id interface{}) {
	s.Domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name).RootRaw(), utils.Record{
		ds.WorkflowDBField:  id,
		sm.NAMEKEY:          "Meta request for " + utils.GetString(task, sm.NAMEKEY) + " task.",
		"current_index":     1,
		"is_meta":           true,
		ds.SchemaDBField:    task[ds.SchemaDBField],
		ds.DestTableDBField: task[ds.DestTableDBField],
		ds.UserDBField:      utils.GetInt(task, ds.UserDBField),
	})
}

func HandleHierarchicalVerification(domain utils.DomainITF, requestID int64, record map[string]interface{}) map[string]interface{} {
	if hierarchy, err := GetHierarchical(domain); err == nil {
		for _, hierarch := range hierarchy {
			CreateHierarchicalTask(domain, requestID, record, hierarch)
		}
	}
	return record
}

func CreateHierarchicalTask(domain utils.DomainITF, requestID int64, record, hierarch map[string]interface{}) {
	newTask := utils.Record{
		ds.SchemaDBField:    record[ds.SchemaDBField],
		ds.DestTableDBField: record[ds.DestTableDBField],
		ds.RequestDBField:   requestID,
		ds.UserDBField:      hierarch["parent_"+ds.UserDBField],
		"description":       "hierarchical verification expected by the system.",
		"urgency":           "normal",
		"priority":          "normal",
		sm.NAMEKEY:          "hierarchical verification",
	}
	if i, err := domain.GetDb().CreateQuery(ds.DBTask.Name, newTask, func(s string) (string, bool) {
		return "", true
	}); err == nil {
		currentTime := time.Now()
		sqlFilter := []string{
			"('" + currentTime.Format("2000-01-01") + "' < start_date OR '" + currentTime.Format("2000-01-01") + "' > end_date)",
		}
		sqlFilter = append(sqlFilter, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"all_tasks":    true,
			ds.UserDBField: domain.GetUserID(),
		}, false))
		if res, err := domain.GetDb().SelectQueryWithRestriction(ds.DBDelegation.Name, sqlFilter, false); err == nil && len(res) > 0 {
			tmpUser := utils.GetInt(newTask, ds.UserDBField)
			for _, delegated := range res {
				newTask["binded_dbtask"] = i
				newTask[ds.UserDBField] = delegated["delegated_"+ds.UserDBField]
				if y, err := domain.GetDb().CreateQuery(ds.DBTask.Name, newTask, func(s string) (string, bool) {
					return "", true
				}); err == nil {
					notify(newTask, y, domain)
				}
			}
			delete(newTask, "binded_dbtask")
			newTask[ds.UserDBField] = tmpUser
		}
		notify(newTask, i, domain)
	}
}

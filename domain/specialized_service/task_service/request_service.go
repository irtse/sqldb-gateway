package task_service

import (
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/domain_service/view_convertor"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

type RequestService struct {
	servutils.AbstractSpecializedService
}

func (s *RequestService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	// TODO: here send back my passive task...
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
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
	for _, rec := range results {
		p := utils.AllParams(ds.DBNotification.Name)
		p.Set(ds.UserDBField, utils.ToString(rec[ds.UserDBField]))
		p.Set(ds.DestTableDBField, utils.ToString(rec[utils.SpecialIDParam]))
		switch rec["state"] {
		case "dismiss":
			p.Set(sm.NAMEKEY, "Dissmissed "+utils.GetString(rec, sm.NAMEKEY))
			p.Set("description", utils.GetString(rec, sm.NAMEKEY)+" is dissmissed and closed.")
			task.SetEndedRequest(utils.GetString(rec, ds.SchemaDBField),
				utils.GetString(rec, ds.DestTableDBField), utils.GetString(rec, utils.SpecialIDParam), s.Domain.GetDb())
		case "refused":
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
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
}

// vérifier qu'il n'existe pas déjà une request méta en cour... si oui... faire une tache méta dans la nouvelle request

func (s *RequestService) Write(record utils.Record, tableName string) {
	if _, ok := record["is_draft"]; ok && utils.GetBool(record, "is_draft") {
		return
	}

	if utils.GetInt(record, "current_index") == 0 {
		found := false
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
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
		params := utils.GetRowTargetParameters(ds.DBRequest.Name, utils.GetString(record, utils.SpecialIDParam))
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
		if i, err := s.Domain.GetDb().CreateQuery(schema.Name, utils.Record{"is_draft": false}, func(s string) (string, bool) {
			return "", true
		}); err == nil {
			newTask[ds.DestTableDBField] = i
		}
	}
	if utils.GetBool(newTask, "assign_to_creator") {
		newTask[ds.UserDBField] = s.Domain.GetUserID()
	}
	s.createTaskAndNotify(newTask, record)
}

func (s *RequestService) createTaskAndNotify(newTask, record map[string]interface{}) {
	task := s.constructNotificationTask(newTask, record)
	i, err := s.Domain.GetDb().CreateQuery(ds.DBTask.Name, task, func(s string) (string, bool) {
		return "", true
	})
	if err != nil {
		return
	}
	if id, ok := newTask["wrapped_"+ds.WorkflowDBField]; ok && id != nil {
		s.createMetaRequest(task, id)
	}

	if schema, err := schserv.GetSchema(ds.DBTask.Name); err == nil {
		delete(task, ds.SchemaDBField)
		task[ds.DestTableDBField] = i
		task["link_id"] = schema.ID
		s.Domain.GetDb().CreateQuery(ds.DBNotification.Name, task, func(s string) (string, bool) {
			return "", true
		})
	}
}

func (s *RequestService) constructNotificationTask(newTask utils.Record, record map[string]interface{}) map[string]interface{} {
	task := map[string]interface{}{
		sm.NAMEKEY:          "Task affected : " + newTask.GetString(sm.NAMEKEY),
		"description":       "Task is affected : " + newTask.GetString(sm.NAMEKEY),
		ds.UserDBField:      newTask[ds.UserDBField],
		ds.SchemaDBField:    newTask[ds.SchemaDBField],
		ds.DestTableDBField: record[ds.DestTableDBField],
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
		if schema, err := schserv.GetSchema(ds.DBTask.Name); err == nil {
			domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), utils.Record{
				sm.NAMEKEY:          "Hierarchical verification on " + utils.GetString(record, sm.NAMEKEY) + " request",
				"description":       utils.GetString(record, sm.NAMEKEY) + " request needs a hierarchical verification.",
				ds.UserDBField:      hierarch["parent_"+ds.UserDBField],
				"link_id":           schema.ID,
				ds.DestTableDBField: i,
			})
		}
	}
}

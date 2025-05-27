package task_service

import (
	"errors"
	"math"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/domain_service/view_convertor"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	conn "sqldb-ws/infrastructure/connector"
	"strings"
	"time"
)

// this cache must be use to ... match things with things exemple : view

// TODO
type TaskService struct {
	servutils.AbstractSpecializedService
}

func (s *TaskService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	task.CreateTask(s.Domain, record)
	currentTime := time.Now()
	sqlFilter := []string{
		"('" + currentTime.Format("2000-01-01") + "' < start_date OR '" + currentTime.Format("2000-01-01") + "' > end_date)",
	}
	sqlFilter = append(sqlFilter, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"all_tasks":    true,
		ds.UserDBField: s.Domain.GetUserID(),
	}, false))
	if dels, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBDelegation.Name, utils.ToListAnonymized(sqlFilter), false); err == nil && len(dels) > 0 {
		tmpUser := utils.GetInt(record, ds.UserDBField)
		for _, delegated := range dels {
			record["binded_dbtask"] = record[utils.SpecialIDParam]
			record[ds.UserDBField] = delegated["delegated_"+ds.UserDBField]
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name), record)
		}
		delete(record, "binded_dbtask")
		record[ds.UserDBField] = tmpUser
	}
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}
func (s *TaskService) Entity() utils.SpecializedServiceInfo { return ds.DBTask }
func (s *TaskService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	// TODO: here send back my passive task...
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}
func (s *TaskService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	switch s.Domain.GetMethod() {
	case utils.CREATE:
		record[ds.DBUser.Name] = s.Domain.GetUserID()
		if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
			return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
		} else {
			return record, err, false
		}
	case utils.UPDATE:
		// check if task is already closed
		if elder, _ := s.Domain.SuperCall(utils.GetRowTargetParameters(ds.DBTask.Name, record[utils.SpecialIDParam]),
			utils.Record{}, utils.SELECT, false); len(elder) > 0 && CheckStateIsEnded(utils.ToString(elder[0]["state"])) {
			return record, errors.New("task is already closed, cannot change its state"), false
		}
		record = SetClosureStatus(record) // check if task is already progressing
		if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
			return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
		} else {
			return record, err, false
		}
	}
	return record, nil, true
}

func (s *TaskService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for i, res := range results {
		task.RemoveTask(res, utils.GetString(res, ds.UserDBField))
		res["state"] = "refused"
		results[i] = SetClosureStatus(res)
	}
	s.Write(results, map[string]interface{}{})
}

func (s *TaskService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	currentTime := time.Now()
	sqlFilter := []string{
		"('" + currentTime.Format("2000-01-01") + "' < start_date OR '" + currentTime.Format("2000-01-01") + "' > end_date)",
	}
	sqlFilter = append(sqlFilter, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"all_tasks":    true,
		ds.UserDBField: s.Domain.GetUserID(),
	}, false))
	if dels, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBDelegation.Name, utils.ToListAnonymized(sqlFilter), false); err == nil && len(dels) > 0 {
		for _, delegated := range dels {
			for _, r := range results {
				tmpUser := utils.GetInt(r, ds.UserDBField)
				tmpID := utils.GetInt(r, utils.SpecialIDParam)
				delete(r, utils.SpecialIDParam)
				r["binded_dbtask"] = tmpID
				r[ds.UserDBField] = delegated["delegated_"+ds.UserDBField]
				s.Domain.UpdateSuperCall(utils.AllParams(ds.DBTask.Name), r)
				delete(r, "binded_dbtask")
				r[utils.SpecialIDParam] = tmpID
				r[ds.UserDBField] = tmpUser
			}
		}
	}
	s.Write(results, record)
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
}

func (s *TaskService) deleteAll(destID string, schID int64) {
	if sch, err := schserv.GetSchemaByID(schID); err == nil {
		s.Domain.GetDb().ClearQueryFilter().DeleteQueryWithRestriction(sch.Name, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, false)
		s.Domain.GetDb().ClearQueryFilter().DeleteQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.DestTableDBField: destID,
			ds.SchemaDBField:    schID,
		}, false)
		if reqs, err := s.Domain.DeleteSuperCall(utils.AllParams(ds.DBRequest.Name).Enrich(map[string]interface{}{
			ds.DestTableDBField: destID,
			ds.SchemaDBField:    schID,
		}), false); err == nil {
			for _, req := range reqs {
				s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), utils.Record{
					"link_id":        nil,
					sm.NAMEKEY:       "Request cancelled : " + utils.GetString(req, "name"),
					"description":    "Request is cancelled : " + utils.GetString(req, "name"),
					UserDBField:      req[UserDBField],
					EntityDBField:    req[EntityDBField],
					DestTableDBField: destID,
				})
			}
		}
	}
}

func (s *TaskService) Write(results []map[string]interface{}, record map[string]interface{}) {
	for _, res := range results {
		if _, ok := res["is_draft"]; ok && utils.GetBool(res, "is_draft") {
			continue
		}

		task.CreateTask(s.Domain, res)
		if binded, ok := res["binded_dbtask"]; ok && utils.GetBool(res, "is_close") && binded != nil {
			s.Domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBTask.Name, map[string]interface{}{
				"is_close":                    res["is_close"],
				"state":                       res["state"],
				"closing_by" + ds.UserDBField: utils.GetInt(res, ds.UserDBField),
				"closing_date":                res["closing_date"],
			}, map[string]interface{}{
				utils.SpecialIDParam: binded,
				utils.SpecialIDParam + "_1": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
					"binded_dbtask":            binded,
					"!" + utils.SpecialIDParam: res[utils.SpecialIDParam],
				}, false, utils.SpecialIDParam),
			}, true)
		}
		requests, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
			utils.SpecialIDParam: utils.GetInt(res, RequestDBField),
		}, false)
		if err != nil || len(requests) == 0 {
			continue
		}
		order := requests[0]["current_index"]
		if otherPendingTasks, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name,
			map[string]interface{}{ // delete all notif
				RequestDBField:  utils.ToString(res[RequestDBField]),
				"state":         []string{"'pending'", "'progressing'"},
				"binded_dbtask": nil,
			}, false); err == nil && len(otherPendingTasks) > 0 {
			continue
		}

		current_index := utils.ToFloat64(order)
		switch res["state"] {
		case "completed":
			current_index = math.Floor(current_index + 1)
		case "refused":
			s.Domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBRequest.Name, utils.Record{"state": "refused"},
				map[string]interface{}{
					utils.SpecialIDParam: utils.GetInt(res, RequestDBField),
				}, false)
		case "dismiss":
			if current_index >= 1 {
				current_index = math.Floor(current_index - 1)
			} else { // Dismiss will close requests.
				s.Domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBRequest.Name, utils.Record{"state": "dismiss"},
					map[string]interface{}{
						utils.SpecialIDParam: utils.GetInt(res, RequestDBField),
					}, false)
			} // no before task close request and task
		}
		schemes, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name,
			map[string]interface{}{
				"index":            current_index,
				ds.WorkflowDBField: requests[0][ds.WorkflowDBField],
			}, false)
		allOptionnal := true
		if err == nil {
			for _, scheme := range schemes {
				if !utils.GetBool(scheme, "optionnal") {
					allOptionnal = false
					break
				}
			}
		}
		newRecRequest := utils.Record{utils.SpecialIDParam: requests[0][utils.SpecialIDParam]}
		if allOptionnal { // no new task in workflow
			newRecRequest["state"] = "completed"
		} else {
			newRecRequest["state"] = "progressing"
			if s := utils.GetString(schemes[0], "custom_progressing_status"); s != "" {
				newRecRequest["state"] = s
			}
		}
		newRecRequest["current_index"] = current_index
		for _, scheme := range schemes { // verify before
			if utils.GetBool(scheme, "before_hierarchical_validation") {
				newRecRequest["current_index"] = current_index - 0.1
				break
			}
		}

		newRecRequest = SetClosureStatus(newRecRequest)
		s.Domain.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBRequest.Name, newRecRequest[utils.SpecialIDParam]), newRecRequest)
		for _, scheme := range schemes {
			if current_index != newRecRequest.GetFloat("current_index") {
				HandleHierarchicalVerification(s.Domain, utils.GetInt(res, ds.RequestDBField), record)
			} else {
				newTask := NewTask(
					scheme["name"],
					scheme["description"],
					scheme["urgency"],
					scheme["priority"],
					scheme[utils.SpecialIDParam],
					scheme[SchemaDBField],
					requests[0][utils.SpecialIDParam],
					scheme[UserDBField],
					scheme[EntityDBField],
					scheme["send_mail_to"],
				)
				if utils.GetBool(scheme, "assign_to_creator") {
					newTask[ds.UserDBField] = s.Domain.GetUserID()
				}
				if scheme[SchemaDBField] == requests[0][SchemaDBField] {
					newTask[SchemaDBField] = requests[0][SchemaDBField]
					newTask[DestTableDBField] = requests[0][DestTableDBField]
				} else if schema, err := schserv.GetSchemaByID(utils.GetInt(scheme, SchemaDBField)); err == nil {
					r := utils.Record{"is_draft": true}
					if schema.HasField("name") {
						r["name"] = utils.GetString(newTask, "name")
					}
					if schema.HasField(ds.DestTableDBField) && schema.HasField(ds.SchemaDBField) {
						// get workflow source schema + dest ID
						r[ds.DestTableDBField] = requests[0][ds.DestTableDBField]
						r[ds.SchemaDBField] = requests[0][ds.SchemaDBField]
					}
					if schema.HasField(ds.UserDBField) {
						r[ds.UserDBField] = record[ds.UserDBField]
					}
					if schema.HasField(ds.EntityDBField) {
						r[ds.EntityDBField] = record[ds.EntityDBField]
					}
					for _, f := range schema.Fields {
						if f.GetLink() == requests[0][ds.SchemaDBField] {
							r[f.Name] = requests[0][ds.DestTableDBField]
						}
					}
					if i, err := s.Domain.GetDb().CreateQuery(schema.Name, r, func(s string) (string, bool) {
						return "", true
					}); err == nil {
						newTask[SchemaDBField] = scheme[SchemaDBField]
						newTask[DestTableDBField] = i
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
				if strings.Contains(utils.GetString(res, "nexts"), utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField)) && utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField) != "" {
					newMetaRequest := utils.Record{
						ds.WorkflowDBField: scheme["wrapped_"+ds.WorkflowDBField],
						sm.NAMEKEY:         "Meta request for " + newTask.GetString(sm.NAMEKEY) + " task",
						"current_index":    1,
						"is_meta":          true,
						SchemaDBField:      newTask[SchemaDBField],
						DestTableDBField:   newTask[DestTableDBField],
						UserDBField:        requests[0][UserDBField],
					}
					requests, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name), newMetaRequest)
					if err == nil && len(requests) > 0 {
						newTask["meta_"+RequestDBField] = requests[0][utils.SpecialIDParam]
					}
				}
				if utils.GetString(res, "nexts") == utils.ReservedParam || (strings.Contains(utils.GetString(res, "nexts"),
					utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField)) && utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField) != "") {
					i, err := s.Domain.GetDb().CreateQuery(ds.DBTask.Name, newTask, func(s string) (string, bool) {
						return "", true
					})
					if err != nil {
						continue
					}
					schema, err := schserv.GetSchema(ds.DBTask.Name)
					if err == nil && newTask["meta_"+RequestDBField] == nil {
						s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), utils.Record{"link_id": schema.ID,
							sm.NAMEKEY:       newTask.GetString(sm.NAMEKEY),
							"description":    "Task is affected : " + newTask.GetString(sm.NAMEKEY),
							UserDBField:      utils.GetInt(newTask, UserDBField),
							EntityDBField:    scheme[EntityDBField],
							UserDBField:      scheme[UserDBField],
							DestTableDBField: i,
						})
					}
				}
			}
		}
	}

}

func (s *TaskService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	if !s.Domain.IsSuperCall() {
		innerestr = append(innerestr, conn.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"meta_" + RequestDBField: nil,
		}, true))
	}
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

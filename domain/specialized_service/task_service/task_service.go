package task_service

import (
	"errors"
	"math"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	conn "sqldb-ws/infrastructure/connector/db"
)

type TaskService struct {
	servutils.AbstractSpecializedService
	Redirect bool
}

func NewTaskService() utils.SpecializedServiceITF {
	return &TaskService{}
}

func (s *TaskService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	// TODO: here send back my passive task...
	res := view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
	if len(results) == 1 && s.Redirect && utils.GetBool(results[0], "is_close") {
		// retrieve... tasks affected to you
		if r, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.RequestDBField: results[0][ds.RequestDBField],
			"is_close":        false,
			utils.SpecialIDParam: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
				ds.EntityDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBEntityUser.Name, map[string]interface{}{
					ds.UserDBField: s.Domain.GetUserID(),
				}, false, ds.EntityDBField),
			}, true, utils.SpecialIDParam),
		}, false); err == nil && len(r) > 0 {
			if sch, err := schema.GetSchema(ds.DBTask.Name); err == nil {
				res[0]["inner_redirection"] = utils.BuildPath(sch.ID, utils.GetString(r[0], utils.SpecialIDParam))
			}
		} else {
			if sch, err := schema.GetSchema(ds.DBRequest.Name); err == nil {
				res[0]["inner_redirection"] = utils.BuildPath(sch.ID, utils.GetString(results[0], ds.RequestDBField))
			}
		}
	} else {
		if sch, err := schema.GetSchema(ds.DBTask.Name); err == nil && len(results) > 0 {
			res[0]["inner_redirection"] = utils.BuildPath(sch.ID, utils.GetString(results[0], utils.SpecialIDParam))
		}
	} // inner_redirection is the way to redirect any closure... to next data or data
	return res
}

func (s *TaskService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
		utils.SpecialIDParam: record[ds.RequestDBField],
	}, false); err == nil && len(res) > 0 {
		CreateDelegated(record, res[0], utils.GetInt(record, utils.SpecialIDParam), s.Domain)
	}
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}
func (s *TaskService) Entity() utils.SpecializedServiceInfo { return ds.DBTask }

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
		if elder, _ := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{utils.SpecialIDParam: record[utils.SpecialIDParam]},
			false); len(elder) > 0 && CheckStateIsEnded(utils.ToString(elder[0]["state"])) {
			return record, errors.New("task is already closed, you cannot change its state"), false
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
		res["state"] = "refused"
		results[i] = SetClosureStatus(res)
	}
	s.Write(results, map[string]interface{}{})
}

func (s *TaskService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	s.Write(results, record)
	s.Redirect = true
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
}

func (s *TaskService) Write(results []map[string]interface{}, record map[string]interface{}) {
	for _, res := range results {
		if _, ok := res["is_draft"]; ok && utils.GetBool(res, "is_draft") {
			continue
		}
		if !CheckStateIsEnded(res["state"]) {
			continue
		}
		requests, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
			utils.SpecialIDParam: utils.GetInt(res, RequestDBField),
		}, false)
		if err != nil || len(requests) == 0 {
			continue
		}
		UpdateDelegated(res, requests[0], s.Domain)

		order := requests[0]["current_index"]
		if otherPendingTasks, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name,
			map[string]interface{}{ // delete all notif
				utils.SpecialIDParam: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name,
					map[string]interface{}{
						ds.UserDBField: s.Domain.GetUserID(),
						ds.EntityDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBEntityUser.Name,
							map[string]interface{}{
								ds.UserDBField: s.Domain.GetUserID(),
							}, false, ds.EntityDBField),
					}, true, utils.SpecialIDParam),
				RequestDBField:  utils.ToString(res[RequestDBField]),
				"state":         []string{"'pending'", "'progressing'"},
				"binded_dbtask": nil,
			}, false); err == nil && len(otherPendingTasks) > 0 {
			continue
		}
		beforeSchemes, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name,
			map[string]interface{}{
				utils.SpecialIDParam: res[ds.WorkflowSchemaDBField],
			}, false)
		isOptionnal := false
		if len(beforeSchemes) > 0 && err == nil {
			isOptionnal = utils.GetBool(beforeSchemes[0], "optionnal")
		}
		current_index := utils.ToFloat64(order)
		switch res["state"] {
		case "completed":
			current_index = math.Floor(current_index + 1)
		case "dismiss":
			if !isOptionnal {
				if current_index >= 1 {
					current_index = math.Floor(current_index - 1)
				} else { // Dismiss will close requests.
					res["state"] = "refused"
				}
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
		if res["state"] == "refused" && !isOptionnal {
			newRecRequest["state"] = res["state"]
		} else {
			newRecRequest["current_index"] = current_index
			for _, scheme := range schemes { // verify before
				if utils.GetBool(scheme, "before_hierarchical_validation") {
					newRecRequest["current_index"] = current_index - 0.1
					break
				}
			}
		}
		newRecRequest = SetClosureStatus(newRecRequest)
		if utils.GetString(res, "closing_comment") != "" && CheckStateIsEnded(newRecRequest["state"]) {
			newRecRequest["closing_comment"] = utils.GetString(res, "closing_comment")
		}
		s.Domain.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBRequest.Name, newRecRequest[utils.SpecialIDParam]).RootRaw(), newRecRequest)

		for _, scheme := range schemes {
			if current_index != newRecRequest.GetFloat("current_index") && current_index != (newRecRequest.GetFloat("current_index")-1) {
				HandleHierarchicalVerification(s.Domain, requests[0], res)
			} else if current_index == newRecRequest.GetFloat("current_index") {
				PrepareAndCreateTask(scheme, requests[0], res, s.Domain, true)
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

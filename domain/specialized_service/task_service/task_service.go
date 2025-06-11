package task_service

import (
	"errors"
	"math"
	"sqldb-ws/domain/domain_service/filter"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	conn "sqldb-ws/infrastructure/connector/db"
	"time"
)

type TaskService struct {
	servutils.AbstractSpecializedService
}

func (s *TaskService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	CreateDelegated(record, s.Domain)
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
	sqlFilter = append(sqlFilter, conn.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"all_tasks":    true,
		ds.UserDBField: s.Domain.GetUserID(),
	}, false))
	if dels, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDelegation.Name, utils.ToListAnonymized(sqlFilter), false); err == nil && len(dels) > 0 {
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

func (s *TaskService) Write(results []map[string]interface{}, record map[string]interface{}) {
	for _, res := range results {
		if _, ok := res["is_draft"]; ok && utils.GetBool(res, "is_draft") {
			continue
		}
		if binded, ok := res["binded_dbtask"]; ok && utils.GetBool(res, "is_close") && binded != nil {
			s.Domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBTask.Name, map[string]interface{}{
				"is_close":                    res["is_close"],
				"state":                       res["state"],
				"closing_by" + ds.UserDBField: utils.GetInt(res, ds.UserDBField),
				"closing_date":                res["closing_date"],
			}, map[string]interface{}{
				utils.SpecialIDParam: binded,
				utils.SpecialIDParam + "_1": s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
					ds.DBTask.Name, map[string]interface{}{
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
		case "dismiss":
			if current_index >= 1 {
				current_index = math.Floor(current_index - 1)
			} else { // Dismiss will close requests.
				res["state"] = "refused"
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
		if res["state"] == "refused" {
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
		s.Domain.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBRequest.Name, newRecRequest[utils.SpecialIDParam]).RootRaw(), SetClosureStatus(newRecRequest))
		for _, scheme := range schemes {
			if current_index != newRecRequest.GetFloat("current_index") {
				HandleHierarchicalVerification(s.Domain, utils.GetInt(res, ds.RequestDBField), record)
			} else {
				PrepareAndCreateTask(scheme, record, s.Domain, true)
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

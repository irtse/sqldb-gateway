package user_service

import (
	"errors"
	ds "sqldb-ws/domain/schema/database_resources"
	task "sqldb-ws/domain/specialized_service/task_service"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
)

type DelegationService struct {
	servutils.AbstractSpecializedService
}

func (s *DelegationService) Entity() utils.SpecializedServiceInfo { return ds.DBDelegation }

func (s *DelegationService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record[ds.UserDBField] = s.Domain.GetUserID() // affected create_by
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBDelegation.Name, map[string]interface{}{
		ds.SchemaDBField:              record[ds.SchemaDBField],
		ds.DestTableDBField:           record[ds.DestTableDBField],
		ds.UserDBField:                record[ds.UserDBField],
		"delegated_" + ds.UserDBField: record["delegated_"+ds.UserDBField],
	}, false); err == nil && len(res) > 0 {
		return map[string]interface{}{}, errors.New("can't add a delegated to an already delegated user"), false
	}

	if _, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
	} else {
		return record, err, false
	}
}

func (s *DelegationService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.Write([]map[string]interface{}{record}, record)
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}

func (s *DelegationService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for i, res := range results {
		share := map[string]interface{}{
			"shared_" + ds.UserDBField: res["delegated_"+ds.UserDBField],
			ds.UserDBField:             res[ds.UserDBField],
			ds.SchemaDBField:           res[ds.SchemaDBField],
			ds.DestTableDBField:        res[ds.DestTableDBField],
		}
		s.Domain.GetDb().DeleteQueryWithRestriction(ds.DBShare.Name, share, false)
		res["state"] = "completed"
		results[i] = task.SetClosureStatus(res)
	}
	s.SpecializedUpdateRow(results, map[string]interface{}{})
}

func (s *DelegationService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	s.Write(results, record)
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
}

func (s *DelegationService) Write(results []map[string]interface{}, record map[string]interface{}) {
	for _, rr := range results {
		share := map[string]interface{}{
			"shared_" + ds.UserDBField: rr["delegated_"+ds.UserDBField],
			ds.UserDBField:             rr[ds.UserDBField],
			ds.SchemaDBField:           rr[ds.SchemaDBField],
			ds.DestTableDBField:        rr[ds.DestTableDBField],
		}
		s.Domain.GetDb().DeleteQueryWithRestriction(ds.DBShare.Name, share, false)
		share["start_date"] = rr["start_date"]
		share["end_date"] = rr["end_date"]
		share["delete_access"] = rr["delete_access"]
		s.Domain.CreateSuperCall(utils.AllParams(ds.DBShare.Name), share)
		if taskID := utils.GetInt(rr, ds.TaskDBField); taskID >= 0 {
			if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				"is_close":           false,
				utils.SpecialIDParam: taskID,
			}, false); err == nil && len(res) > 0 {
				for _, r := range res {
					newTask := utils.Record{}
					for k, v := range r {
						newTask[k] = v
					}
					newTask[ds.UserDBField] = rr["delegated_"+ds.UserDBField]
					newTask[ds.EntityDBField] = nil
					newTask["binded_"+ds.TaskDBField] = r[utils.SpecialIDParam]
					delete(newTask, utils.SpecialIDParam)
					s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name).RootRaw(), newTask)
				}
			}
		} else if utils.GetBool(rr, "all_tasks") {
			if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				"is_close": false,
				ds.EntityDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
					ds.UserDBField: s.Domain.GetUserID(),
				}, false, ds.EntityDBField),
				ds.UserDBField: s.Domain.GetUserID(),
			}, false); err == nil && len(res) > 0 {
				for _, r := range res {
					go func() {
						newTask := utils.Record{}
						for k, v := range r {
							newTask[k] = v
						}
						newTask[ds.UserDBField] = rr["delegated_"+ds.UserDBField]
						newTask[ds.EntityDBField] = nil
						newTask["binded_"+ds.TaskDBField] = r[utils.SpecialIDParam]
						delete(newTask, utils.SpecialIDParam)
						s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name).RootRaw(), newTask)
					}()
				}
			}
		}
	}
}

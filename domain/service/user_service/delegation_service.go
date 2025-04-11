package user_service

import (
	ds "sqldb-ws/domain/schema/database_resources"
	task "sqldb-ws/domain/service/task_service"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
)

type DelegationService struct {
	servutils.AbstractSpecializedService
}

func (s *DelegationService) Entity() utils.SpecializedServiceInfo { return ds.DBDelegation }

func (s *DelegationService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record[ds.UserDBField] = s.Domain.GetUserID() // affected create_by
	if _, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
	} else {
		return record, err, false
	}
}

func (s *DelegationService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

func (s *DelegationService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.Write([]map[string]interface{}{record}, record)
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}

func (s *DelegationService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for i, res := range results {
		res["state"] = "completed"
		results[i] = task.SetClosureStatus(res)
	}
	s.SpecializedUpdateRow(results, map[string]interface{}{})
}

func (s *DelegationService) Write(results []map[string]interface{}, record map[string]interface{}) {
	if taskID := utils.GetInt(record, ds.TaskDBField); taskID >= 0 {
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			utils.SpecialIDParam: taskID,
		}, false); err == nil && len(res) > 0 {
			r := res[0]
			newTask := utils.Record{}
			for k, v := range r {
				newTask[k] = v
			}
			newTask[ds.UserDBField] = res[0]["delegated_"+ds.UserDBField]
			newTask[ds.EntityDBField] = nil
			newTask["binded_"+ds.TaskDBField] = r[utils.SpecialIDParam]
			delete(r, utils.SpecialIDParam)
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name), newTask)
		}
	}
}

func (s *DelegationService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	s.Write(results, record)
	s.AbstractSpecializedService.SpecializedUpdateRow(results, record)
}

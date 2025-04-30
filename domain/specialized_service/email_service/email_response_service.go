package email_service

import (
	"encoding/json"
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/specialized_service/task_service"
	servutils "sqldb-ws/domain/specialized_service/utils"
	utils "sqldb-ws/domain/utils"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailResponseService struct {
	servutils.AbstractSpecializedService
	Code string
}

func (s *EmailResponseService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailResponse }

func (s *EmailResponseService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	// check waiting for response
	if s.Code == "" {
		return record, errors.New("no code found"), false
	}
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		"code": s.Code,
	}, false); err == nil && len(res) > 0 {
		record[ds.EmailSendedDBField] = res[0][utils.SpecialIDParam]
	} else {
		return record, errors.New("no related found"), false
	}
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailResponseService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		utils.SpecialIDParam: utils.GetString(record, ds.EmailSendedDBField),
	}, false); err == nil {
		for _, r := range res {
			if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
				ds.EmailTemplateDBField: r[ds.EmailTemplateDBField],
			}, false); err == nil && len(res) > 0 && utils.GetBool(res[0], "generate_task") {

			}
			if templs, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
				utils.SpecialIDParam: utils.GetString(record, ds.EmailSendedDBField),
			}, false); err == nil {
				for _, t := range templs {
					if utils.GetBool(t, "generate_task") {
						if rr, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
							ds.DestTableDBField: r["mapped_with"+ds.DestTableDBField],
							ds.SchemaDBField:    r["mapped_with"+ds.SchemaDBField],
							"is_close":          false,
							"name":              utils.GetString(r, "code"),
						}, false); err == nil {
							for _, rec := range rr {
								if utils.GetBool(r, "got_response") {
									rec["state"] = "completed"
								} else {
									rec["state"] = "dismiss"
								}
								rec = task_service.SetClosureStatus(rec)
								s.Domain.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBTask.Name, rec[utils.SpecialIDParam]), rec)
							}
						}
					}
					if t["action_on_response"] == nil || t[ds.SchemaDBField+"_on_response"] == nil || r[ds.DestTableDBField+"_on_response"] == nil {
						continue
					}
					var body utils.Record
					meth := utils.GetString(t, "action_on_response")
					method := utils.SELECT
					switch meth {
					case "create":
						method = utils.CREATE
						if utils.GetBool(record, "got_response") {
							if t["body_on_true_response"] == nil {
								continue
							} else {
								json.Unmarshal([]byte(utils.GetString(t, "body_on_true_response")), &body)
							}
						} else {
							if t["body_on_false_response"] == nil {
								continue
							} else {
								json.Unmarshal([]byte(utils.GetString(t, "body_on_false_response")), &body)
							}
						}
					case "update":
						method = utils.UPDATE
						if utils.GetBool(r, "got_response") {
							if t["body_on_true_response"] == nil {
								continue
							} else {
								json.Unmarshal([]byte(utils.GetString(t, "body_on_true_response")), &body)
							}
						} else {
							if t["body_on_false_response"] == nil {
								continue
							} else {
								json.Unmarshal([]byte(utils.GetString(t, "body_on_false_response")), &body)
							}
						}
					case "delete":
						method = utils.DELETE
					}
					if sch, err := schema.GetSchemaByID(utils.GetInt(t, ds.SchemaDBField+"_on_response")); err == nil {
						s.Domain.Call(utils.GetRowTargetParameters(sch.Name, r[ds.DestTableDBField+"_on_response"]), body, method)
					}
				}
			}
		}
	}
}

func (s *EmailResponseService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	s.Code, _ = s.Domain.GetParams().Get("code")
	s.Domain.GetParams().SimpleDelete("code")
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}
func (s *EmailResponseService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

package email_service

import (
	"encoding/json"
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/specialized_service/task_service"
	servutils "sqldb-ws/domain/specialized_service/utils"
	utils "sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailResponseService struct {
	servutils.AbstractSpecializedService
}

func (s *EmailResponseService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailResponse }

func (s *EmailResponseService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	code, _ := s.Domain.GetParams().Get("code")
	s.Domain.GetParams().SimpleDelete("code")
	// check waiting for response
	record["got_response"] = record["got_response"] == "true" || record["got_response"] == true
	if code == "" {
		return record, errors.New("no code found"), false
	}
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		"code": connector.Quote(code),
	}, false); err == nil && len(res) > 0 {
		record[ds.EmailSendedDBField] = res[0][utils.SpecialIDParam]
	} else {
		return record, errors.New("no related found"), false
	}
	delete(record, "code")
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailResponseService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		utils.SpecialIDParam: utils.GetString(record, ds.EmailSendedDBField),
	}, false); err == nil {
		for _, r := range res {
			if templs, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
				utils.SpecialIDParam: utils.GetString(record, ds.EmailSendedDBField),
			}, false); err == nil {
				for _, t := range templs {
					if utils.GetBool(t, "generate_task") {
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
							ds.DestTableDBField: r["mapped_with"+ds.DestTableDBField],
							ds.SchemaDBField:    r["mapped_with"+ds.SchemaDBField],
							"is_close":          false,
							"name":              connector.Quote(utils.GetString(r, "code")),
						}, false); err == nil {
							for _, rec := range rr {
								if utils.GetBool(r, "got_response") {
									rec["state"] = "completed"
								} else {
									rec["state"] = "refused"
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
					method := utils.SELECT
					switch utils.GetString(t, "action_on_response") {
					case "create":
						method = utils.CREATE
					case "update":
						method = utils.UPDATE
					case "delete":
						method = utils.DELETE
					}
					if (method == utils.CREATE || method == utils.UPDATE) && utils.GetBool(record, "got_response") {
						if t["body_on_true_response"] == nil && t["body_on_false_response"] == nil {
							continue
						} else if t["body_on_true_response"] == nil {
							json.Unmarshal([]byte(utils.GetString(t, "body_on_true_response")), &body)
						} else if t["body_on_false_response"] == nil {
							json.Unmarshal([]byte(utils.GetString(t, "body_on_false_response")), &body)
						}
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
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

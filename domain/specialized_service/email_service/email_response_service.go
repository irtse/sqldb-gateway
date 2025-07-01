package email_service

import (
	"encoding/json"
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/specialized_service/task_service"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	connector "sqldb-ws/infrastructure/connector"
	db "sqldb-ws/infrastructure/connector/db"
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
		"code": db.Quote(code),
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
				utils.SpecialIDParam:  utils.GetString(record, ds.EmailSendedDBField),
				"is_response_valid":   false,
				"is_response_refused": false,
			}, false); err == nil {
				for _, t := range templs {
					if utils.GetBool(t, "generate_task") {
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
							ds.DestTableDBField: r["mapped_with"+ds.DestTableDBField],
							ds.SchemaDBField:    r["mapped_with"+ds.SchemaDBField],
							"is_close":          false,
							"name":              db.Quote(utils.GetString(r, "code")),
						}, false); err == nil {
							for _, rec := range rr {
								if utils.GetBool(r, "got_response") {
									rec["state"] = "completed"
								} else {
									rec["state"] = "refused"
								}
								rec = task_service.SetClosureStatus(rec)
								s.Domain.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBTask.Name, rec[utils.SpecialIDParam]).RootRaw(), rec)
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
			var key = "is_response_valid"
			if utils.GetBool(record, "got_response") {
				key = "is_response_refused"
			}
			if templs, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
				key: true,
			}, false); err == nil && len(templs) > 0 {
				tmp := templs[0]
				if usr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
					utils.SpecialIDParam: r["from_email"],
				}, false); err == nil && len(usr) > 0 {
					sch, _ := schema.GetSchema(ds.DBEmailResponse.Name)
					rec, err := connector.ForgeMail(usr[0], usr[0],
						utils.GetString(tmp, "subject"), utils.GetString(tmp, "template"),
						map[string]interface{}{
							"from_email": utils.GetString(usr[0], "email"),
						}, s.Domain, utils.GetInt(tmp, utils.SpecialIDParam),
						utils.ToInt64(sch.ID), -1, -1, "", "")
					if err == nil {
						go connector.SendMail(
							utils.GetString(usr[0], "email"), utils.GetString(usr[0], "email"), rec, false)
					}
				}
			}
		}
	}
}

func (s *EmailResponseService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

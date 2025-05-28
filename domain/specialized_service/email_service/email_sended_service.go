package email_service

import (
	"sqldb-ws/domain/domain_service/view_convertor"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"

	"github.com/google/uuid"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailSendedService struct {
	servutils.AbstractSpecializedService
	To string
}

func (s *EmailSendedService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailSended }

func (s *EmailSendedService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
		utils.SpecialIDParam: record[ds.EmailTemplateDBField],
	}, false); err == nil && len(res) > 0 {
		if utils.GetBool(res[0], "generate_task") {
			i := int64(-1)
			if t, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
				"is_meta":           false,
				"is_close":          false,
				ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
				ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
			}, false); err == nil && len(t) > 0 {
				if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
					"name":              connector.Quote("waiting mails responses"),
					"current_index":     utils.GetFloat(t[0], "current_index"),
					"is_meta":           true,
					"is_close":          false,
					ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
					ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
				}, false); err == nil && len(res) > 0 {
					i = utils.GetInt(res[0], utils.SpecialIDParam)
				} else {
					if id, err := s.Domain.GetDb().CreateQuery(ds.DBRequest.Name, map[string]interface{}{
						"name":              "waiting mails responses",
						"current_index":     1,
						"is_meta":           true,
						ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
						ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
					}, func(s string) (string, bool) { return "", true }); err == nil {
						i = id
					} else {
						return
					}
				}
				if i >= 0 {
					for _, r := range t {
						if tt, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
							ds.RequestDBField:           r[utils.SpecialIDParam],
							"meta_" + ds.RequestDBField: i,
							"name":                      connector.Quote("waiting mails responses"),
						}, false); err != nil || len(tt) == 0 {
							s.Domain.GetDb().CreateQuery(ds.DBTask.Name, map[string]interface{}{
								ds.DestTableDBField:         r[ds.DestTableDBField],
								"name":                      "waiting mails responses",
								ds.SchemaDBField:            r[ds.SchemaDBField],
								ds.RequestDBField:           r[utils.SpecialIDParam],
								"meta_" + ds.RequestDBField: i,
							}, func(v string) (string, bool) { return "", true })
						}
					}
				}
				s.Domain.GetDb().CreateQuery(ds.DBTask.Name, map[string]interface{}{
					ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
					ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
					ds.RequestDBField:   i,
					"name":              utils.GetString(record, "code"),
				}, func(s string) (string, bool) { return "", true })
			}
		}
	}
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
	if s.To != "" {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: s.To,
		}, false); err == nil && len(res) > 0 {
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBEmailSendedUser.Name).RootRaw(), map[string]interface{}{
				"name":                utils.GetString(res[0], "email"),
				ds.UserDBField:        s.To,
				ds.EmailSendedDBField: record[utils.SpecialIDParam],
			}, func(s string) (string, bool) { return "", true })
		}
	}
}

func (s *EmailSendedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		"code": connector.Quote(utils.GetString(record, "code")),
	}, false); err == nil && len(res) > 0 {
		record["code"] = uuid.New()
	}
	if to := utils.GetString(record, "to_email"); to != "" {
		s.To = to
		delete(record, "to_email")
	}
	if record["code"] == nil || record["code"] == "" {
		record["code"] = uuid.New()
	}
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailSendedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

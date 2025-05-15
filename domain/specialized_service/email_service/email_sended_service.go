package email_service

import (
	"sqldb-ws/domain/domain_service/view_convertor"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
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
	}, false); err == nil && len(res) > 0 && utils.GetBool(res[0], "generate_task") {
		i := int64(-1)
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
			ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
			ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
		}, false); err == nil && len(res) > 0 && utils.GetBool(res[0], "generate_task") {
			i = utils.GetInt(res[0], utils.SpecialIDParam)
		} else {
			if id, err := s.Domain.GetDb().CreateQuery(ds.DBRequest.Name, map[string]interface{}{
				"name":              "generate from trigger mail",
				"is_close":          true,
				"current_index":     0,
				ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
				ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
			}, func(s string) (string, bool) { return "", true }); err == nil && len(res) > 0 {
				i = id
			}
		}
		if i >= 0 {
			s.Domain.GetDb().CreateQuery(ds.DBTask.Name, map[string]interface{}{
				ds.DestTableDBField: record["mapped_with"+ds.DestTableDBField],
				ds.SchemaDBField:    record["mapped_with"+ds.SchemaDBField],
				ds.RequestDBField:   i,
				"name":              utils.GetString(record, "code"),
			}, func(s string) (string, bool) { return "", true })
		}
	}
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
	if s.To != "" {
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: s.To,
		}, false); err == nil && len(res) > 0 {
			s.Domain.GetDb().CreateQuery(ds.DBEmailSendedUser.Name, map[string]interface{}{
				"name":                utils.GetString(res[0], "email"),
				ds.UserDBField:        s.To,
				ds.EmailSendedDBField: record[utils.SpecialIDParam],
			}, func(s string) (string, bool) { return "", true })
		}
	}
}

func (s *EmailSendedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if to := utils.GetString(record, "to_email"); to != "" {
		s.To = to
		delete(record, "to_email")
	} /*else {
		return record, errors.New("no user to send mail"), false
	}*/
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailSendedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

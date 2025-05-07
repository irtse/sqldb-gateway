package email_service

import (
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/domain_service/view_convertor"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailSendedService struct {
	servutils.AbstractSpecializedService
}

func (s *EmailSendedService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailSended }

func (s *EmailSendedService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	isValid := false
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
		utils.SpecialIDParam: record[ds.EmailTemplateDBField],
	}, false); err == nil && len(res) > 0 && utils.GetBool(res[0], "generate_task") {
		if utils.GetBool(res[0], "waiting_response") {
			// should enrich with a binary response yes or no.
			isValid = true
		}
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

	triggers.SendMail(utils.GetString(record, "from_email"), utils.GetString(record, "to_email"), record, utils.GetString(record, "id"), isValid)
}

func (s *EmailSendedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record["got_response"] = record["got_response"] == "true"
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailSendedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

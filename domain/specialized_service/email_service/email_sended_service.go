package email_service

import (
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	utils "sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"

	"github.com/google/uuid"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailSendedService struct {
	servutils.AbstractSpecializedService
}

func (s *EmailSendedService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailSended }

func (s *EmailSendedService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
		utils.SpecialIDParam: utils.GetString(record, ds.EmailTemplateDBField),
	}, false); err == nil && len(res) > 0 {
		tmpl := res[0]
		if utils.GetBool(tmpl, "waiting_response") {
			// create a email response
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBEmailResponse.Name), utils.Record{
				ds.EmailSendedDBField: utils.GetString(record, utils.SpecialIDParam),
			})
		}
	}
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}

func (s *EmailSendedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record["code"] = uuid.New()
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailSendedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	// TODO add fill mail on empty
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

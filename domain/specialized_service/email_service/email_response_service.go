package email_service

import (
	"errors"
	"sqldb-ws/domain/filter"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	utils "sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
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

func (s *EmailResponseService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	s.Code, _ = s.Domain.GetParams().Get("code")
	s.Domain.GetParams().SimpleDelete("code")
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}
func (s *EmailResponseService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

package email_service

import (
	"errors"
	"fmt"
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/domain_service/view_convertor"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type EmailSendedUserService struct {
	servutils.AbstractSpecializedService
}

func (s *EmailSendedUserService) Entity() utils.SpecializedServiceInfo { return ds.DBEmailSendedUser }

func (s *EmailSendedUserService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	isValid := false
	emailTo := ""
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
		utils.SpecialIDParam: s.Domain.GetDb().BuildSelectQueryWithRestriction(
			ds.DBEmailSended.Name, map[string]interface{}{
				utils.SpecialIDParam: record[ds.EmailSendedDBField],
			}, false, ds.EmailTemplateDBField,
		),
	}, false); err == nil && len(res) > 0 {
		if utils.GetBool(res[0], "waiting_response") {
			// should enrich with a binary response yes or no.
			isValid = true
		}
	}
	if record[ds.UserDBField] != nil {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: record[ds.UserDBField],
		}, false); err == nil && len(res) > 0 {
			emailTo = utils.GetString(res[0], "email")
		}
	} else if record["name"] != nil {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			"name": record["name"],
		}, false); err == nil && len(res) > 0 {
			emailTo = utils.GetString(res[0], "email")
		}
	}
	fmt.Println(emailTo)
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
		utils.SpecialIDParam: s.Domain.GetUserID(),
	}, false); err == nil && len(res) > 0 && emailTo != "" {
		if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
			utils.SpecialIDParam: record[ds.EmailSendedDBField],
		}, false); err == nil && len(rr) > 0 {
			err = triggers.SendMail(utils.GetString(res[0], "email"), emailTo, rr[0], isValid)
			fmt.Println("SENDING MAIL :", err)
		}
		fmt.Println("SENDING MAIL :", err)
	} else {
		fmt.Println(res, emailTo)
		fmt.Println("can't email because of a missing <send to> user")
	}
}

func (s *EmailSendedUserService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if utils.GetString(record, "name") == "" && utils.GetString(record, ds.UserDBField) == "" {
		return record, errors.New("no email to send to"), false
	}
	return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *EmailSendedUserService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

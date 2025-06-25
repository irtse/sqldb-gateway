package user_service

import (
	"fmt"
	"sqldb-ws/domain/domain_service/filter"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	connector "sqldb-ws/infrastructure/connector/db"
	"strings"
)

type UserService struct {
	servutils.AbstractSpecializedService
}

func (s *UserService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record["name"] = strings.ToLower(utils.GetString(record, "name"))
	record["email"] = strings.ToLower(utils.GetString(record, "email"))
	return record, nil, true
}
func (s *UserService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}
func (s *UserService) Entity() utils.SpecializedServiceInfo { return ds.DBUser }

func (s *UserService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	if scope, ok := s.Domain.GetParams().Get(utils.RootScope); ok && scope == "enable" && s.Domain.GetUserID() != "" {
		innerestr = append(innerestr, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"!" + utils.SpecialIDParam: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBShare.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
			}, true, "shared_"+ds.UserDBField),
		}, true))
	} else if scope, ok := s.Domain.GetParams().Get(utils.RootScope); ok && scope == "disable" && s.Domain.GetUserID() != "" {
		innerestr = append(innerestr, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			utils.SpecialIDParam: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBShare.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
			}, true, "shared_"+ds.UserDBField),
		}, true))
		fmt.Println("SHARE", innerestr)
	}
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

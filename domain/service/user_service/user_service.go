package user_service

import (
	"sqldb-ws/domain/filter"
	ds "sqldb-ws/domain/schema/database_resources"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	"sqldb-ws/infrastructure/connector"
)

type UserService struct {
	servutils.AbstractSpecializedService
}

func (s *UserService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}
func (s *UserService) Entity() utils.SpecializedServiceInfo { return ds.DBUser }
func (s *UserService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}
func (s *UserService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	if scope, ok := s.Domain.GetParams().Get(utils.RootScope); ok && scope == "enable" && s.Domain.GetUserID() != "" {
		innerestr = append(innerestr, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			utils.SpecialIDParam: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBEntityUser.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
			}, true, ds.UserDBField),
			utils.SpecialIDParam + "_1": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
				"parent_" + ds.UserDBField: s.Domain.GetUserID(),
			}, true, ds.UserDBField),
			utils.SpecialIDParam + "_2": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
				ds.UserDBField: s.Domain.GetUserID(),
			}, true, "parent_"+ds.UserDBField),
		}, true))
	}
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

package user_service

import (
	"errors"
	"sqldb-ws/domain/filter"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"

	"sqldb-ws/domain/view_convertor"
	"sqldb-ws/infrastructure/connector"
)

type ShareService struct {
	servutils.AbstractSpecializedService
}

func (s *ShareService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}
func (s *ShareService) Entity() utils.SpecializedServiceInfo { return ds.DBDelegation }
func (s *ShareService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}
func (s *ShareService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	record[ds.UserDBField] = s.Domain.GetUserID() // affected create_by
	sch, err := schema.GetSchema(tablename)
	if err != nil {
		return record, errors.New("not schema found"), false
	}
	if !s.Domain.VerifyAuth(sch.Name, "", sm.LEVELNORMAL, utils.UPDATE) {
		record["update_access"] = false
	}
	if !s.Domain.VerifyAuth(sch.Name, "", sm.LEVELNORMAL, utils.CREATE) {
		record["create_access"] = false
	}
	if !s.Domain.VerifyAuth(sch.Name, "", sm.LEVELNORMAL, utils.DELETE) {
		record["delete_access"] = false
	}
	if _, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(record, tablename)
	} else {
		return record, err, false
	}
}

func (s *ShareService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	if s.Domain.IsSuperCall() {
		innerestr = append(innerestr, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			ds.UserDBField: s.Domain.GetUserID(),
		}, true))
	}
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

package utils

import (
	"sqldb-ws/domain/filter"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
)

type AbstractSpecializedService struct{ Domain utils.DomainITF }

func (s *AbstractSpecializedService) SetDomain(d utils.DomainITF) { s.Domain = d }

type SpecializedService struct{ AbstractSpecializedService }

func (s *SpecializedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true)
}
func (s *SpecializedService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, innerestr...)
}

func CheckAutoLoad(tablename string, record utils.Record, domain utils.DomainITF) (utils.Record, error, bool) {
	if domain.GetMethod() != utils.DELETE {
		if rec, err := sch.ValidateBySchema(record, tablename,
			domain.GetMethod(), domain.VerifyAuth); err != nil && !domain.GetAutoload() {
			return rec, err, false
		}
	}
	return record, nil, domain.GetMethod() == utils.DELETE
}

func GetUserRecord(domain utils.DomainITF) (utils.Record, bool) {
	userRecords, _ := domain.GetDb().SelectQueryWithRestriction(
		ds.DBUser.Name,
		map[string]interface{}{
			"name":  domain.GetUser(),
			"email": domain.GetUser(),
		}, true)
	if len(userRecords) > 0 {
		return userRecords[0], true
	}
	return nil, false
}

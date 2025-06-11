package history

import (
	"slices"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	connector "sqldb-ws/infrastructure/connector/db"
)

func NewDataAccess(schemaID int64, destIDs []string, domain utils.DomainITF) {
	if users, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
		"name":  connector.Quote(domain.GetUser()),
		"email": connector.Quote(domain.GetUser()),
	}, true); err == nil && len(users) > 0 {
		for _, destID := range destIDs {
			id := utils.GetString(users[0], utils.SpecialIDParam)
			if domain.GetMethod() == utils.SELECT {
				if res, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(
					ds.DBDataAccess.Name, map[string]interface{}{
						"write":             false,
						"update":            false,
						ds.DestTableDBField: destID,
						ds.SchemaDBField:    schemaID,
						ds.UserDBField:      id,
					}, false); err == nil && len(res) == 0 {
					domain.GetDb().ClearQueryFilter().CreateQuery(ds.DBDataAccess.Name,
						utils.Record{
							"write":             domain.GetMethod() == utils.CREATE,
							"update":            domain.GetMethod() == utils.UPDATE,
							ds.DestTableDBField: destID,
							ds.SchemaDBField:    schemaID,
							ds.UserDBField:      id}, func(s string) (string, bool) {
							return "", true
						})
				}
				return
			}
			domain.GetDb().ClearQueryFilter().CreateQuery(ds.DBDataAccess.Name,
				utils.Record{
					"write":             domain.GetMethod() == utils.CREATE,
					"update":            domain.GetMethod() == utils.UPDATE,
					ds.DestTableDBField: destID,
					ds.SchemaDBField:    schemaID,
					ds.UserDBField:      id}, func(s string) (string, bool) {
					return "", true
				})
		}
	}
}

func CountMaxDataAccess(schema *sm.SchemaModel, filter []string, domain utils.DomainITF) (int64, string) {
	if domain.GetUserID() == "" {
		return 0, ""
	}
	restr, _, _, _ := domain.GetSpecialized(schema.Name).GenerateQueryFilter(schema.Name, filter...)
	count := int64(0)
	res, err := domain.GetDb().ClearQueryFilter().SimpleMathQuery("COUNT", schema.Name, []interface{}{restr}, false)
	if len(res) == 0 || err != nil || res[0]["result"] == nil {
		return 0, restr
	} else {
		count = utils.ToInt64(res[0]["result"])
	}
	return count, restr
}

func CountNewDataAccess(schema *sm.SchemaModel, filter []string, domain utils.DomainITF) (int64, int64) {
	if domain.GetUserID() == "" || domain.GetEmpty() {
		return 0, 0
	}
	newCount := int64(0)
	count, restr := CountMaxDataAccess(schema, filter, domain)
	newFilter := map[string]interface{}{
		"!id": domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
			"write":  false,
			"update": false,
			ds.SchemaDBField: domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
				ds.DBSchema.Name, map[string]interface{}{
					"name": connector.Quote(schema.Name),
				}, false, "id"),
			ds.UserDBField: domain.GetUserID(),
		}, false, ds.DestTableDBField),
	}
	filter = []string{restr}
	filter = append(filter, connector.FormatSQLRestrictionWhereByMap("", newFilter, false))
	if res, err := domain.GetDb().ClearQueryFilter().SimpleMathQuery("COUNT", schema.Name, utils.ToListAnonymized(filter), false); err == nil && len(res) > 0 {
		newCount = utils.ToInt64(res[0]["result"])
	}
	return newCount, count
}

func GetCreatedAccessData(schemaID string, domain utils.DomainITF) []string {
	ids := []string{}
	if dataAccess, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDataAccess.Name,
		map[string]interface{}{
			"write":          true,
			ds.SchemaDBField: schemaID,
			ds.UserDBField:   domain.GetUserID(),
		}, false); err == nil && len(dataAccess) > 0 {
		for _, access := range dataAccess {
			if !slices.Contains(ids, utils.ToString(access[utils.RootDestTableIDParam])) && utils.ToString(access[utils.RootDestTableIDParam]) != "" {
				ids = append(ids, utils.ToString(access[utils.RootDestTableIDParam]))
			}
		}
	}
	return ids
}

func GetNew(id string, schemaID string, domain utils.DomainITF) bool {
	if id == "" {
		return false
	}
	if res, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
		"write":             false,
		"update":            false,
		ds.DestTableDBField: id,
		ds.SchemaDBField:    schemaID,
		ds.UserDBField:      domain.GetUserID(),
	}, false); err == nil && len(res) > 0 {
		return false
	}
	return true
}

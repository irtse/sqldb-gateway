package filter

import (
	"fmt"
	"sort"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

// DONE - ~ 100 LINES - NOT TESTED
func (d *FilterService) GetEntityFilterQuery() string {
	return d.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
		ds.DBEntityUser.Name,
		map[string]interface{}{
			ds.UserDBField: d.Domain.GetUserID(),
		}, true, ds.EntityDBField)
}

func (d *FilterService) CountMaxDataAccess(tableName string, filter []string) (int64, string) {
	if d.Domain.GetUserID() == "" {
		return 0, ""
	}
	restr, _, _, _ := d.Domain.GetSpecialized(tableName).GenerateQueryFilter(tableName, filter...)
	fmt.Println(tableName, restr)
	count := int64(0)
	res, err := d.Domain.GetDb().ClearQueryFilter().SimpleMathQuery("COUNT", tableName, []interface{}{restr}, false)
	if len(res) == 0 || err != nil || res[0]["result"] == nil {
		return 0, restr
	} else {
		count = utils.ToInt64(res[0]["result"])
	}
	return count, restr
}

func (d *FilterService) CountNewDataAccess(tableName string, filter []string) (int64, int64) {
	if d.Domain.GetUserID() == "" || d.Domain.GetEmpty() {
		return 0, 0
	}
	newCount := int64(0)
	count, restr := d.CountMaxDataAccess(tableName, filter)
	newFilter := map[string]interface{}{
		"!id": d.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
			"write":  false,
			"update": false,
			ds.SchemaDBField: d.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
				ds.DBSchema.Name, map[string]interface{}{
					"name": connector.Quote(tableName),
				}, false, "id"),
			ds.UserDBField: d.Domain.GetUserID(),
		}, false, ds.DestTableDBField),
	}
	filter = []string{restr}
	filter = append(filter, connector.FormatSQLRestrictionWhereByMap("", newFilter, false))
	if res, err := d.Domain.GetDb().ClearQueryFilter().SimpleMathQuery("COUNT", tableName, utils.ToListAnonymized(filter), false); err == nil && len(res) > 0 {
		newCount = utils.ToInt64(res[0]["result"])
	}
	return newCount, count
}

func (s *FilterService) GetFilterFields(viewfilterID string, schemaID string) []map[string]interface{} {
	if viewfilterID == "" {
		return []map[string]interface{}{}
	}
	restriction := map[string]interface{}{}
	if schemaID != "" {
		restriction[ds.SchemaFieldDBField] = s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
			ds.DBSchemaField.Name,
			map[string]interface{}{ds.SchemaDBField: schemaID}, false, utils.SpecialIDParam)
	}
	restriction[ds.FilterDBField] = viewfilterID
	if fields, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, restriction, false); err == nil {
		sort.SliceStable(fields, func(i, j int) bool {
			return utils.ToInt64(fields[i]["index"]) <= utils.ToInt64(fields[j]["index"])
		})
		return fields
	}
	return []map[string]interface{}{}
}

func (s *FilterService) GetFilterIDs(filterID string, viewfilterID string, schemaID string) map[string]string {
	params := utils.AllParams(ds.DBFilter.Name).Enrich(map[string]interface{}{
		ds.RootID(ds.DBSchema.Name): schemaID,
	})
	filtersID := map[string]string{utils.RootFilter: filterID, utils.RootViewFilter: viewfilterID}
	for _, v := range filtersID {
		if p, ok := s.Domain.GetParams().Get(v); ok && p != "" {
			params.Set(ds.FilterDBField, p)
			restriction := map[string]interface{}{
				ds.SchemaDBField: schemaID,
				ds.FilterDBField: p,
			}
			restriction["is_view"] = v == utils.RootViewFilter
			if fields, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(
				ds.DBFilterField.Name, restriction, false); err == nil && len(fields) > 0 {
				filtersID[v] = p
			}
		}
	}
	return filtersID
}

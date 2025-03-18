package filter

import (
	"fmt"
	"slices"
	"sort"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

// DONE - ~ 100 LINES - NOT TESTED
func (d *FilterService) GetEntityFilterQuery(field string) string {
	return d.Domain.GetDb().BuildSelectQueryWithRestriction(
		ds.DBEntityUser.Name,
		map[string]interface{}{
			ds.UserDBField: d.GetUserFilterQuery(ds.UserDBField),
		}, true, field)
}

func (d *FilterService) GetUserFilterQuery(field string) string {
	return d.Domain.GetDb().BuildSelectQueryWithRestriction(
		ds.DBUser.Name,
		map[string]interface{}{
			"name":  connector.Quote(d.Domain.GetUser()),
			"email": connector.Quote(d.Domain.GetUser()),
		}, true, field)
}

func (d *FilterService) CountNewDataAccess(tableName string, filter []string) ([]string, int64) {
	newFilter := []interface{}{
		connector.FormatSQLRestrictionWhereByMap("",
			map[string]interface{}{
				utils.SpecialIDParam: "!" + d.Domain.GetDb().BuildSelectQueryWithRestriction(
					ds.DBDataAccess.Name, map[string]interface{}{
						ds.SchemaDBField: d.Domain.GetDb().BuildSelectQueryWithRestriction(
							ds.DBSchema.Name, map[string]interface{}{
								"name": tableName,
							}, false, "id"),
						ds.UserDBField: d.GetUserFilterQuery("id"),
					}, true, ds.DestTableDBField),
			}, false)}
	for _, v := range filter {
		newFilter = append(newFilter, v)
	}
	ids := []string{}
	if res, err := d.Domain.GetDb().SelectQueryWithRestriction(tableName, newFilter, false); err != nil {
		return ids, 0
	} else {
		for _, rec := range res {
			if !slices.Contains(ids, utils.GetString(rec, utils.SpecialIDParam)) {
				ids = append(ids, utils.GetString(rec, utils.SpecialIDParam))
			}
		}
	}
	res, err := d.Domain.GetDb().SimpleMathQuery("COUNT", tableName, filter, false)
	fmt.Println("newFilter", tableName, filter, res, filter)
	if len(res) == 0 || err != nil || res[0]["result"] == nil || utils.ToInt64(res[0]["result"]) == 0 {
		return ids, 0
	}
	return ids, utils.ToInt64(res[0]["result"])
}

func (s *FilterService) GetFilterFields(viewfilterID string, schemaID string) []map[string]interface{} {
	if viewfilterID == "" {
		return []map[string]interface{}{}
	}
	restriction := map[string]interface{}{}
	if schemaID != "" {
		restriction[ds.SchemaDBField] = s.Domain.GetDb().BuildSelectQueryWithRestriction(
			ds.DBSchemaField.Name,
			map[string]interface{}{ds.SchemaDBField: schemaID}, false)
	}
	restriction[ds.FilterDBField] = viewfilterID
	if fields, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilterField.Name, restriction, false); err == nil {
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
	utils.ParamsMutex.Lock()
	defer utils.ParamsMutex.Unlock()
	for _, v := range filtersID {
		if p, ok := s.Domain.GetParams()[v]; ok && p != "" {
			params[ds.FilterDBField] = p
			restriction := map[string]interface{}{
				ds.SchemaDBField: schemaID,
				ds.FilterDBField: p,
			}
			restriction["is_view"] = v == utils.RootViewFilter
			if fields, err := s.Domain.GetDb().SelectQueryWithRestriction(
				ds.DBFilterField.Name, restriction, false); err == nil && len(fields) > 0 {
				filtersID[v] = p
			}
		}
	}
	return filtersID
}

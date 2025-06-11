package view_convertor

import (
	"slices"
	"sqldb-ws/domain/domain_service/filter"
	scheme "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"strconv"
)

func (s *ViewConvertor) GetFieldsFill(sch *sm.SchemaModel, values map[string]interface{}) map[string]interface{} {
	if !s.Domain.GetEmpty() {
		return values
	}
	for k := range values {
		if f, err := sch.GetField(k); err == nil {
			values[k], _ = s.GetFieldInfo(&f, ds.DBFieldAutoFill.Name)
		}
	}
	return values
}

func (s *ViewConvertor) GetFieldsRules(sch *sm.SchemaModel, values map[string]interface{}) map[string]interface{} {
	if !s.Domain.GetEmpty() {
		return values
	}
	for k := range values {
		if f, err := sch.GetField(k); err == nil {
			val, operator := s.GetFieldInfo(&f, ds.DBFieldRule.Name)
			if utils.ToMap(values[k])["rules"] == nil {
				utils.ToMap(values[k])["rules"] = []map[string]interface{}{}
			}
			rules := utils.ToMap(values[k])["rules"].([]map[string]interface{})
			rules = append(rules, map[string]interface{}{
				"value":    val,
				"operator": operator,
			})
			utils.ToMap(values[k])["rules"] = rules
		}
	}
	return values
}

func (s *ViewConvertor) GetFieldInfo(f *sm.FieldModel, from string) (interface{}, string) {
	var value interface{}
	operator := ""
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(from, map[string]interface{}{
		ds.SchemaFieldDBField: f.ID,
	}, true); err == nil && len(res) > 0 {
		r := res[0]
		if val, ok := r["value"]; ok && val != nil {
			value = s.fromITF(val)
		} else if schFrom, err := scheme.GetSchemaByID(utils.ToInt64(r["from_"+ds.SchemaDBField])); err == nil {
			if dest, ok := r["from_"+ds.DestTableDBField]; ok && dest != nil {
				if ff, err2 := schFrom.GetFieldByID(utils.GetInt(r, "from_"+ds.SchemaFieldDBField)); err2 == nil {
					if ress, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name, map[string]interface{}{
						utils.SpecialIDParam: dest,
					}, true); err == nil && len(ress) > 0 {
						value = s.fromITF(ress[0][ff.Name])
						operator = utils.GetString(ress[0], "operator")
					}
				}
			} else if ff, err2 := schFrom.GetFieldByID(utils.GetInt(r, "from_"+ds.SchemaFieldDBField)); err2 == nil && utils.GetBool(r, "first_own") {
				if schFrom.Name == ds.DBUser.Name && ff.Name == utils.SpecialIDParam {
					value = s.Domain.GetUserID()
				} else {
					if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name,
						utils.ToListAnonymized(filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)), false); err == nil && len(rr) > 0 {
						value = s.fromITF(rr[0][ff.Name])
						operator = utils.GetString(rr[0], "operator")
					}
				}
			} else if utils.GetBool(r, "first_own") {
				if schFrom.Name == ds.DBUser.Name {
					value = s.Domain.GetUserID()
				} else {
					if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name,
						utils.ToListAnonymized(filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)),
						false); err == nil && len(rr) > 0 {
						value = s.fromITF(rr[0][utils.SpecialIDParam])
						operator = utils.GetString(rr[0], "operator")
					}
				}
			}
		}
	}
	return value, operator
}

func (s *ViewConvertor) fromITF(val interface{}) interface{} {
	if slices.Contains([]string{"true", "false"}, utils.ToString(val)) {
		return val == "true" // should set type
	} else if i, err := strconv.Atoi(utils.ToString(val)); err == nil && i >= 0 {
		return i // should set type
	} else {
		return utils.ToString(val) // should set type
	}
}

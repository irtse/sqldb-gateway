package filter

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"strings"
)

// DONE - ~ 260 LINES - NOT TESTED
type FilterService struct {
	Domain utils.DomainITF
}

func NewFilterService(domain utils.DomainITF) *FilterService {
	return &FilterService{Domain: domain}
}

func (f *FilterService) GetQueryFilter(tableName string, domainParams utils.Params, innerRestriction ...string) (string, string, string, string) {
	schema, err := sch.GetSchema(tableName)
	if err != nil {
		return "", "", "", ""
	}
	var SQLview, SQLrestriction, SQLOrder []string = []string{}, []string{}, []string{}
	var SQLLimit string
	restr, view, order, dir, state := f.GetFilterForQuery("", "", schema, domainParams)
	if restr != "" && !f.Domain.IsSuperCall() {
		SQLrestriction = append(SQLrestriction, restr)
	}
	later := []string{}
	for _, restr := range innerRestriction {
		if strings.Contains(restr, " IN ") {
			later = append(later, restr)
			continue
		}
		if restr != "" {
			r := []string{"(" + restr + ")"}
			r = append(r, SQLrestriction...)
			SQLrestriction = r
		}
	}
	if view != "" {
		domainParams.Add(utils.RootColumnsParam, view, func(v string) bool { return !f.Domain.IsSuperCall() })
	}
	if order != "" {
		domainParams.Add(utils.RootOrderParam, order, func(v string) bool { return true })
		if dir != "" {
			domainParams.Add(utils.RootDirParam, dir, func(v string) bool { return true })
		}
	}
	SQLrestriction = f.RestrictionBySchema(tableName, SQLrestriction, domainParams)
	SQLOrder = domainParams.GetOrder(func(el string) bool { return schema.HasField(el) }, SQLOrder)
	SQLLimit = domainParams.GetLimit(SQLLimit)
	SQLview = f.viewbyFields(schema, domainParams)
	if f.Domain.IsSuperCall() {
		return strings.Join(SQLrestriction, " AND "), strings.Join(SQLview, ","), strings.Join(SQLOrder, ","), SQLLimit
	}
	SQLrestriction = f.RestrictionByEntityUser(schema, SQLrestriction, false) // admin can see all on admin view
	if s, ok := domainParams.Get(utils.RootFilterNewState); ok && s != "" {
		state = s
	}
	SQLrestriction = append(SQLrestriction, later...)
	if state != "" {
		SQLrestriction = f.LifeCycleRestriction(tableName, SQLrestriction, state)
	}
	if len(SQLview) > 0 {
		SQLview = append(SQLview, "is_draft")
	}
	return strings.Join(SQLrestriction, " AND "), strings.Join(SQLOrder, ","), SQLLimit, strings.Join(SQLview, ",")
}

func (d *FilterService) RestrictionBySchema(tableName string, restr []string, domainParams utils.Params) []string {
	restriction := map[string]interface{}{}
	restriction["active"] = true
	if schema, err := sch.GetSchema(tableName); err == nil {
		if schema.HasField("is_meta") && !d.Domain.IsSuperCall() {
			restriction["is_meta"] = false
		}
		alterRestr := []string{}
		f := func(s string, search string) {
			splitted := strings.Split(s, ",")
			for _, str := range splitted {
				d.Domain.AddDetectFileToSearchIn(str, search)
			}
		}
		if line, ok := domainParams.Get(utils.RootFilterLine); ok && tableName != ds.DBView.Name {
			if connector.FormatSQLRestrictionWhereInjection(line, schema.GetTypeAndLinkForField, f) != "" {
				alterRestr = append(alterRestr, connector.FormatSQLRestrictionWhereInjection(line, schema.GetTypeAndLinkForField, f))
			}
		}
		for key, val := range domainParams.Values {
			typ, foreign, err := schema.GetTypeAndLinkForField(key, val, f)
			if err != nil && key != utils.SpecialIDParam {
				continue
			}
			newSTR := ""
			ors := strings.Split(utils.ToString(val), ",")
			for _, or := range ors {
				if len(newSTR) > 0 {
					newSTR += " OR "
				}
				newSTR += connector.MakeSqlItem("", typ, foreign, key, or, "=")
			}
			if newSTR != "" {
				alterRestr = append(alterRestr, "("+newSTR+")")
			}
		}
		newRestr := []string{}
		for _, alt := range alterRestr {
			if alt != "" {
				newRestr = append(newRestr, alt)
			}
		}
		restr = append(newRestr, restr...)
		if schema.HasField(ds.SchemaDBField) && !d.Domain.IsSuperAdmin() {
			except := []string{ds.DBRequest.Name, ds.DBTask.Name, ds.DBDelegation.Name}
			enum := []string{}
			for _, s := range sm.SchemaRegistry {
				notOK := !d.Domain.IsSuperAdmin() && ds.IsRootDB(s.Name) && !slices.Contains(except, s.Name)
				notOK2 := !d.Domain.VerifyAuth(s.Name, "", sm.LEVELNORMAL, utils.SELECT)
				if !notOK && !notOK2 {
					enum = append(enum, utils.ToString(s.ID))
				}
			}
			if connector.FormatSQLRestrictionWhereByMap(
				"", map[string]interface{}{ds.SchemaDBField: enum}, false) != "" {
				restr = append(restr, connector.FormatSQLRestrictionWhereByMap(
					"", map[string]interface{}{ds.SchemaDBField: enum}, false))
			}

		}
	}
	if strings.Trim(connector.FormatSQLRestrictionWhereByMap("", restriction, false), " ") != "" {
		restr = append(restr, strings.Trim(connector.FormatSQLRestrictionWhereByMap("", restriction, false), " "))
	}
	return restr
}

func (s *FilterService) RestrictionByEntityUser(schema sm.SchemaModel, restr []string, overrideOwn bool) []string {
	newRestr := map[string]interface{}{}
	restrictions := map[string]interface{}{}
	if s.Domain.IsOwn(false, false, s.Domain.GetMethod()) || overrideOwn {
		ids := s.GetCreatedAccessData(schema.ID)
		if len(ids) > 0 {
			newRestr[utils.SpecialIDParam] = ids
		} else {
			newRestr[utils.SpecialIDParam] = nil
		}
	} else if !s.Domain.IsShallowed() {
		restr = append(restr, "("+connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"is_draft": false,
			utils.SpecialIDParam + "_10": s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name+" as d",
				map[string]interface{}{
					"d." + ds.SchemaDBField:    schema.ID,
					"d." + ds.DestTableDBField: "main.id",
					"d." + ds.UserDBField:      s.Domain.GetUserID(),
					"d.write":                  true,
				}, false, ds.DestTableDBField),
		}, true)+")")
	} else {
		restrictions["is_draft"] = false
	}
	isUser := false
	isUser = schema.HasField(ds.UserDBField) || s.Domain.GetTable() == ds.DBUser.Name
	if scope, ok := s.Domain.GetParams().Get(utils.RootScope); !(ok && scope == "enable" && schema.Name == ds.DBTask.Name) {
		if isUser {
			key := ds.UserDBField
			if s.Domain.GetTable() == ds.DBUser.Name {
				key = utils.SpecialIDParam
			}
			if s.Domain.GetUserID() != "" {
				if scope, ok := s.Domain.GetParams().Get(utils.RootScope); ok && scope == "enable" {
					restrictions[key] = s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBHierarchy.Name, map[string]interface{}{
						"parent_" + ds.UserDBField: s.Domain.GetUserID(),
					}, true, ds.UserDBField)
				} else {
					restrictions[key] = s.Domain.GetUserID()
				}
			}
		}
		if schema.HasField(ds.EntityDBField) || s.Domain.GetTable() == ds.DBEntity.Name {
			key := ds.EntityDBField
			if s.Domain.GetTable() == ds.DBEntity.Name {
				if !ok {
					key = utils.SpecialIDParam
				}
			}
			if s.Domain.GetUserID() != "" {
				restrictions[key] = s.GetEntityFilterQuery()
			}
		}
	}
	if len(newRestr) > 0 {
		for k, r := range newRestr {
			if r != "" {
				restrictions[k] = r
			}
		}
	}
	idParamsOk := len(s.Domain.GetParams().GetAsArgs(utils.SpecialIDParam)) > 0
	if len(restrictions) > 0 && !(idParamsOk && slices.Contains(ds.PUPERMISSIONEXCEPTION, schema.Name)) {
		var s = connector.FormatSQLRestrictionWhereByMap("", restrictions, true)
		if s != "" {
			restr = append(restr, "("+s+")")
		}

	}
	return restr
}

func (d *FilterService) viewbyFields(schema sm.SchemaModel, domainParams utils.Params) []string {
	SQLview := []string{}
	views, _ := domainParams.Get(utils.RootColumnsParam)

	for _, field := range schema.Fields {
		manyOK := strings.Contains(field.Type, "many")
		if len(views) > 0 && !strings.Contains(views, field.Name) || manyOK {
			continue
		}
		if d.Domain.VerifyAuth(d.Domain.GetTable(), field.Name, field.Level, utils.SELECT) {
			SQLview = append(SQLview, field.Name)
		}
	}
	if p, ok := domainParams.Get(utils.RootCommandRow); ok {
		decodedLine, err := url.QueryUnescape(p)
		if err == nil {
			SQLview = append(SQLview, decodedLine)
		}
	}
	if len(SQLview) > 0 {
		SQLview = append(SQLview, "id")
	}
	return SQLview
}

func (s *FilterService) GetFilterForQuery(filterID string, viewfilterID string, schema sm.SchemaModel, domainParams utils.Params) (string, string, string, string, string) {
	ids := s.GetFilterIDs(filterID, viewfilterID, schema.ID)
	view, order, dir := s.ProcessViewAndOrder(ids[utils.RootViewFilter], schema.ID, domainParams)
	filter := s.ProcessFilterRestriction(ids[utils.RootFilter], schema)
	state := ""
	if filterID != "" {
		if fils, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilter.Name,
			map[string]interface{}{
				utils.SpecialIDParam: filterID,
			}, false); err == nil && len(fils) > 0 {
			state = utils.ToString(fils[0]["elder"]) // get elder filter
		}
	}
	return filter, view, order, dir, state
}

func (s *FilterService) ProcessFilterRestriction(filterID string, schema sm.SchemaModel) string {
	fmt.Println(filterID, schema.Name)
	if filterID == "" {
		return ""
	}
	var filter []string
	var orFilter []string
	restriction := map[string]interface{}{
		ds.FilterDBField: filterID,
	}
	s.Domain.GetDb().ClearQueryFilter()
	fields, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, restriction, false)
	if err == nil && len(fields) > 0 {
		for _, field := range fields {
			if f, err := schema.GetFieldByID(utils.GetInt(field, ds.SchemaFieldDBField)); err == nil {
				if utils.GetBool(field, "is_own") && len(s.RestrictionByEntityUser(schema, orFilter, true)) > 0 {
					if field["separator"] == "or" {
						orFilter = append(orFilter, s.RestrictionByEntityUser(schema, orFilter, true)...)
					} else {
						filter = append(filter, s.RestrictionByEntityUser(schema, filter, true)...)
					}
				} else if connector.FormatOperatorSQLRestriction(field["operator"], field["separator"], f.Name, field["value"], f.Type) != "" {
					if field["separator"] == "or" {
						orFilter = append(orFilter,
							"("+connector.FormatOperatorSQLRestriction(field["operator"], field["separator"], f.Name, field["value"], f.Type)+")")
					} else {
						filter = append(filter,
							"("+connector.FormatOperatorSQLRestriction(field["operator"], field["separator"], f.Name, field["value"], f.Type)+")")
					}
				}

			}
		}
	}
	if len(orFilter) > 0 {
		filter = append(filter, "("+strings.Join(orFilter, " OR ")+")")
	}
	return strings.Join(filter, " AND ")
}

func (s *FilterService) ProcessViewAndOrder(viewfilterID string, schemaID string, domainParams utils.Params) (string, string, string) {
	var viewFilter, order, dir []string = []string{}, []string{}, []string{}
	cols, ok := domainParams.Get(utils.RootColumnsParam)
	fields := []sm.FieldModel{}
	if viewfilterID != "" {
		for _, field := range s.GetFilterFields(viewfilterID, schemaID) {
			if f, err := sch.GetFieldByID(utils.GetInt(field, ds.RootID(ds.DBSchemaField.Name))); err == nil {
				fields = append(fields, f)
			}
		}
	}
	// TODO
	sort.SliceStable(fields, func(i, j int) bool {
		return fields[i].Index <= fields[j].Index
	})
	for _, field := range fields {
		if field.Name != "id" && (!ok || cols == "" || (strings.Contains(cols, field.Name))) && !field.Hidden {
			viewFilter = append(viewFilter, field.Name)
			if field.Dir != "" {
				dir = append(dir, strings.ToUpper(field.Dir))
			} else if !slices.Contains(order, field.Name) {
				dir = append(dir, field.Name+" ASC")
			}
		}
	}
	if p, ok := domainParams.Get(utils.RootGroupBy); ok {
		if len(viewFilter) != 0 && !slices.Contains(viewFilter, p) {
			viewFilter = append(viewFilter, p)
		}
		if !slices.Contains(order, p) {
			order = append(order, p)
			dir = append(dir, "ASC")
		}
	}
	return strings.Join(viewFilter, ","), strings.Join(order, ","), strings.Join(dir, ",")
}

func (d *FilterService) LifeCycleRestriction(tableName string, SQLrestriction []string, state string) []string {
	if state == "all" || tableName == ds.DBView.Name {
		return SQLrestriction
	}
	if state == "new" {
		SQLrestriction = append(SQLrestriction, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"!" + utils.SpecialIDParam: d.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name,
				map[string]interface{}{
					"write":  false,
					"update": false,
					ds.SchemaDBField: d.Domain.GetDb().BuildSelectQueryWithRestriction(
						ds.DBSchema.Name, map[string]interface{}{
							"name": connector.Quote(tableName),
						}, false, "id"),
					ds.UserDBField: d.Domain.GetUserID(),
				}, false, ds.DestTableDBField),
		}, false))
	} else {
		SQLrestriction = append(SQLrestriction, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			utils.SpecialIDParam: d.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBDataAccess.Name,
				map[string]interface{}{
					"write":  false,
					"update": false,
					ds.SchemaDBField: d.Domain.GetDb().BuildSelectQueryWithRestriction(
						ds.DBSchema.Name, map[string]interface{}{
							"name": connector.Quote(tableName),
						}, false, "id"),
					ds.UserDBField: d.Domain.GetUserID(),
				}, false, ds.DestTableDBField),
		}, false))
	}

	return SQLrestriction
}

func (d *FilterService) GetCreatedAccessData(schemaID string) []string {
	ids := []string{}
	if dataAccess, err := d.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDataAccess.Name,
		map[string]interface{}{
			"write":          true,
			ds.SchemaDBField: schemaID,
			ds.UserDBField:   d.Domain.GetUserID(),
		}, false); err == nil && len(dataAccess) > 0 {
		for _, access := range dataAccess {
			if !slices.Contains(ids, utils.ToString(access[utils.RootDestTableIDParam])) && utils.ToString(access[utils.RootDestTableIDParam]) != "" {
				ids = append(ids, utils.ToString(access[utils.RootDestTableIDParam]))
			}
		}
	}
	return ids
}

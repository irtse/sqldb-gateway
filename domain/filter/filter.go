package filter

import (
	"fmt"
	"net/url"
	"slices"
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

func (f *FilterService) GetQueryFilter(tableName string, innerRestriction ...string) (string, string, string, string) {
	schema, err := sch.GetSchema(tableName)
	if err != nil {
		return "", "", "", ""
	}
	var SQLview, SQLrestriction, SQLOrder []string = []string{}, []string{}, []string{}
	var SQLLimit string
	restr, view, order, dir, state := f.GetFilterForQuery("", "", fmt.Sprintf("%v", schema.ID))
	if restr != "" && !f.Domain.IsSuperCall() {
		SQLrestriction = append(SQLrestriction, restr)
	}
	later := []string{}
	for _, restr := range innerRestriction {
		if strings.Contains(restr, " IN ") {
			later = append(later, restr)
			continue
		}
		r := []string{restr}
		r = append(r, SQLrestriction...)
		SQLrestriction = r
	}
	f.Domain.GetParams().Add(utils.RootColumnsParam, view, func(v string) bool { return !f.Domain.IsSuperCall() })
	f.Domain.GetParams().Add(utils.RootOrderParam, order, func(v string) bool { return true })
	f.Domain.GetParams().Add(utils.RootDirParam, dir, func(v string) bool { return true })

	SQLrestriction = f.restrictionBySchema(tableName, SQLrestriction)
	SQLOrder = f.Domain.GetParams().GetOrder(func(el string) bool { return schema.HasField(el) }, SQLOrder)
	SQLLimit = f.Domain.GetParams().GetLimit(SQLLimit)
	SQLview = f.viewbyFields(schema)

	if f.Domain.IsSuperCall() {
		return strings.Join(SQLrestriction, ","), strings.Join(SQLview, ","), strings.Join(SQLOrder, ","), SQLLimit
	}
	SQLrestriction = f.restrictionByEntityUser(schema, SQLrestriction) // admin can see all on admin view
	if s, ok := f.Domain.GetParams()[utils.RootFilterNewState]; ok && s != "" {
		state = s
	}
	SQLrestriction = append(SQLrestriction, later...)
	if state != "" {
		SQLrestriction = f.LifeCycleRestriction(tableName, SQLrestriction, state)
	}
	return strings.Join(SQLrestriction, ","), strings.Join(SQLview, ","), strings.Join(SQLOrder, ","), SQLLimit
}

func (d *FilterService) restrictionBySchema(tableName string, restr []string) []string {
	restriction := map[string]interface{}{}
	restriction["active"] = true
	if schema, err := sch.GetSchema(tableName); err == nil {
		if schema.HasField("is_meta") && !d.Domain.IsSuperCall() {
			restriction["is_meta"] = false
		}
		alterRestr := []string{}
		if line, ok := d.Domain.GetParams()[utils.RootFilterLine]; ok && tableName != ds.DBView.Name {
			alterRestr = append(alterRestr, connector.FormatSQLRestrictionWhereInjection(line, schema.GetTypeAndLinkForField))
		}
		for key, val := range d.Domain.GetParams() {
			typ, foreign, err := schema.GetTypeAndLinkForField(key)
			if (err != nil && key != utils.SpecialIDParam) || key != utils.SpecialIDParam && tableName == ds.DBView.Name {
				continue
			}
			ands := strings.Split(fmt.Sprintf("%v", val), ",")
			for _, and := range ands {
				alterRestr = append(alterRestr, connector.MakeSqlItem("", typ, foreign, key, and, "="))
			}
		}
		restr = append(alterRestr, restr...)
		if schema.HasField(ds.SchemaDBField) && !d.Domain.IsSuperCall() {
			except := []string{ds.DBRequest.Name, ds.DBTask.Name}
			enum := []string{}
			for _, s := range sm.SchemaRegistry {
				notOK := !d.Domain.IsSuperAdmin() && ds.IsRootDB(s.Name) && !slices.Contains(except, s.Name)
				notOK2 := !s.HasField("name") || !d.Domain.VerifyAuth(s.Name, "", sm.LEVELNORMAL, utils.SELECT)
				if !notOK && !notOK2 {
					enum = append(enum, fmt.Sprintf("%v", s.ID))
				}
			}
			restr = append(restr, connector.FormatSQLRestrictionWhereByMap(
				"", map[string]interface{}{ds.SchemaDBField: enum}, false))
		}
	}
	restr = append(restr, connector.FormatSQLRestrictionWhereByMap("", restriction, false))
	return restr
}

func (s *FilterService) restrictionByEntityUser(schema sm.SchemaModel, restr []string) []string {
	newRestr := map[string]interface{}{}
	if schema.HasField(ds.UserDBField) || schema.HasField(ds.EntityDBField) {
		if !s.Domain.IsOwn(true, false, s.Domain.GetMethod()) {
			return restr
		}
	} else if s.Domain.IsOwn(false, false, s.Domain.GetMethod()) {
		ids := []string{}
		if dataAccess, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBDataAccess.Name,
			map[string]interface{}{
				"write":          true,
				ds.SchemaDBField: schema.ID,
				ds.UserDBField:   s.GetUserFilterQuery("id"),
			}, false); err == nil && len(dataAccess) > 0 {
			for _, access := range dataAccess {
				if !slices.Contains(ids, fmt.Sprintf("%v", access[utils.RootDestTableIDParam])) {
					ids = append(ids, fmt.Sprintf("%v", access[utils.RootDestTableIDParam]))
				}
			}
			if len(ids) > 0 {
				newRestr[utils.SpecialIDParam] = ids
			}
		}
		if ds.IsRootDB(s.Domain.GetTable()) && len(ids) == 0 {
			newRestr[utils.SpecialIDParam] = nil
			if len(newRestr) > 0 {
				restr = append(restr, "("+connector.FormatSQLRestrictionWhereByMap("", newRestr, false)+")")
			}
			return restr
		}
	}
	isUser := false
	restrictions := map[string]interface{}{}
	isUser = schema.HasField(ds.UserDBField) || s.Domain.GetTable() == ds.DBUser.Name
	if isUser {
		key := ds.UserDBField
		if s.Domain.GetTable() == ds.DBUser.Name {
			key = utils.SpecialIDParam
		}
		restrictions[key] = s.GetUserFilterQuery(ds.UserDBField)
	}
	if schema.HasField(ds.EntityDBField) || s.Domain.GetTable() == ds.DBEntity.Name {
		key := ds.EntityDBField
		if s.Domain.GetTable() == ds.DBEntity.Name {
			key = utils.SpecialIDParam
		}
		restrictions[key] = s.GetEntityFilterQuery(ds.EntityDBField)
	}
	if len(newRestr) > 0 {
		restr = append(restr, "("+connector.FormatSQLRestrictionWhereByMap("", newRestr, isUser)+")")
	}
	return restr
}

func (d *FilterService) viewbyFields(schema sm.SchemaModel) []string {
	SQLview := []string{}
	views := d.Domain.GetParams()[utils.RootColumnsParam]

	SQLview = append(SQLview, "id")
	for _, field := range schema.Fields {
		manyOK := field.Type == sm.MANYTOMANY.String() || field.Type == sm.ONETOMANY.String()
		if len(views) > 0 && !strings.Contains(views, field.Name) || manyOK {
			continue
		}
		if d.Domain.VerifyAuth(d.Domain.GetTable(), field.Name, field.Level, utils.SELECT) {
			SQLview = append(SQLview, field.Name)
		}
	}
	if p, ok := d.Domain.GetParams()[utils.RootCommandRow]; ok {
		decodedLine, err := url.QueryUnescape(p)
		if err == nil {
			SQLview = append(SQLview, decodedLine)
		}
	}
	return SQLview
}

func (s *FilterService) GetFilterForQuery(filterID string, viewfilterID string, schemaID string) (string, string, string, string, string) {
	ids := s.GetFilterIDs(filterID, viewfilterID, schemaID)
	view, order, dir := s.ProcessViewAndOrder(ids[utils.RootViewFilter], schemaID)
	filter := s.ProcessFilterRestriction(ids[utils.RootFilter], schemaID)
	state := ""
	if fils, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilter.Name,
		map[string]interface{}{utils.SpecialIDParam: filterID}, false); err == nil && len(fils) > 0 {
		state = fmt.Sprintf("%v", fils[0]["elder"]) // get elder filter
	}
	return filter, view, order, dir, state
}

func (s *FilterService) ProcessFilterRestriction(filterID string, schemaID string) string {
	if schemaID != "" || filterID != "" {
		return ""
	}
	var filter string
	restriction := map[string]interface{}{
		ds.SchemaFieldDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(
			ds.DBSchemaField.Name,
			map[string]interface{}{ds.SchemaDBField: schemaID}, false),
		ds.FilterDBField: filterID,
	}
	fields, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilterField.Name, restriction, false)
	if err == nil && len(fields) > 0 {
		for _, field := range fields {
			if f, err := sch.GetFieldByID(utils.GetInt(field, ds.SchemaFieldDBField)); err == nil {
				filter += connector.FormatOperatorSQLRestriction(field["operator"], field["separator"], f.Name, field["value"], f.Type)
			}
		}
	}
	return filter
}

func (s *FilterService) ProcessViewAndOrder(viewfilterID string, schemaID string) (string, string, string) {
	var viewFilter, order, dir []string = []string{}, []string{}, []string{}
	for _, field := range s.GetFilterFields(viewfilterID, schemaID) {
		f, err := sch.GetFieldByID(utils.GetInt(field, ds.RootID(ds.DBSchemaField.Name)))
		cols, ok := s.Domain.GetParams()[utils.RootColumnsParam]
		if err == nil && !slices.Contains(viewFilter, f.Name) && ok && strings.Contains(cols, f.Name) {
			viewFilter = append(viewFilter, f.Name)
			if field["dir"] != nil {
				dir = append(dir, strings.ToUpper(fmt.Sprintf("%v", field["dir"])))
				order = append(order, f.Name)
			} else {
				dir = append(dir, "ASC")
				order = append(order, f.Name)
			}
		}
	}
	return strings.Join(viewFilter, ","), strings.Join(order, ","), strings.Join(dir, ",")
}

func (d *FilterService) LifeCycleRestriction(tableName string, SQLrestriction []string, state string) []string {
	if state == "all" || tableName == ds.DBView.Name {
		return SQLrestriction
	}
	var operator string
	news, _ := d.CountNewDataAccess(tableName, SQLrestriction, map[string]interface{}{})
	if state == "new" {
		operator = "IN"
		if len(news) == 0 {
			news = append(news, "NULL")
		}
	}
	if state == "old" {
		if len(news) == 0 {
			return SQLrestriction
		}
		operator = "NOT IN"
	}
	if operator != "" {
		t := "id " + operator + " (" + strings.Join(news, ",") + ")"
		SQLrestriction = append(SQLrestriction, t)
	}
	return SQLrestriction
}

// set up in DB SIDE

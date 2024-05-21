package domain

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"strconv"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// define filter whatever what happen on sql...
func (d *MainService) ViewDefinition(tableName string, innerRestriction... string) (string, string, string, string) {
	SQLview := ""; SQLrestriction := ""; SQLOrder := ""; SQLLimit := ""
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return SQLrestriction, SQLview, SQLOrder, SQLLimit }
	restr, view, order, dir := d.GetFilter("", "", fmt.Sprintf("%v", schema.ID))
	if restr != "" { innerRestriction = append(innerRestriction, restr) }
	if view != "" { d.Params[utils.RootColumnsParam] = view }
	if order != "" { d.Params[utils.RootOrderParam] = order }
	if dir != "" { d.Params[utils.RootDirParam] = dir }
	SQLrestriction = d.restrictionBySchema(tableName, SQLrestriction)
	SQLOrder = d.orderFromParams(tableName, SQLOrder)
	SQLLimit = d.limitFromParams(SQLLimit)
	SQLview = d.viewbyFields(tableName)
	for _, restr := range innerRestriction {
		if len(strings.TrimSpace(restr)) == 0 { continue }
		if len(SQLrestriction) > 0 { SQLrestriction += " AND (" + restr + ")" } else { SQLrestriction = restr  }
	}
	if !d.IsSuperCall() { SQLrestriction = d.restrictionByEntityUser(tableName, SQLrestriction) } // admin can see all on admin view
	return SQLrestriction, SQLview, SQLOrder, SQLLimit
}
func (d *MainService) restrictionBySchema(tableName string, restr string) (string) {
	if len(restr) > 0 { restr +=  " AND " }
	restr += "active=true"
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return restr }
	if schema.HasField("is_meta") && !d.IsSuperCall() { 
		if len(restr) > 0 { restr +=  " AND " }
		restr += "is_meta=false"
	}
	already := []string{}
	isSchema := false
	isOr := false
	alterRestr := ""
    for key, element := range d.Params {
		field, err := schema.GetField(key)
		if (err != nil && key != utils.SpecialIDParam) || slices.Contains(already, key) || key != utils.SpecialIDParam && tableName == schserv.DBView.Name { 
			continue 
		}
		if strings.Contains(key, schserv.RootID(schserv.DBSchema.Name)) { isSchema = true }
		already = append(already, key)
		if strings.Contains(element, ",") { 
			els := ""
			for _, el := range strings.Split(element, ",") { els += conn.FormatForSQL(field.Type, conn.SQLInjectionProtector(el)) + "," }
			if len(alterRestr) > 0 { 
				if isOr { alterRestr +=  " OR " } else { alterRestr +=  " AND " }
			}
			alterRestr += key + " IN (" + conn.RemoveLastChar(els) + ")"
		} else { 
			if len(element) > 2 &&  element[len(element) - 3:] == "%7C"{ isOr = true } else { isOr = false }
			ands := strings.Split(element, "+")
			for _, and := range ands {
				ors := strings.Split(and, "%7C")
				if len(ors) == 0 { continue }
				if len(alterRestr) > 0 { 
					if isOr { alterRestr +=  " OR " } else { alterRestr +=  " AND " }
				}
				alterRestr += "("
				count := 0
				for _, or := range ors {
					sql := strings.ReplaceAll(field.Type, "%25", "%")
					sql = conn.FormatForSQL(sql, or)
					if count > 0 { alterRestr +=  " OR " }
					if field.Link > 0 {
						foreign, _ := schserv.GetSchemaByID(field.Link)
						if strings.Contains(sql, "%") { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE name::text LIKE " + sql + " OR id::text LIKE " + sql + ")"
						} else { 
							if strings.Contains(sql, "'") { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE name = " + sql + ")" 
							} else { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE id = " + sql + ")" }
						}
					} else if strings.Contains(sql, "%") { alterRestr += key + "::text LIKE " + sql } else { alterRestr += key + "=" + sql }
					count++
				}
				alterRestr += ")"
			}
		}
	}
	if len(alterRestr) > 0 { if len(restr) > 0 { restr +=  " AND (" + alterRestr + ")"  } else { restr = alterRestr } }
	if schema.HasField(schserv.RootID(schserv.DBSchema.Name)) && !d.IsSuperCall() && !isSchema { 
		restr += " AND " + schserv.RootID(schserv.DBSchema.Name) + " IN (" 
		for _, sch := range schserv.SchemaRegistry {
			if sch.Name == schserv.DBWorkflow.Name || !d.PermsCheck(sch.Name, "", schserv.LEVELNORMAL, utils.SELECT) { continue }
			restr += fmt.Sprintf("%v", sch.ID) + ","
		}
		restr = conn.RemoveLastChar(restr) + ")"
	}
	return restr
}

func (d *MainService) limitFromParams(limited string) (string) {
	if limit, ok := d.Params[utils.RootLimit]; ok {
		i, err := strconv.Atoi(limit)
		if err == nil { 
			limited = "LIMIT " + fmt.Sprintf("%v", i)
			if offset, ok := d.Params["offset"]; ok {
				i2, err := strconv.Atoi(offset)
				if err == nil { limited += " OFFSET " + fmt.Sprintf("%v", i2) }
			}
		}
	}
	return limited
}

func (d *MainService) orderFromParams(tableName string, order string) (string) {
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return "id DESC" }
	if orderBy, ok := d.Params[utils.RootOrderParam]; ok {
		direction := []string{}
		if dir, ok2 := d.Params[utils.RootDirParam]; ok2 { direction = strings.Split(fmt.Sprintf("%v", dir), ",") }
		for i, el := range strings.Split(fmt.Sprintf("%v", orderBy), ",") {
			if (!schema.HasField(el) && el != utils.SpecialIDParam) || len(direction) <= i  { continue } // ???
			upper := strings.Replace(strings.ToUpper(direction[i]), " ", "", -1)
			if upper == "ASC" || upper == "DESC" { order += conn.SQLInjectionProtector(el + " " + upper + ","); continue }
			order += conn.SQLInjectionProtector(el + " ASC,") 
		}
		order = conn.RemoveLastChar(order)
	} else { return "id DESC" }
	return order
}

func (s *MainService) restrictionByEntityUser(tableName string, restr string) string {
	schema, err := schserv.GetSchema(tableName)
	if !s.IsOwnPermission(tableName, false, s.Method) || err != nil { return restr }
	userID := schserv.RootID(schserv.DBUser.Name); entityID := schserv.RootID(schserv.DBEntity.Name)
	isUser := false
	if schema.HasField(userID) || tableName == schserv.DBUser.Name {
		if len(restr) > 0 { restr +=  " AND " }
		isUser = true
		if tableName == schserv.DBUser.Name  { userID = utils.SpecialIDParam }
		restr += userID + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" 
		restr += conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")" 
	}
	if schema.HasField(entityID) || tableName == schserv.DBEntity.Name  {
		if tableName == schserv.DBEntity.Name  { entityID = utils.SpecialIDParam }
		if isUser { restr +=  " OR " } else if len(restr) > 0 { restr +=  " AND " }
		restr += entityID + " IN (SELECT " + entityID + " FROM " + schserv.DBEntityUser.Name + " WHERE " + schserv.RootID(schserv.DBUser.Name) + " IN (" 
		restr += "SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + "))"
		// TODO GET FROM PARENT ID MISSING + OWN
	}
	return restr
}

func (d *MainService) viewbyFields(tableName string) (string) {
	SQLview := "id,"
	views := d.Params[utils.RootColumnsParam]
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return "" }
	for _, field := range schema.Fields {
		if len(views) > 0 && !strings.Contains(views, field.Name) || field.Type == schserv.MANYTOMANY.String() || field.Type == schserv.ONETOMANY.String() { continue }
		if d.PermsCheck(tableName, field.Name, field.Level, utils.SELECT) { SQLview += field.Name + "," }
	}
	if len(SQLview) > 0 { SQLview = SQLview[:len(SQLview) - 1] }
	return SQLview
}
func (s *MainService) GetFilter(filterID string, viewfilterID string, schemaID string) (string, string, string, string) {
	viewFilter := ""; filter := ""; order := ""; dir := ""
	params := utils.AllParams(schserv.DBFilter.Name)
	params[schserv.RootID(schserv.DBSchema.Name)] = schemaID
	utils.ParamsMutex.Lock()
	if s.GetParams()[utils.RootFilter] != "" { 
		params[schserv.RootID(schserv.DBFilter.Name)] = s.GetParams()[utils.RootFilter]
		fields, err := s.PermsSuperCall(params, utils.Record{}, utils.SELECT)
		if len(fields) > 0 && err == nil { filterID = s.GetParams()[utils.RootFilter] }
	}
	if s.GetParams()[utils.RootViewFilter] != "" { 
		params[schserv.RootID(schserv.DBFilter.Name)] = s.GetParams()[utils.RootViewFilter]
		fields, err := s.PermsSuperCall(params, utils.Record{}, utils.SELECT)
		if len(fields) > 0 && err == nil { viewfilterID = s.GetParams()[utils.RootViewFilter] }
	}
	utils.ParamsMutex.Unlock()
	sqlFilter := "SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " // TODO VIEW FILTER FOR MAIN PURPOSE...
	if schemaID != "" { sqlFilter += schserv.RootID(schserv.DBSchemaField.Name) + " IN (SELECT id FROM " + schserv.DBSchemaField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + " = " + schemaID + " ) AND " }
	if viewfilterID == "" { sqlFilter += schserv.RootID(schserv.DBFilter.Name) + " IN (SELECT id FROM " + schserv.DBFilter.Name + " WHERE " + schserv.RootID(schserv.DBUser.Name) + " IS NULL AND " + schserv.RootID(schserv.DBEntity.Name) + " IS NULL)" 
	} else { sqlFilter += schserv.RootID(schserv.DBFilter.Name) + "=" + viewfilterID } 
	fields, err := s.Db.QueryAssociativeArray(sqlFilter)
	if err == nil && len(fields) > 0 {
		sort.SliceStable(fields, func(i, j int) bool{ return fields[i]["index"].(int64) <= fields[j]["index"].(int64) })
		for _, field := range fields {
			f, err := schserv.GetFieldByID(utils.GetInt(field, schserv.RootID(schserv.DBSchemaField.Name)))
			if err != nil || strings.Contains(viewFilter, f.Name) || !strings.Contains(s.Params[utils.RootColumnsParam], f.Name) { continue }
			viewFilter += f.Name + ","
			if field["dir"] != nil { dir += strings.ToUpper(fmt.Sprintf("%v", field["dir"])) + ","; order += f.Name + ","
			} else { dir += "ASC,"; order += f.Name + "," }
		}
		if len(viewFilter) > 0 { viewFilter = viewFilter[:len(viewFilter)-1] }
		if len(order) > 0 { order = order[:len(order)-1] }
		if len(dir) > 0 { dir = dir[:len(dir)-1] }
	}
	sqlFilter = "SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " // TODO VIEW FILTER FOR MAIN PURPOSE...
	if schemaID != "" { sqlFilter += schserv.RootID(schserv.DBSchemaField.Name) + " IN  (SELECT id FROM " + schserv.DBSchemaField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + " = " + schemaID + " ) AND " }
	if filterID == "" { sqlFilter += schserv.RootID(schserv.DBUser.Name) + " IS NULL AND " + schserv.RootID(schserv.DBEntity.Name) + " IS NULL" 
	} else { sqlFilter += schserv.RootID(schserv.DBFilter.Name) + "=" + filterID } 
	fields, err = s.Db.QueryAssociativeArray(sqlFilter)
	if err == nil && len(fields) > 0 {
		for _, field := range fields {
			f, err := schserv.GetFieldByID(utils.GetInt(field, schserv.RootID(schserv.DBSchemaField.Name)))
			if err != nil || field["operator"] == nil || field["separator"] == nil { continue }
			if len(filter) > 0 { filter += " " + fmt.Sprintf("%v", field["separator"]) + " " }
			if fmt.Sprintf("%v", field["operator"]) == "LIKE" {
				filter += f.Name + "::text " + fmt.Sprintf("%v", field["operator"]) + " '%" + fmt.Sprintf("%v", field["value"]) + "%'"
			}
			filter += f.Name + " " + fmt.Sprintf("%v", field["operator"]) + " " + conn.FormatForSQL(f.Type, field["value"])
		}
	}
	return filter, viewFilter, order, dir
}

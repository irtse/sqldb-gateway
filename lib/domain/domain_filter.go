package domain

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"strconv"
	"net/url"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
)
func (d *MainService) LifeCycleRestriction(tableName string, restr string, state string) string {
	if state == "all" || tableName == schserv.DBView.Name { return restr }
	operator := ""
	news, _ := d.CountNewDataAccess(tableName, restr, utils.Params{})
	if state == "new" { 
		operator = "IN" 
		if len(news) == 0 { news = append(news, "NULL")}
	}
	if state == "old" { 
		if len(news) == 0 { return restr }
		operator = "NOT IN" 
	}
	if operator != "" { 
		t := "id " + operator + " (" + strings.Join(news, ",") + ")" 
		if len(restr) > 0 { restr += " AND " }
		restr = restr + t
	}
	return restr
}
// define filter whatever what happen on sql...
func (d *MainService) ViewDefinition(tableName string, innerRestriction... string) (string, string, string, string) {
	SQLview := ""; SQLrestriction := ""; SQLOrder := ""; SQLLimit := ""
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return SQLrestriction, SQLview, SQLOrder, SQLLimit }
	
	restr, view, order, dir, state := d.GetFilter("", "", fmt.Sprintf("%v", schema.ID))
	if restr != "" && !d.IsSuperCall() { 
		if len(SQLrestriction) > 0 { SQLrestriction += " AND " }
		SQLrestriction += restr
	}
	later := []string{}
	for _, restr := range innerRestriction {
		if strings.Contains(restr, " IN ") { later = append(later, restr); continue }
		if len(SQLrestriction) > 0  && len(restr) > 0 { SQLrestriction = restr + " AND " + SQLrestriction } else { SQLrestriction = restr  }
	}
	if view != "" && !d.IsSuperCall() { d.Params[utils.RootColumnsParam] = view }
	if order != "" { d.Params[utils.RootOrderParam] = order }
	if dir != "" { d.Params[utils.RootDirParam] = dir }
	SQLrestriction = d.restrictionBySchema(tableName, SQLrestriction)
	SQLOrder = d.orderFromParams(tableName, SQLOrder)
	SQLLimit = d.limitFromParams(SQLLimit)
	SQLview = d.viewbyFields(tableName)
	
	if d.IsSuperCall() { return SQLrestriction, SQLview, SQLOrder, SQLLimit }
	SQLrestriction = d.restrictionByEntityUser(tableName, SQLrestriction) // admin can see all on admin view
	if s, ok := d.Params[utils.RootFilterNewState]; ok && s != "" { state = s }
	for _, restr := range later {
		if len(SQLrestriction) > 0  && len(restr) > 0 { SQLrestriction = SQLrestriction + " AND " + restr } else { SQLrestriction = restr  }
	}
	if state != "" { SQLrestriction = d.LifeCycleRestriction(tableName, SQLrestriction, state) }
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
	alterRestr := ""
	if line, ok := d.Params[utils.RootFilterLine]; ok && tableName != schserv.DBView.Name {
		decodedLine, err := url.QueryUnescape(fmt.Sprint(line))
		if err == nil {
			ands := strings.Split(decodedLine, "+")
			// todo order depending on the field index
			for _, and := range ands {
				if len(strings.Trim(alterRestr, " ")) > 0 { alterRestr +=  " AND " }
				ors := strings.Split(and, "|")
				if len(ors) == 0 { continue }
				orRestr := ""
				for _, or := range ors {
					operator := "~"
					keyVal := []string{} 
					if strings.Contains(or, "<>~") { keyVal = strings.Split(or, "<>~"); operator = " NOT LIKE "
					} else if strings.Contains(or, "~") { keyVal = strings.Split(or, "~"); operator = " LIKE " 
					} else if strings.Contains(or, "<>") { keyVal = strings.Split(or, "<>"); operator = "<>"
					} else if strings.Contains(or, "<:") { keyVal = strings.Split(or, "<:"); operator = "<=" 
					} else if strings.Contains(or, ">:") { keyVal = strings.Split(or, ">:"); operator = ">=" 
					} else if strings.Contains(or, ":") { keyVal = strings.Split(or, ":"); operator = "="
					} else if strings.Contains(or, "<") { keyVal = strings.Split(or, "<"); operator = "<"  
					} else if strings.Contains(or, ">") { keyVal = strings.Split(or, ">"); operator = ">"  }
					if len(keyVal) != 2 { continue }
					field, err := schema.GetField(keyVal[0])
					if (err != nil && keyVal[0] != utils.SpecialIDParam) { continue  }
					if len(strings.Trim(orRestr, " ")) > 0 { orRestr +=  " OR " }
					orRestr = d.sqlItem(orRestr, field, keyVal[0], keyVal[1], operator)
				}
				if len(orRestr) > 0 { alterRestr += "( " + orRestr + " )" }
			}
		}
	}
	alterRestr = strings.ReplaceAll(strings.ReplaceAll(alterRestr, " OR ()", ""), " AND ()", "")
	alterRestr = strings.ReplaceAll(alterRestr, "()", "")
	for key, val := range d.Params {
		field, err := schema.GetField(key)
		if (err != nil && key != utils.SpecialIDParam) || slices.Contains(already, key) || key != utils.SpecialIDParam && tableName == schserv.DBView.Name { continue  }
		ands := strings.Split(fmt.Sprintf("%v", val), ",")
		for _, and := range ands {
			if len(strings.Trim(alterRestr, " ")) > 0 { alterRestr +=  " AND " }
			alterRestr = d.sqlItem(alterRestr, field, key, and, "=")
		}
	}
	if len(alterRestr) > 0 { 
		if len(restr) > 0 { restr = alterRestr + " AND " + restr } else { restr = alterRestr } 
	}
	if schema.HasField(schserv.RootID(schserv.DBSchema.Name)) && !d.IsSuperCall() && !isSchema { 
		except := []string{schserv.DBRequest.Name, schserv.DBTask.Name}
		restr += " AND " + schserv.RootID(schserv.DBSchema.Name) + " IN (" 
		for _, sch := range schserv.SchemaRegistry {
			if (!d.IsSuperAdmin() &&  sch.Name[:2] == "db" && !slices.Contains(except, sch.Name)) || !sch.HasField("name") || !d.PermsCheck(sch.Name, "", schserv.LEVELNORMAL, utils.SELECT) { continue }
			restr += fmt.Sprintf("%v", sch.ID) + ","
		}
		restr = conn.RemoveLastChar(restr) + ")"
	}
	return restr
}

func (d *MainService) sqlItem(alterRestr string, field schserv.FieldModel, key string, or string, operator string) (string) {
	sql := or
	sql = conn.FormatForSQL(field.Type, sql)
	if sql == "" { return alterRestr }
	if strings.Contains(sql, "NULL") { operator = "IS " }
	if field.Link > 0 {
		foreign, _ := schserv.GetSchemaByID(field.Link)
		if strings.Contains(sql, "%") { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE name::text LIKE " + sql + " OR id::text " + operator + sql + ")"
		} else { 			
			if strings.Contains(sql, "'") {  
				if strings.Contains(sql, "NULL") { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE name IS " + sql + ")"  
				} else { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE name = " + sql + ")" }
			} else { alterRestr += key + " IN (SELECT id FROM " + foreign.Name + " WHERE id " + operator + " " + sql + ")" }
		}
	} else if strings.Contains(sql, "%") { alterRestr += key + "::text " + operator + sql } else { alterRestr += key + " " + operator + " " + sql }
	return alterRestr
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
	if err != nil { return restr }
	newRestr := ""
	userID := schserv.RootID(schserv.DBUser.Name); entityID := schserv.RootID(schserv.DBEntity.Name)
	if (schema.HasField(userID) || schema.HasField(entityID)) {
		if !s.IsOwnPermission(tableName, false, s.Method) && !s.IsOwn() { return restr }
	} else if s.IsOwn() {
		quer := "SELECT * FROM " + schserv.DBRequest.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID) + " AND " 
		quer += userID + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")" 
		requests, err := s.Db.QueryAssociativeArray(quer)
		ids := ""; 
		if err == nil && len(requests) > 0 {
			for _, request := range requests { 
				if !strings.Contains(ids, fmt.Sprintf("%v", request[utils.RootDestTableIDParam])) {
					ids += fmt.Sprintf("%v", request[utils.RootDestTableIDParam]) + "," 
				}	
			}
			if len(ids) > 0 { 
				ids = conn.RemoveLastChar(ids) 
				if len(newRestr) > 0 { newRestr +=  " AND " }
				newRestr += "id IN (" + ids + ")"
			}
		} 
		if tableName[:2] != "db" && len(ids) == 0 { 
			if len(newRestr) > 0 { newRestr +=  " AND " }
			newRestr += "id IS NULL"
			if len(newRestr) > 0 {
				if len(restr) > 0 { restr +=  " AND " }
				restr += "(" + newRestr + ")"
			}
			return restr 
		}
	}
	isUser := false
	if schema.HasField(userID) || tableName == schserv.DBUser.Name {
		if len(newRestr) > 0 { newRestr +=  " AND " }
		isUser = true
		if tableName == schserv.DBUser.Name  { userID = utils.SpecialIDParam }
		newRestr += userID + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" 
		newRestr += conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + ")" 
	}
	if schema.HasField(entityID) || tableName == schserv.DBEntity.Name  {
		if tableName == schserv.DBEntity.Name  { entityID = utils.SpecialIDParam }
		if isUser { newRestr +=  " OR " } else if len(newRestr) > 0 { newRestr +=  " AND " }
		newRestr += entityID + " IN (SELECT " + entityID + " FROM " + schserv.DBEntityUser.Name + " WHERE " + schserv.RootID(schserv.DBUser.Name) + " IN (" 
		newRestr += "SELECT id FROM " + schserv.DBUser.Name + " WHERE name=" + conn.Quote(s.GetUser()) + " OR email=" + conn.Quote(s.GetUser()) + "))"
		// TODO GET FROM PARENT ID MISSING + OWN
	}
	if len(newRestr) > 0 {
		if len(restr) > 0 { restr +=  " AND " }
		restr += "(" + newRestr + ")"
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
	if d.Params[utils.RootCommandRow] != "" { 
		decodedLine, err := url.QueryUnescape(fmt.Sprint(d.Params[utils.RootCommandRow]))
		if err == nil { SQLview += "," + decodedLine }
	}
	return SQLview
}
func (s *MainService) GetFilter(filterID string, viewfilterID string, schemaID string) (string, string, string, string, string) {
	viewFilter := ""; filter := ""; order := ""; dir := ""
	params := utils.AllParams(schserv.DBFilter.Name)
	params[schserv.RootID(schserv.DBSchema.Name)] = schemaID
	utils.ParamsMutex.Lock()
	if s.GetParams()[utils.RootFilter] != "" { 
		params[schserv.RootID(schserv.DBFilter.Name)] = s.GetParams()[utils.RootFilter]
		fields, err := s.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + schemaID + " AND " + schserv.RootID(schserv.DBFilter.Name) + "=" + s.GetParams()[utils.RootFilter])
		if len(fields) > 0 && err == nil { filterID = s.GetParams()[utils.RootFilter] }
	}
	if s.GetParams()[utils.RootViewFilter] != "" { 
		params[schserv.RootID(schserv.DBFilter.Name)] = s.GetParams()[utils.RootViewFilter]
		params["is_view"] = "true"
		fields, err := s.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + schemaID + " AND " + schserv.RootID(schserv.DBFilter.Name) + "=" + s.GetParams()[utils.RootViewFilter] + " AND is_view=true")
		if len(fields) > 0 && err == nil { viewfilterID = s.GetParams()[utils.RootViewFilter] }
	}
	utils.ParamsMutex.Unlock()
	sqlFilter := "SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " // TODO VIEW FILTER FOR MAIN PURPOSE...
	if schemaID != "" { sqlFilter += schserv.RootID(schserv.DBSchemaField.Name) + " IN (SELECT id FROM " + schserv.DBSchemaField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + " = " + schemaID + " ) AND " }
	if viewfilterID != "" { 
		sqlFilter += schserv.RootID(schserv.DBFilter.Name) + "=" + viewfilterID 
		fields, err := s.Db.QueryAssociativeArray(sqlFilter)
		if err == nil && len(fields) > 0 {
			sort.SliceStable(fields, func(i, j int) bool{ return fields[i]["index"].(int64) <= fields[j]["index"].(int64) })
			for _, field := range fields {
				f, err := schserv.GetFieldByID(utils.GetInt(field, schserv.RootID(schserv.DBSchemaField.Name)))
				if err != nil || strings.Contains(viewFilter, f.Name) || (len(s.Params[utils.RootColumnsParam]) > 0 && !strings.Contains(s.Params[utils.RootColumnsParam], f.Name)) { continue }
				viewFilter += f.Name + ","
				if field["dir"] != nil { dir += strings.ToUpper(fmt.Sprintf("%v", field["dir"])) + ","; order += f.Name + ","
				} else { dir += "ASC,"; order += f.Name + "," }
			}
			if len(viewFilter) > 0 { viewFilter = viewFilter[:len(viewFilter)-1] }
			if len(order) > 0 { order = order[:len(order)-1] }
			if len(dir) > 0 { dir = dir[:len(dir)-1] }
		}
	}
	sqlFilter = "SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " // TODO VIEW FILTER FOR MAIN PURPOSE...
	state := ""
	if schemaID != "" { 
		sqlFilter += schserv.RootID(schserv.DBSchemaField.Name) + " IN  (SELECT id FROM " + schserv.DBSchemaField.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + " = " + schemaID + " ) AND " 
		if filterID != "" { 
			sqlFilter += schserv.RootID(schserv.DBFilter.Name) + "=" + filterID 
			fields, err := s.Db.QueryAssociativeArray(sqlFilter)
			if err == nil && len(fields) > 0 {
				for _, field := range fields {
					f, err := schserv.GetFieldByID(utils.GetInt(field, schserv.RootID(schserv.DBSchemaField.Name)))
					if err != nil || field["operator"] == nil || field["separator"] == nil { continue }
					if len(filter) > 0 { 
						filter += " " + fmt.Sprintf("%v", field["separator"]) + " " 
					}
					if fmt.Sprintf("%v", field["operator"]) == "LIKE" {
						filter += f.Name + "::text " + fmt.Sprintf("%v", field["operator"]) + " '%" + fmt.Sprintf("%v", field["value"]) + "%'"
					} else { 
						filter += f.Name + " " + fmt.Sprintf("%v", field["operator"]) + " " + conn.FormatForSQL(f.Type, field["value"]) 
					}
				}
			}
			/*sqlFilter = "SELECT elder FROM " + schserv.DBFilter.Name + " WHERE id=" + filterID 
			fils, err := s.Db.QueryAssociativeArray(sqlFilter)
			if err == nil && len(fils) > 0 { state = fmt.Sprintf("%v", fils[0]["elder"]) }*/
		}
	}
	return filter, viewFilter, order, dir, state
}

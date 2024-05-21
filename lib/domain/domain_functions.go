package domain

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"encoding/json"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)
// DONT FORGET DATA ACCESS
func (d *MainService) CountNewDataAccess(tableName string, filter string, countParams utils.Params) ([]string, int64) {
	sqlFilter := "id NOT IN (SELECT " + schserv.RootID("dest_table") + " FROM " + schserv.DBDataAccess.Name + " WHERE "
	sqlFilter += schserv.RootID(schserv.DBSchema.Name) + " IN (SELECT id FROM " + schserv.DBSchema.Name + " WHERE name = '" + tableName + "') AND " 
	sqlFilter += schserv.RootID(schserv.DBUser.Name) + " IN (SELECT id FROM " + schserv.DBUser.Name + " WHERE name = '" + d.GetUser() + "' OR email='" + d.GetUser() + "') AND write=false AND update=false)"
	if len(filter) > 0 { sqlFilter += " AND " + filter }
	p := utils.Params{ utils.RootTableParam : tableName, utils.RootRowsParam : utils.ReservedParam, 
		               utils.RootColumnsParam : utils.SpecialIDParam }
	res, err := d.PermsSuperCall( p, utils.Record{}, utils.SELECT, sqlFilter)
	ids := []string{}
	if err != nil { return ids, 0 }
	for _, rec := range res { 
		if !slices.Contains(ids, rec.GetString(utils.SpecialIDParam)) { ids = append(ids, rec.GetString(utils.SpecialIDParam)) }
	}
	sqlFilter = ""
	if len(filter) > 0 { sqlFilter += " AND " + filter }
	res, err = d.SuperCall( utils.AllParams(tableName), utils.Record{}, utils.COUNT, sqlFilter)
	if len(res) == 0 || err != nil || res[0]["count"] == nil { return ids, 0 }
	return ids, int64(res[0]["count"].(float64))
}

func (d *MainService) NewDataAccess(schemaID int64, destIDs []string, meth utils.Method) {
	users, err := d.SuperCall(utils.AllParams(schserv.DBUser.Name), utils.Record{}, utils.SELECT, "name='"+ d.GetUser() + "' OR email='" + d.GetUser() + "'")
	if err == nil && len(users) > 0 {
		for _, destID := range destIDs {
			id := users[0].GetString(utils.SpecialIDParam)
			if meth == utils.DELETE {
				d.SuperCall( utils.Params{ utils.RootTableParam : schserv.DBDataAccess.Name, utils.RootRowsParam : utils.ReservedParam,
							schserv.RootID("dest_table") : destID,
							schserv.RootID(schserv.DBSchema.Name) : fmt.Sprintf("%v", schemaID),
							schserv.RootID(schserv.DBUser.Name) : id }, utils.Record{}, utils.DELETE)
			} else {
				d.SuperCall(utils.AllParams(schserv.DBDataAccess.Name), utils.Record{
						"write" : meth == utils.CREATE,
						"update" : meth == utils.UPDATE,
						schserv.RootID("dest_table") : destID,
						schserv.RootID(schserv.DBSchema.Name) : schemaID,
						schserv.RootID(schserv.DBUser.Name) : id, }, utils.CREATE)
			}
		}
	}	
}

func (d *MainService) GetViewFields(tableName string, noRecursive bool) (map[string]interface{}, int64, []string, map[string]schserv.FieldModel, []string, bool) {
	tableName = schserv.GetTablename(tableName)
	cols := map[string]schserv.FieldModel{}; schemes := map[string]interface{}{}
	keysOrdered := []string{}; additionnalAction := []string{}
	readonly := true
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return schemes, -1, keysOrdered, cols, additionnalAction, true }
	_, view, _, _ := d.GetFilter("", "", fmt.Sprintf("%v", schema.ID))
	for _, scheme := range schema.Fields {
		if (!d.SuperAdmin && !d.PermsCheck(tableName, scheme.Name, scheme.Level, utils.SELECT)) { continue }
		var shallowField schserv.ViewFieldModel
		cols[scheme.Name]=scheme		
		b, _ := json.Marshal(scheme)
		json.Unmarshal(b, &shallowField)
		shallowField.ActionPath = ""
		shallowField.Actions=[]string{}
		if scheme.Link > 0 && !d.LowerRes {
			schema, _ := schserv.GetSchemaByID(scheme.Link)
			shallowField.ActionPath = "/" + utils.MAIN_PREFIX + "/" + schema.Name + "?rows=all"
			shallowField.LinkPath = shallowField.ActionPath + "&" + utils.RootShallow + "=enable"
			if strings.Contains(scheme.Type, "many") {
				for _, field := range schema.Fields {
					if strings.Contains(field.Name, "_id") && !strings.Contains(field.Name, tableName) && field.Link > 0 {
						schField, _ := schserv.GetSchemaByID(field.Link)
						shallowField.LinkPath = "/" + utils.MAIN_PREFIX + "/" + schField.Name + "?rows=all" + "&" + utils.RootShallow + "=enable"
					}
				}
			}
		}
		for _, meth := range []utils.Method{ utils.SELECT, utils.CREATE, utils.UPDATE, utils.DELETE } {
			if d.PermsCheck(tableName, "", "", meth) && (((meth == utils.SELECT || meth == utils.CREATE) && d.Empty) || !d.Empty){ 
				if !slices.Contains(additionnalAction, meth.Method()) { additionnalAction = append(additionnalAction, meth.Method()) }
				if meth == utils.CREATE && !slices.Contains(additionnalAction, "import") {
					res, err := d.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBWorkflow.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
					if err == nil && len(res) > 0 {
						ids := ""
						for _, rec := range res { ids += fmt.Sprintf("%v", rec[utils.SpecialIDParam]) + "," }
						res, err = d.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBWorkflowSchema.Name + " WHERE " + schserv.RootID(schserv.DBWorkflow.Name) + " IN (" + ids[:len(ids) - 1] + ")")
						if len(res) == 0 { additionnalAction = append(additionnalAction, "import") }
					}
				}
			} 
			if scheme.Link > 0 && !noRecursive{
				schema, _ := schserv.GetSchemaByID(scheme.Link)
				if d.PermsCheck(schema.Name, "", "", meth) { 
					sch, _, _, _, _, _ := d.GetViewFields(schema.Name, true)
					shallowField.DataSchema = sch
					// shallowField.DataSchemaOrder = ordered
					shallowField.ActionPath = "/" + utils.MAIN_PREFIX + "/" + schema.Name + "?rows=" + utils.ReservedParam
					shallowField.Actions=append(shallowField.Actions, meth.Method())
				} 
			}
			if meth == utils.UPDATE && d.Empty { 
				readonly = false
				shallowField.Readonly = false 
			} else if meth == utils.CREATE && d.Empty { shallowField.Readonly = true } 
		} 
		if !(view != "" && !strings.Contains(view, scheme.Name)) { keysOrdered = append(keysOrdered, scheme.Name) }
		schemes[scheme.Name]=shallowField
	}
	sort.SliceStable(keysOrdered, func(i, j int) bool{
        return schemes[keysOrdered[i]].(schserv.ViewFieldModel).Index <= schemes[keysOrdered[j]].(schserv.ViewFieldModel).Index
    })
	return schemes, schema.ID, keysOrdered, cols, additionnalAction, readonly
}

func (d *MainService) BuildPath(tableName string, rows string, extra... string) string {
	path := "/" + utils.MAIN_PREFIX + "/" + tableName + "?rows=" + rows
	for _, ext := range extra { path += "&" + ext }
	return path
}
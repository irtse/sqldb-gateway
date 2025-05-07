package view_convertor

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"strings"
)

func (v *ViewConvertor) GetShortcuts(schemaID string, actions []string) map[string]string {
	shortcuts := map[string]string{}
	m := map[string]interface{}{
		"shortcut_on_schema": schemaID,
	}
	if results, err := v.Domain.GetDb().SelectQueryWithRestriction(ds.DBView.Name, m, false); err == nil {
		for _, shortcut := range results {
			if utils.GetBool(shortcut, "is_empty") {
				scheme, err := sch.GetSchemaByID(utils.ToInt64(schemaID))
				if err != nil || !v.Domain.VerifyAuth(scheme.Name, "", "", utils.CREATE) {
					continue
				}
			}
			shortcuts[utils.GetString(shortcut, sm.NAMEKEY)] = "#" + utils.GetString(shortcut, utils.SpecialIDParam)
		}
	}
	return shortcuts
}

func (d *ViewConvertor) FetchRecord(tableName string, m map[string]interface{}) []map[string]interface{} {
	t, err := d.Domain.GetDb().SelectQueryWithRestriction(tableName, m, false)
	if err != nil || len(t) == 0 {
		return nil
	}
	return t
}

func (d *ViewConvertor) NewDataAccess(schemaID int64, destIDs []string, meth utils.Method) {
	d.Domain.GetDb().ClearQueryFilter()
	if users, err := d.Domain.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
		"name":  connector.Quote(d.Domain.GetUser()),
		"email": connector.Quote(d.Domain.GetUser()),
	}, true); err == nil && len(users) > 0 {
		for _, destID := range destIDs {
			id := utils.GetString(users[0], utils.SpecialIDParam)
			d.Domain.CreateSuperCall(utils.AllParams(ds.DBDataAccess.Name),
				utils.Record{
					"write":             meth == utils.CREATE,
					"update":            meth == utils.UPDATE,
					ds.DestTableDBField: destID,
					ds.SchemaDBField:    schemaID,
					ds.UserDBField:      id})
		}
	}
}

func (d *ViewConvertor) GetViewFields(tableName string, noRecursive bool) (map[string]interface{}, int64, []string, map[string]sm.FieldModel, []string, bool) {
	tableName = sch.GetTablename(tableName)
	cols := make(map[string]sm.FieldModel)
	schemes := make(map[string]interface{})
	keysOrdered := []string{}
	additionalActions := []string{}

	schema, err := sch.GetSchema(tableName)
	if err != nil {
		return schemes, -1, keysOrdered, cols, additionalActions, true
	}
	for _, scheme := range schema.Fields {
		if !d.Domain.IsSuperAdmin() && !d.Domain.VerifyAuth(tableName, scheme.Name, scheme.Level, utils.SELECT) {
			continue
		}
		shallowField := sm.ViewFieldModel{
			ActionPath: "",
			Actions:    []string{},
		}
		cols[scheme.Name] = scheme
		b, _ := json.Marshal(scheme)
		json.Unmarshal(b, &shallowField)

		if scheme.Name == utils.RootDestTableIDParam {
			shallowField.Type = "link"
		} else {
			shallowField.Type = utils.TransformType(scheme.Type)
		}
		if scheme.GetLink() > 0 {
			d.ProcessLinkedSchema(&shallowField, scheme, tableName, schema)
		}
		shallowField, additionalActions = d.ProcessPermissions(shallowField, scheme, tableName, additionalActions, schema)
		var m map[string]interface{}
		b, _ = json.Marshal(shallowField)
		err := json.Unmarshal(b, &m)

		if err == nil {
			m["autofill"] = d.getFieldFill(schema, scheme.Name)
			m["translatable"] = scheme.Translatable
			m["hidden"] = scheme.Hidden
			schemes[scheme.Name] = m
		}
		keysOrdered = append(keysOrdered, scheme.Name)
	}

	sort.SliceStable(keysOrdered, func(i, j int) bool {
		return utils.ToInt64(utils.ToMap(schemes[keysOrdered[i]])["index"]) <= utils.ToInt64(utils.ToMap(schemes[keysOrdered[j]])["index"])
	})
	return schemes, schema.GetID(), keysOrdered, cols, additionalActions,
		!(slices.Contains(additionalActions, "post") && d.Domain.GetEmpty()) && !slices.Contains(additionalActions, "put")
}

func (d *ViewConvertor) ProcessLinkedSchema(shallowField *sm.ViewFieldModel, scheme sm.FieldModel, tableName string, s sm.SchemaModel) {
	schema, _ := sch.GetSchemaByID(scheme.GetLink())
	if !strings.Contains(shallowField.Type, "enum") && !strings.Contains(shallowField.Type, "many") {
		shallowField.Type = "link"
	} else {
		shallowField.Type = utils.TransformType(scheme.Type)
	}
	shallowField.ActionPath = fmt.Sprintf("/%s/%s?rows=all&%s=enable", utils.MAIN_PREFIX, schema.Name, utils.RootShallow)
	if (s.HasField(ds.SchemaDBField) && s.HasField(ds.DestTableDBField)) || schema.HasField(ds.SchemaDBField) {
		shallowField.LinkPath = shallowField.ActionPath
	}
	if strings.Contains(scheme.Type, "many") {
		for _, field := range schema.Fields {
			if strings.Contains(field.Name, "_id") && !strings.Contains(field.Name, tableName) && field.GetLink() > 0 {
				schField, _ := sch.GetSchemaByID(field.GetLink())
				shallowField.LinkPath = fmt.Sprintf("/%s/%s?rows=all&%s=enable", utils.MAIN_PREFIX, schField.Name, utils.RootShallow)
			}
		}
	}
}

func (d *ViewConvertor) ProcessPermissions(shallowField sm.ViewFieldModel, scheme sm.FieldModel,
	tableName string, additionalActions []string, schema sm.SchemaModel) (sm.ViewFieldModel, []string) {
	for _, meth := range []utils.Method{utils.SELECT, utils.CREATE, utils.UPDATE, utils.DELETE} {
		if d.Domain.VerifyAuth(tableName, "", "", meth) && (((meth == utils.SELECT || meth == utils.CREATE) && d.Domain.GetEmpty()) || !d.Domain.GetEmpty()) {
			if !slices.Contains(additionalActions, meth.Method()) {
				additionalActions = append(additionalActions, meth.Method())
			}
			if meth == utils.CREATE && !slices.Contains(additionalActions, "import") {
				additionalActions = d.CheckAndAddImportAction(additionalActions, schema)
			}
		}
		if scheme.GetLink() > 0 {
			shallowField = d.HandleRecursivePermissions(shallowField, scheme, meth)
		}

		if (meth == utils.UPDATE || meth == utils.CREATE) && d.Domain.GetEmpty() {
			shallowField.Readonly = false
		}
	}
	return shallowField, additionalActions
}

func (d *ViewConvertor) CheckAndAddImportAction(additionalActions []string, schema sm.SchemaModel) []string {
	d.Domain.GetDb().ClearQueryFilter()
	res, err := d.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{ds.SchemaDBField: schema.GetID()}, false)
	if err == nil && len(res) > 0 {
		ids := []string{}
		for _, rec := range res {
			ids = append(ids, utils.ToString(rec[utils.SpecialIDParam]))
		}
		d.Domain.GetDb().ClearQueryFilter()
		res, _ = d.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
			utils.SpecialIDParam: ids,
		}, false)
		if len(res) == 0 {
			additionalActions = append(additionalActions, "import")
		}
	}
	return additionalActions
}

func (d *ViewConvertor) HandleRecursivePermissions(shallowField sm.ViewFieldModel, scheme sm.FieldModel, meth utils.Method) sm.ViewFieldModel {
	schema, _ := sch.GetSchemaByID(scheme.GetLink())
	if d.Domain.VerifyAuth(schema.Name, "", "", meth) {
		if s, ok := d.SchemaSeen[schema.Name]; !ok {
			sch, _, _, _, _, _ := d.GetViewFields(schema.Name, true)
			d.SchemaSeen[schema.Name] = sch
			shallowField.DataSchema = sch
		} else {
			shallowField.DataSchema = s
		}
		if !strings.Contains(shallowField.Type, "enum") && !strings.Contains(shallowField.Type, "many") {
			shallowField.Type = "link"
		} else {
			shallowField.Type = utils.TransformType(scheme.Type)
		}
		shallowField.ActionPath = fmt.Sprintf("/%s/%s?rows=%s&%s=enable", utils.MAIN_PREFIX, schema.Name, utils.ReservedParam, utils.RootShallow)
		shallowField.Actions = append(shallowField.Actions, meth.Method())
	}
	return shallowField
}

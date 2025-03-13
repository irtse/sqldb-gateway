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
	"strings"
)

func (v *ViewConvertor) GetShortcuts() map[string]string {
	shortcuts := map[string]string{}
	if results, err := v.Domain.GetDb().SelectQueryWithRestriction(ds.DBView.Name, map[string]interface{}{"is_shortcut": true}, false); err == nil {
		for _, shortcut := range results {
			shortcuts[utils.GetString(shortcut, sm.NAMEKEY)] = "#" + utils.GetString(shortcut, utils.SpecialIDParam)
		}
	}
	return shortcuts
}

func (d *ViewConvertor) FetchRecord(tableName, id string) []map[string]interface{} {
	t, err := d.Domain.GetDb().SelectQueryWithRestriction(tableName, map[string]interface{}{utils.SpecialIDParam: id}, false)
	if err != nil || len(t) == 0 {
		return nil
	}
	return t
}

func (d *ViewConvertor) NewDataAccess(schemaID int64, destIDs []string, meth utils.Method) {
	if users, err := d.Domain.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
		"name": d.Domain.GetUser(), "email": d.Domain.GetUser(),
	}, true); err == nil && len(users) > 0 {
		for _, destID := range destIDs {
			id := utils.GetString(users[0], utils.SpecialIDParam)
			if meth == utils.DELETE {
				d.Domain.DeleteSuperCall(
					utils.GetRowTargetParameters(ds.DBDataAccess.Name, utils.ReservedParam).Enrich(
						map[string]interface{}{
							ds.DestTableDBField: destID,
							ds.SchemaDBField:    schemaID,
							ds.UserDBField:      id,
						}))
			} else {
				d.Domain.CreateSuperCall(utils.AllParams(ds.DBDataAccess.Name),
					utils.Record{
						"write":             meth == utils.CREATE,
						"update":            meth == utils.UPDATE,
						ds.DestTableDBField: destID,
						ds.SchemaDBField:    schemaID,
						ds.UserDBField:      id}, utils.CREATE)
			}
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
	fmt.Println("sch", schema.Fields)

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

		if scheme.GetLink() > 0 && !d.Domain.IsLowerResult() {
			d.processLinkedSchema(&shallowField, scheme, tableName)
		}

		d.processPermissions(&shallowField, scheme, tableName, &additionalActions, schema)
		schemes[scheme.Name] = shallowField
	}

	sort.SliceStable(keysOrdered, func(i, j int) bool {
		return schemes[keysOrdered[i]].(sm.ViewFieldModel).Index <= schemes[keysOrdered[j]].(sm.ViewFieldModel).Index
	})

	return schemes, schema.GetID(), keysOrdered, cols, additionalActions,
		!(slices.Contains(additionalActions, "post") && d.Domain.GetEmpty()) && !slices.Contains(additionalActions, "put")
}

func (d *ViewConvertor) processLinkedSchema(shallowField *sm.ViewFieldModel, scheme sm.FieldModel, tableName string) {
	schema, _ := sch.GetSchemaByID(scheme.GetLink())

	if !strings.Contains(shallowField.Type, "enum") && !strings.Contains(shallowField.Type, "many") {
		shallowField.Type = "link"
	} else {
		shallowField.Type = utils.TransformType(scheme.Type)
	}

	shallowField.ActionPath = fmt.Sprintf("/%s/%s?rows=all", utils.MAIN_PREFIX, schema.Name)
	shallowField.LinkPath = shallowField.ActionPath + "&" + utils.RootShallow + "=enable"

	if strings.Contains(scheme.Type, "many") {
		for _, field := range schema.Fields {
			if strings.Contains(field.Name, "_id") && !strings.Contains(field.Name, tableName) && field.GetLink() > 0 {
				schField, _ := sch.GetSchemaByID(field.GetLink())
				shallowField.LinkPath = fmt.Sprintf("/%s/%s?rows=all&%s=enable", utils.MAIN_PREFIX, schField.Name, utils.RootShallow)
			}
		}
	}
}

func (d *ViewConvertor) processPermissions(shallowField *sm.ViewFieldModel, scheme sm.FieldModel,
	tableName string, additionalActions *[]string, schema sm.SchemaModel) {
	for _, meth := range []utils.Method{utils.SELECT, utils.CREATE, utils.UPDATE, utils.DELETE} {
		if d.Domain.VerifyAuth(tableName, "", "", meth) && (((meth == utils.SELECT || meth == utils.CREATE) && d.Domain.GetEmpty()) || !d.Domain.GetEmpty()) {
			if !slices.Contains(*additionalActions, meth.Method()) {
				*additionalActions = append(*additionalActions, meth.Method())
			}

			if meth == utils.CREATE && !slices.Contains(*additionalActions, "import") {
				d.checkAndAddImportAction(additionalActions, schema)
			}
		}
		if scheme.GetLink() > 0 {
			d.handleRecursivePermissions(shallowField, scheme, meth)
		}

		if meth == utils.UPDATE && d.Domain.GetEmpty() {
			shallowField.Readonly = false
		} else if meth == utils.CREATE && d.Domain.GetEmpty() {
			shallowField.Readonly = true
		}
	}
}

func (d *ViewConvertor) checkAndAddImportAction(additionalActions *[]string, schema sm.SchemaModel) {
	res, err := d.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{ds.SchemaDBField: schema.GetID()}, false)
	if err == nil && len(res) > 0 {
		ids := []string{}
		for _, rec := range res {
			ids = append(ids, utils.ToString(rec[utils.SpecialIDParam]))
		}
		res, _ = d.Domain.GetDb().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
			ds.WorkflowDBField: ids,
		}, false)
		if len(res) == 0 {
			*additionalActions = append(*additionalActions, "import")
		}
	}
}

func (d *ViewConvertor) handleRecursivePermissions(shallowField *sm.ViewFieldModel, scheme sm.FieldModel, meth utils.Method) {
	schema, _ := sch.GetSchemaByID(scheme.GetLink())
	if d.Domain.VerifyAuth(schema.Name, "", "", meth) {
		sch, _, _, _, _, _ := d.GetViewFields(schema.Name, true)
		shallowField.DataSchema = sch
		if !strings.Contains(shallowField.Type, "enum") && !strings.Contains(shallowField.Type, "many") {
			shallowField.Type = "link"
		} else {
			shallowField.Type = utils.TransformType(scheme.Type)
		}
		shallowField.ActionPath = fmt.Sprintf("/%s/%s?rows=%s", utils.MAIN_PREFIX, schema.Name, utils.ReservedParam)
		shallowField.Actions = append(shallowField.Actions, meth.Method())
	}
}

package view_convertor

import (
	"fmt"
	"slices"
	"sqldb-ws/domain/domain_service/triggers"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
)

func CompareOrder(schema *sm.SchemaModel, order []string, domain utils.DomainITF) []string {
	newOrder := []string{}
	if res, err := GetFilterFields(schema, domain); err == nil && len(res) > 0 {
		for _, ord := range res {
			fmt.Println(order, utils.GetString(ord, "name"))
			if len(order) == 0 || slices.Contains(order, utils.GetString(ord, "name")) {
				newOrder = append(newOrder, utils.GetString(ord, "name"))
			}
		}
	}
	fmt.Println("GetFilterFields", newOrder, len(newOrder), len(order))
	if len(newOrder) == 0 {
		return order
	}
	return newOrder
}

func getRedirection(domainID string) string {
	if triggers.HasRedirection(domainID) {
		s, _ := triggers.GetRedirection(domainID)
		return s
	}
	return ""
}

func GetOrder(schema *sm.SchemaModel, record utils.Record, values map[string]interface{}, newOrder []string, domain utils.DomainITF) ([]string, map[string]interface{}) {
	if res, err := GetFilterFields(schema, domain); err == nil && len(res) > 0 {
		if utils.GetBool(record, "is_list") {
			for _, r := range res {
				fmt.Println("GetFilterFields2", len(res), r)
				if val, err := schema.GetField(utils.GetString(r, "name")); err == nil {
					utils.ToMap(val)["readonly"] = true
					utils.ToMap(val)["hidden"] = true
				}
				values[utils.GetString(r, "name")] = nil
			}
		} else {
			fmt.Println("GetFilterFields2", newOrder, utils.GetString(res[0], "name"), len(res))
			newOrder = append(newOrder, utils.GetString(res[0], "name"))
		}
	}
	return newOrder, values
}

func GetFilterFields(schema *sm.SchemaModel, domain utils.DomainITF) ([]map[string]interface{}, error) {
	m := map[string]interface{}{
		ds.FilterDBField: domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
			"is_view":              true,
			"dashboard_restricted": false,
		}, false, utils.SpecialIDParam),
	}
	if domain.GetEmpty() {
		m[ds.FilterDBField+"_100"] = domain.GetDb().BuildSelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
			ds.SchemaDBField: schema.ID,
		}, false, "view_"+ds.FilterDBField)
	} else if domain.GetUserID() != "" {
		m[ds.FilterDBField+"_100"] = domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
			utils.SpecialIDParam: domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.UserDBField: domain.GetUserID(),
				ds.EntityDBField: domain.GetDb().BuildSelectQueryWithRestriction(
					ds.DBEntityUser.Name,
					map[string]interface{}{
						ds.UserDBField: domain.GetUserID(),
					}, true, ds.EntityDBField),
			}, true, ds.WorkflowSchemaDBField),
		}, false, "view_"+ds.FilterDBField)
		m[ds.FilterDBField+"_101"] = domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
			utils.SpecialIDParam: domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.SchemaDBField: schema.ID,
				"is_close":       false,
			}, false, ds.WorkflowSchemaDBField),
		}, false, "view_"+ds.FilterDBField)
	}
	return domain.GetDb().SelectQueryWithRestriction(ds.DBSchemaField.Name, map[string]interface{}{
		utils.SpecialIDParam: domain.GetDb().BuildSelectQueryWithRestriction(ds.DBFilterField.Name, m, false, ds.SchemaFieldDBField),
	}, false)
}

func GetSharing(schemaID string, rec sm.ViewItemModel, domain utils.DomainITF) sm.ViewItemModel {
	id := rec.Values[utils.SpecialIDParam]
	m := map[string]interface{}{
		ds.UserDBField:      domain.GetUserID(),
		ds.SchemaDBField:    schemaID,
		ds.DestTableDBField: id,
	}
	addDate := []string{}
	addBool := []string{"update_access", "delete_access"}
	table := ds.DBShare.Name
	kind := "share"
	if domain.GetTable() == ds.DBTask.Name {
		kind = "delegate"
		table = ds.DBDelegation.Name
		addDate = []string{"start_date", "end_date"}
		addBool = []string{"all_tasks"}
		m["all_tasks"] = true
		m[ds.TaskDBField] = id
	} else {
		m["delete_access"] = true
		m["update_access"] = true
		m[ds.SchemaDBField] = schemaID
		m[ds.DestTableDBField] = id
	}
	rec.Sharing = sm.SharingModel{
		SharedWithPath: fmt.Sprintf("/%s/%s?%s=%s&%s=disable_"+kind, utils.MAIN_PREFIX, ds.DBUser.Name, utils.RootRowsParam,
			utils.ReservedParam, utils.RootScope),
		Body:            m,
		AdditionnalDate: addDate,
		AdditionnalBool: addBool,
		ShallowPath: map[string]string{
			kind + "d_" + ds.UserDBField: fmt.Sprintf("/%s/%s?%s=%s&%s=enable&%s=enable_"+kind, utils.MAIN_PREFIX, ds.DBUser.Name,
				utils.RootRowsParam, utils.ReservedParam, utils.RootShallow, utils.RootScope),
		},
		Path: fmt.Sprintf("/%s/%s?%s=%s&%s=enable", utils.MAIN_PREFIX, table, utils.RootRowsParam, utils.ReservedParam, utils.RootShallow),
	}
	return rec
}

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
			if len(order) == 0 || slices.Contains(order, utils.GetString(ord, "name")) {
				newOrder = append(newOrder, utils.GetString(ord, "name"))
			}
		}
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
				if val, err := schema.GetField(utils.GetString(r, "name")); err == nil {
					utils.ToMap(val)["readonly"] = true
					utils.ToMap(val)["hidden"] = true
				}
				values[utils.GetString(r, "name")] = nil
			}
		} else {
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
	} else {
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
	return domain.GetDb().SelectQueryWithRestriction(ds.SchemaFieldDBField, map[string]interface{}{
		utils.SpecialIDParam: domain.GetDb().BuildSelectQueryWithRestriction(ds.FilterFieldDBField, m, false, ds.SchemaFieldDBField),
	}, false)
}

func GetSharing(schemaID string, rec sm.ViewItemModel, domain utils.DomainITF) sm.ViewItemModel {
	id := rec.Values[utils.SpecialIDParam]
	m := map[string]interface{}{
		ds.UserDBField:      domain.GetUserID(),
		ds.SchemaDBField:    schemaID,
		ds.DestTableDBField: id,
		"read_access":       true,
		"update_access":     true,
		"delete_access":     true,
	}
	rec.Sharing = sm.SharingModel{
		SharedWithPath: fmt.Sprintf("/%s/%s?%s=%s&%s=disable", utils.MAIN_PREFIX, ds.DBUser.Name, utils.RootRowsParam,
			utils.ReservedParam, utils.RootScope),
		Body: m,
		ShallowPath: map[string]string{
			"shared_" + ds.UserDBField: fmt.Sprintf("/%s/%s?%s=%s&%s=enable&%s=enable", utils.MAIN_PREFIX, ds.DBUser.Name,
				utils.RootRowsParam, utils.ReservedParam, utils.RootShallow, utils.RootScope),
		},
		Path: fmt.Sprintf("/%s/%s?%s=%s&%s=enable", utils.MAIN_PREFIX, ds.DBShare.Name, utils.RootRowsParam, utils.ReservedParam, utils.RootShallow),
	}
	return rec
}

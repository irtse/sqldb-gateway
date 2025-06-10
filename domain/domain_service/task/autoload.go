package task

import (
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
)

func Load(domain utils.DomainITF) {
	if res, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
		"is_close": true,
	}, false); err == nil {
		for _, r := range res {
			SetEndedRequest(utils.GetString(r, ds.SchemaDBField), utils.GetString(r, ds.DestTableDBField), utils.GetString(r, utils.SpecialIDParam), domain.GetDb())
		}
	}
	if res, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{}, false); err == nil {
		for _, r := range res {
			CreateTask(domain, r)
		}
	}
}

func CreateTask(domain utils.DomainITF, record utils.Record) {
	f, ok := record["view_"+ds.FilterDBField]
	if !utils.GetBool(record, "is_close") {
		view := []string{}
		if ok {
			if res, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.SchemaFieldDBField, map[string]interface{}{
				utils.SpecialIDParam: domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.FilterFieldDBField,
					map[string]interface{}{
						ds.FilterFieldDBField: f,
					}, false, ds.SchemaFieldDBField),
			}, true); err == nil && len(res) > 0 {
				for _, r := range res {
					view = append(view, utils.GetString(r, "name"))
				}
			}
		}
		if _, ok := record["readonly_not_assignee"]; ok && utils.GetBool(record, "readonly_not_assignee") {
			SetReadonlyTask(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam), utils.GetString(record, ds.UserDBField))
		}
		SetViewTask(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam), utils.GetString(record, ds.UserDBField), view)
		SetTasks(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam),
			utils.GetString(record, ds.RequestDBField), utils.GetString(record, utils.SpecialIDParam))
	} else {
		RemoveTask(record, utils.GetString(record, ds.UserDBField))
	}
}

func RemoveTask(record utils.Record, userID string) {
	if _, ok := record["readonly_not_assignee"]; ok && utils.GetBool(record, "readonly_not_assignee") {
		DeleteReadonlyTask(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam), userID)
	}
	DeleteViewTask(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam), utils.GetString(record, ds.UserDBField))
	DeleteTasks(utils.GetString(record, ds.SchemaDBField), utils.GetString(record, utils.RootDestTableIDParam), userID)
}

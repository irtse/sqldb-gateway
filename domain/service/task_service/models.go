package task_service

import (
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"time"
)

var SchemaDBField = ds.RootID(ds.DBSchema.Name)
var RequestDBField = ds.RootID(ds.DBRequest.Name)
var WorkflowDBField = ds.RootID(ds.DBWorkflowSchema.Name)
var UserDBField = ds.RootID(ds.DBUser.Name)
var EntityDBField = ds.RootID(ds.DBEntity.Name)
var DestTableDBField = ds.RootID("dest_table")
var FilterDBField = ds.RootID(ds.DBFilter.Name)

func NewTask(name interface{}, description interface{}, urgency interface{}, priority interface{},
	workflowDB interface{}, schemaDB interface{}, requestDB interface{}, userDB interface{}, entityDB interface{}) utils.Record {
	r := utils.Record{
		"name":          name,
		"description":   description,
		"urgency":       urgency,
		"priority":      priority,
		WorkflowDBField: workflowDB,
		SchemaDBField:   schemaDB,
		RequestDBField:  requestDB,
		UserDBField:     userDB,
		EntityDBField:   entityDB,
	}
	toDelete := []string{}
	for k, v := range r {
		if v == nil || v == "" {
			toDelete = append(toDelete, k)
		}
	}
	for _, k := range toDelete {
		delete(r, k)
	}
	return r
}

func CheckStateIsEnded(state interface{}) bool {
	return state == "completed" || state == "dismiss" || state == "refused" || state == "canceled"
}

func SetClosureStatus(res map[string]interface{}) map[string]interface{} {
	if state, ok := res["state"]; ok && CheckStateIsEnded(utils.ToString(state)) {
		res["is_close"] = true
		res["closing_date"] = time.Now().Format(time.RFC3339)
	} else {
		res["state"] = "progressing"
		res["is_close"] = false
		res["closing_date"] = nil
	}
	return res
}

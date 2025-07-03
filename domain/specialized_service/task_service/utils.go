package task_service

import (
	"fmt"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	connector "sqldb-ws/infrastructure/connector/db"
	"strings"
	"time"
)

var SchemaDBField = ds.RootID(ds.DBSchema.Name)
var RequestDBField = ds.RootID(ds.DBRequest.Name)
var WorkflowSchemaDBField = ds.RootID(ds.DBWorkflowSchema.Name)
var UserDBField = ds.RootID(ds.DBUser.Name)
var EntityDBField = ds.RootID(ds.DBEntity.Name)
var DestTableDBField = ds.RootID("dest_table")
var FilterDBField = ds.RootID(ds.DBFilter.Name)

func ConstructNotificationTask(scheme utils.Record, request utils.Record) map[string]interface{} {
	task := map[string]interface{}{
		sm.NAMEKEY:               scheme.GetString(sm.NAMEKEY),
		"description":            scheme.GetString(sm.NAMEKEY),
		"urgency":                scheme["urgency"],
		"priority":               scheme["priority"],
		ds.WorkflowSchemaDBField: scheme[utils.SpecialIDParam],
		ds.UserDBField:           scheme[ds.UserDBField],
		ds.EntityDBField:         scheme[ds.EntityDBField],
		ds.SchemaDBField:         scheme[ds.SchemaDBField],
		ds.DestTableDBField:      scheme[ds.DestTableDBField],
		ds.RequestDBField:        request[utils.SpecialIDParam],
		"send_mail_to":           scheme["send_mail_to"],
		"opening_date":           time.Now().Format(time.RFC3339),

		"override_state_completed": scheme["override_state_completed"],
		"override_state_dismiss":   scheme["override_state_dismiss"],
		"override_state_refused":   scheme["override_state_refused"],
	}
	return task
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

func CreateNewDataFromTask(schema sm.SchemaModel, newTask utils.Record, record utils.Record, domain utils.DomainITF) utils.Record {
	r := utils.Record{"is_draft": true}
	if schema.HasField("name") {
		if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err == nil {
			if res, err := domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
				utils.SpecialIDParam: record[ds.DestTableDBField],
			}, false); err == nil && len(res) > 0 {
				r[sm.NAMEKEY] = utils.GetString(res[0], "name")
			}
		} else {
			r["name"] = utils.GetString(newTask, "name")
		}
	}
	if schema.HasField(ds.DestTableDBField) && schema.HasField(ds.SchemaDBField) {
		// get workflow source schema + dest ID
		r[ds.DestTableDBField] = record[ds.DestTableDBField]
		r[ds.SchemaDBField] = record[ds.SchemaDBField]
	}
	if schema.HasField(ds.UserDBField) {
		r[ds.UserDBField] = record[ds.UserDBField]
	}
	if schema.HasField(ds.EntityDBField) {
		r[ds.EntityDBField] = record[ds.EntityDBField]
	}
	for _, f := range schema.Fields {
		if f.GetLink() == record[ds.SchemaDBField] {
			r[f.Name] = record[ds.DestTableDBField]
		}
	}

	if i, err := domain.GetDb().ClearQueryFilter().CreateQuery(schema.Name, r, func(s string) (string, bool) { return "", true }); err == nil {
		r[utils.SpecialIDParam] = i

		newTask[ds.DestTableDBField] = i
		domain.GetDb().CreateQuery(ds.DBDataAccess.Name, map[string]interface{}{
			ds.SchemaDBField:    schema.ID,
			ds.DestTableDBField: i,
			ds.UserDBField:      domain.GetUserID(),
			"write":             true,
			"update":            false,
		}, func(s string) (string, bool) {
			return "", true
		})
	}
	return newTask
}

func PrepareAndCreateTask(scheme utils.Record, request map[string]interface{}, record map[string]interface{}, domain utils.DomainITF, fromTask bool) map[string]interface{} {
	newTask := ConstructNotificationTask(scheme, request)
	delete(newTask, utils.SpecialIDParam)
	if utils.GetBool(scheme, "assign_to_creator") {
		newTask[ds.UserDBField] = domain.GetUserID()
	}
	if utils.GetString(newTask, ds.SchemaDBField) == utils.GetString(request, ds.SchemaDBField) {
		newTask[ds.SchemaDBField] = request[ds.SchemaDBField]
		newTask[ds.DestTableDBField] = request[ds.DestTableDBField]
	} else if schema, err := schserv.GetSchemaByID(utils.GetInt(newTask, ds.SchemaDBField)); err == nil {
		newTask = CreateNewDataFromTask(schema, newTask, record, domain)
	}
	isMeta := strings.Contains(utils.GetString(record, "nexts"), utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField)) && utils.GetString(scheme, "wrapped_"+ds.WorkflowDBField) != "" || !fromTask
	if id, ok := scheme["wrapped_"+ds.WorkflowDBField]; ok && id != nil && isMeta {
		createMetaRequest(newTask, id, domain)
	}
	shouldCreate := utils.GetString(record, "nexts") == utils.ReservedParam || utils.GetString(record, "nexts") == "" || isMeta
	if shouldCreate {
		createTaskAndNotify(newTask, domain, fromTask)
	}
	return newTask
}

func createTaskAndNotify(task map[string]interface{}, domain utils.DomainITF, isTask bool) {
	i, err := domain.GetDb().CreateQuery(ds.DBTask.Name, task, func(s string) (string, bool) {
		return "", true
	})
	if err != nil {
		return
	}
	CreateDelegated(task, i, domain)
	notify(task, i, domain)
}

func notify(task utils.Record, i int64, domain utils.DomainITF) {
	if schema, err := schserv.GetSchema(ds.DBTask.Name); err == nil {
		name := utils.GetString(task, "name")
		if res, err := domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
			utils.SpecialIDParam: i,
		}, false); err == nil && len(res) > 0 {
			name += " <" + utils.GetString(res[0], "name") + ">"
		}
		notif := utils.Record{
			"name":              utils.GetString(task, "name"),
			"description":       utils.GetString(task, "description"),
			ds.UserDBField:      task[ds.UserDBField],
			ds.EntityDBField:    task[ds.EntityDBField],
			ds.DestTableDBField: i,
		}
		notif["link_id"] = schema.ID
		domain.GetDb().ClearQueryFilter().CreateQuery(ds.DBNotification.Name, notif, func(s string) (string, bool) {
			return "", true
		})
	}
}

func createMetaRequest(task map[string]interface{}, id interface{}, domain utils.DomainITF) {
	domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name).RootRaw(), utils.Record{
		ds.WorkflowDBField:  id,
		sm.NAMEKEY:          "Meta request for " + utils.GetString(task, sm.NAMEKEY) + " task.",
		"current_index":     1,
		"is_meta":           true,
		ds.SchemaDBField:    task[ds.SchemaDBField],
		ds.DestTableDBField: task[ds.DestTableDBField],
		ds.UserDBField:      utils.GetInt(task, ds.UserDBField),
	})
}

func CreateDelegated(record utils.Record, id int64, domain utils.DomainITF) {
	currentTime := time.Now()
	sqlFilter := []string{
		"('" + currentTime.Format("2000-01-01") + "' < start_date OR '" + currentTime.Format("2000-01-01") + "' > end_date)",
	}
	sqlFilter = append(sqlFilter, connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"all_tasks":    true,
		ds.UserDBField: domain.GetUserID(),
	}, false))
	if dels, err := domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(
		ds.DBDelegation.Name, utils.ToListAnonymized(sqlFilter), false); err == nil && len(dels) > 0 {
		newRec := record.Copy()
		for _, delegated := range dels {
			newRec["binded_dbtask"] = id
			newRec[ds.UserDBField] = delegated["delegated_"+ds.UserDBField]
			delete(newRec, utils.SpecialIDParam)
			fmt.Println("CREATE DELEGATED", newRec)
			domain.GetDb().ClearQueryFilter().CreateQuery(ds.DBTask.Name, record, func(s string) (string, bool) { return "", true })
		}
	}
}

func UpdateDelegated(task utils.Record, domain utils.DomainITF) {
	id := task[utils.SpecialIDParam]
	if task["binded_dbtask"] != nil {
		id := task["binded_dbtask"]
		domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBTask.Name, map[string]interface{}{
			"state":           task["state"],
			"is_close":        task["is_close"],
			"nexts":           task["nexts"],
			"closing_date":    task["closing_date"],
			"closing_by":      task["closing_by"],
			"closing_comment": task["closing_comment"],
		}, map[string]interface{}{
			utils.SpecialIDParam: id,
		}, true)
	}
	domain.GetDb().ClearQueryFilter().UpdateQuery(ds.DBTask.Name, map[string]interface{}{
		"state":           task["state"],
		"is_close":        task["is_close"],
		"nexts":           task["nexts"],
		"closing_date":    task["closing_date"],
		"closing_by":      task["closing_by"],
		"closing_comment": task["closing_comment"],
	}, map[string]interface{}{
		"binded_dbtask": id,
	}, true)
}

func HandleHierarchicalVerification(domain utils.DomainITF, requestID int64, record map[string]interface{}) map[string]interface{} {
	if hierarchy, err := GetHierarchical(domain); err == nil {
		for _, hierarch := range hierarchy {
			CreateHierarchicalTask(domain, requestID, record, hierarch)
		}
	}
	return record
}

func CreateHierarchicalTask(domain utils.DomainITF, requestID int64, record, hierarch map[string]interface{}) {
	newTask := utils.Record{
		ds.SchemaDBField:    record[ds.SchemaDBField],
		ds.DestTableDBField: record[ds.DestTableDBField],
		ds.RequestDBField:   requestID,
		ds.UserDBField:      hierarch["parent_"+ds.UserDBField],
		"description":       "hierarchical verification expected by the system.",
		"urgency":           "normal",
		"priority":          "normal",
		sm.NAMEKEY:          "hierarchical verification",
	}
	if i, err := domain.GetDb().CreateQuery(ds.DBTask.Name, newTask, func(s string) (string, bool) {
		return "", true
	}); err == nil {
		CreateDelegated(newTask, i, domain)
		notify(newTask, i, domain)
	}
}

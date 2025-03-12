package task_service

import (
	"errors"
	"sqldb-ws/domain/filter"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	utils "sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	conn "sqldb-ws/infrastructure/connector"
	"strings"
	"time"
)

// TODO
type TaskService struct {
	servutils.AbstractSpecializedService
}

func (s *TaskService) ShouldVerify() bool                                                   { return true }
func (s *TaskService) SpecializedCreateRow(record map[string]interface{}, tableName string) {}
func (s *TaskService) Entity() utils.SpecializedServiceInfo                                 { return ds.DBTask }
func (s *TaskService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true)
}
func (s *TaskService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	switch s.Domain.GetMethod() {
	case utils.DELETE:
		return servutils.CheckAutoLoad(tablename, record, s.Domain)
	case utils.CREATE:
		user, err := s.Domain.SuperCall(utils.AllParams(ds.DBUser.Name), utils.Record{}, utils.SELECT, false,
			conn.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
				"name":  conn.Quote(s.Domain.GetUser()),
				"email": conn.Quote(s.Domain.GetUser()),
			}, true))
		if err != nil || len(user) == 0 {
			return record, errors.New("user not found"), false
		}
		record[ds.DBUser.Name] = user[0][utils.SpecialIDParam] // affected create_by
		record["created_date"] = time.Now().Format(time.RFC3339)
	case utils.UPDATE:
		// check if task is already closed
		if elder, _ := s.Domain.SuperCall(utils.GetRowTargetParameters(ds.DBTask.Name, record[utils.SpecialIDParam]),
			utils.Record{}, utils.SELECT, false); len(elder) > 0 && CheckStateIsEnded(utils.ToString(elder[0]["state"])) {
			return record, errors.New("task is already closed, cannot change its state"), false
		}
		record = SetClosureStatus(record) // check if task is already progressing
	}
	return record, nil, true
}

func (s *TaskService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for i, res := range results {
		res["state"] = "completed"
		results[i] = SetClosureStatus(res)
	}
	s.SpecializedUpdateRow(results, map[string]interface{}{})
}

func (s *TaskService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	for _, res := range results {
		if CheckStateIsEnded(utils.ToString(res["state"])) {
			continue
		}
		p := utils.GetRowTargetParameters(ds.DBRequest.Name, utils.GetString(res, RequestDBField))
		requests, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT, false)
		if err != nil || len(requests) == 0 || utils.ToInt64(requests[0]["current_index"]) < 1 {
			continue
		}
		order := requests[0]["current_index"]
		if otherPendingTasks, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name,
			map[string]interface{}{ // delete all notif
				RequestDBField: utils.ToString(res[RequestDBField]),
				"state":        []string{"pending", "progressing"},
			}, false); err == nil && len(otherPendingTasks) > 0 {
			continue
		}

		current_index := utils.ToInt64(order)
		switch res["state"] {
		case "completed":
			current_index++
		case "dismiss":
			if current_index > 0 {
				current_index--
			} else { // Dismiss will close requests.
				s.Domain.Call(p, SetClosureStatus(utils.Record{"state": "dismiss"}), utils.UPDATE)
			} // no before task close request and task
		}
		schemes := utils.Results{}
		newRecRequest := utils.Record{utils.SpecialIDParam: requests[0][utils.SpecialIDParam]}
		if schemes, err := s.Domain.SuperCall(utils.AllParams(ds.DBWorkflowSchema.Name), utils.Record{},
			utils.SELECT, false, conn.FormatSQLRestrictionWhereByMap("",
				map[string]interface{}{
					"index":         current_index,
					WorkflowDBField: requests[0][WorkflowDBField],
				}, false)); err != nil || len(schemes) == 0 { // no new task in workflow
			newRecRequest["state"] = "completed"
		} else {
			newRecRequest["current_index"] = current_index
			newRecRequest["state"] = "progressing"
		}
		newRecRequest = SetClosureStatus(newRecRequest)
		s.Domain.UpdateSuperCall(utils.AllParams(ds.DBRequest.Name).RootRaw(), newRecRequest)
		for _, scheme := range schemes {
			params := utils.GetRowTargetParameters(ds.DBTask.Name, nil).Enrich(
				map[string]interface{}{
					WorkflowDBField: scheme.GetString(utils.SpecialIDParam),
					RequestDBField:  requests[0].GetString(utils.SpecialIDParam),
				})
			if beforeTask, err := s.Domain.SuperCall(
				params, utils.Record{}, utils.SELECT, false, "is_close=false",
			); err == nil && len(beforeTask) > 0 {
				continue
			}
			newTask := NewTask(
				scheme["name"],
				scheme["description"],
				scheme["urgency"],
				scheme["priority"],
				scheme[utils.SpecialIDParam],
				scheme[SchemaDBField],
				requests[0][utils.SpecialIDParam],
				requests[0][UserDBField],
				requests[0][EntityDBField])

			if scheme[SchemaDBField] == res[SchemaDBField] {
				newTask[SchemaDBField] = res[SchemaDBField]
				newTask[DestTableDBField] = res[DestTableDBField]
			} else {
				schema, err := schserv.GetSchemaByID(scheme.GetInt(SchemaDBField))
				if err == nil {
					vals, err := s.Domain.CreateSuperCall(utils.AllParams(schema.Name), utils.Record{})
					if err == nil && len(vals) > 0 {
						newTask[SchemaDBField] = scheme[SchemaDBField]
						newTask[DestTableDBField] = vals[0][utils.ReservedParam]
					}
				}
			}
			if strings.Contains(utils.GetString(res, "nexts"), scheme.GetString("wrapped_"+WorkflowDBField)) {
				newMetaRequest := utils.Record{
					WorkflowDBField: scheme["wrapped_"+WorkflowDBField],
					sm.NAMEKEY:      "Meta request for " + newTask.GetString(sm.NAMEKEY) + " task",
					"current_index": 1, "is_meta": true,
					SchemaDBField:    newTask[SchemaDBField],
					DestTableDBField: newTask[DestTableDBField],
					UserDBField:      newTask[UserDBField],
				}
				requests, err := s.Domain.Call(utils.AllParams(ds.DBRequest.Name), newMetaRequest, utils.CREATE)
				if err == nil && len(requests) > 0 {
					newTask["meta_"+RequestDBField] = requests[0][utils.SpecialIDParam]
				}
			}
			if utils.GetString(res, "nexts") == utils.ReservedParam || strings.Contains(utils.GetString(res, "nexts"),
				scheme.GetString("wrapped_"+ds.WorkflowDBField)) {
				tasks, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBTask.Name), newTask)
				if err != nil || len(tasks) == 0 {
					continue
				}
				schema, err := schserv.GetSchema(ds.DBTask.Name)
				if err == nil && tasks[0]["meta_"+RequestDBField] == nil {
					s.Domain.CreateSuperCall(utils.AllParams(ds.DBNotification.Name), utils.Record{"link_id": schema.ID,
						sm.NAMEKEY:       "Task affected : " + tasks[0].GetString(sm.NAMEKEY),
						"description":    "Task is affected : " + tasks[0].GetString(sm.NAMEKEY),
						UserDBField:      tasks[0][UserDBField],
						EntityDBField:    scheme[EntityDBField],
						UserDBField:      scheme[UserDBField],
						DestTableDBField: tasks[0][utils.SpecialIDParam]})
				}
			}
		}
	}
}

func (s *TaskService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	if !s.Domain.IsSuperCall() {
		service := filter.NewFilterService(s.Domain)
		innerestr = append(innerestr, conn.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			"meta_" + RequestDBField: nil,
			UserDBField:              service.GetUserFilterQuery("id"),
			EntityDBField:            service.GetEntityFilterQuery("id"),
		}, false))
		return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, innerestr...)
	}
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, innerestr...)
}

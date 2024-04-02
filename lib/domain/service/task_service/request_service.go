package task_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type RequestService struct { tool.AbstractSpecializedService }

func (s *RequestService) Entity() tool.SpecializedServiceInfo { return entities.DBRequest }
func (s *RequestService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	if create { 
		sqlFilter := "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser())
		params := tool.Params{ 
			tool.RootTableParam : entities.DBUser.Name, 
			tool.RootRowsParam : tool.ReservedParam, 
		}
		user, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
		if err != nil || len(user) == 0 { return record, false, true }
		record[entities.RootID(entities.DBUser.Name)]=user[0][tool.SpecialIDParam] 
		paramsNew := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, tool.RootRowsParam: tool.ReservedParam }
		sqlFilter = "id IN (SELECT id FROM " + entities.DBHierarchy.Name + " WHERE "
		sqlFilter += entities.RootID(entities.DBUser.Name) + " IN ("
		sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")"
		sqlFilter += " OR " + entities.DBEntity.Name + "_id IN ("
		sqlFilter += "SELECT " + entities.DBEntity.Name + "_id FROM "
		sqlFilter += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
		sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
		sqlFilter += "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")))"
		hierarchy, err := s.Domain.SuperCall( 
								paramsNew, 
								tool.Record{},
								tool.SELECT,
								"Get",
								sqlFilter,
							)
		if err != nil || len(hierarchy) > 0 { record["current_index"]=0 
		} else { record["current_index"]=1 }
		params = tool.Params{ 
			tool.RootTableParam : entities.DBWorkflow.Name, 
			tool.RootRowsParam : tool.ReservedParam, 
			entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBWorkflow.Name)]),
		}
		wf, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(wf) == 0 { record[entities.NAMEATTR]= "Anonymous Request"
		} else { record["name"]= wf[0].GetString(entities.NAMEATTR) + " Request" }
		
	} else {
		if record["state"] == "completed" || record["state"] == "dismiss" { record["is_close"] = true }
	}
	return record, true, true
}
func (s *RequestService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *RequestService) UpdateRowAutomation(results tool.Results, record tool.Record) {
	for _, rec := range results {
		if rec["state"] == "dismiss" { 
			params := tool.Params{ tool.RootTableParam : entities.DBTask.Name,
								   tool.RootRowsParam : tool.ReservedParam,
								   tool.RootRawView : "enable",
								   entities.RootID(entities.DBRequest.Name) : rec.GetString(tool.SpecialIDParam) }
			s.Domain.SuperCall( params, tool.Record{ "state" : "dismiss", "is_close" : true, }, tool.UPDATE, "CreateOrUpdate")
			params = tool.Params{ tool.RootTableParam : entities.DBNotification.Name,
				tool.RootRowsParam : tool.ReservedParam,
				tool.RootRawView : "enable", }
			s.Domain.SuperCall( params, tool.Record{ 
				entities.NAMEATTR : "Rejected " + rec.GetString(entities.NAMEATTR) + " request", 
				"description" : rec.GetString(entities.NAMEATTR) + " request is rejected and closed.",
				entities.RootID(entities.DBUser.Name) : rec.GetString(entities.RootID(entities.DBUser.Name)),
				entities.RootID(entities.DBSchema.Name) : rec.GetString(entities.RootID(entities.DBSchema.Name)),
				entities.RootID("dest_table") : rec.GetString("dest_table"), }, tool.CREATE, "CreateOrUpdate")
		}
		if rec["state"] == "completed" {
			params := tool.Params{ tool.RootTableParam : entities.DBNotification.Name,
				tool.RootRowsParam : tool.ReservedParam,
				tool.RootRawView : "enable", }
			s.Domain.SuperCall( params, tool.Record{ 
				entities.NAMEATTR : "Validated " + rec.GetString(entities.NAMEATTR) + " request", 
				"description" : rec.GetString(entities.NAMEATTR) + " request is approved and closed.",
				entities.RootID(entities.DBUser.Name) : rec.GetString(entities.RootID(entities.DBUser.Name)),
				entities.RootID(entities.DBSchema.Name) : rec.GetString(entities.RootID(entities.DBSchema.Name)),
				entities.RootID("dest_table") : rec.GetString("dest_table"), }, tool.CREATE, "CreateOrUpdate")
		}
	}
}

func (s *RequestService) WriteRowAutomation(record tool.Record, tableName string) {
	if record["current_index"].(int64) == 1 {
		params := tool.Params{ tool.RootTableParam : entities.DBWorkflowSchema.Name,
		                       tool.RootRowsParam : tool.ReservedParam,
							   "index": "1", // lauch workflow
							   entities.RootID(entities.DBWorkflow.Name) : fmt.Sprintf("%v", record[entities.RootID(entities.DBWorkflow.Name)]) }
		wfs, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(wfs) == 0 { return }
		newTask := tool.Record{
			entities.RootID(entities.DBSchema.Name) : wfs[0][entities.RootID(entities.DBSchema.Name)],
			entities.RootID(entities.DBWorkflowSchema.Name) : wfs[0][tool.SpecialIDParam],
			entities.RootID(entities.DBRequest.Name) : record[tool.SpecialIDParam],
			entities.RootID(entities.DBUser.Name) : record[entities.RootID(entities.DBUser.Name)],
			"description" : wfs[0]["description"],
			"urgency" : wfs[0]["urgency"],
			"priority" : wfs[0]["priority"],
			entities.NAMEATTR : wfs[0][entities.NAMEATTR],
		}
		if fmt.Sprintf("%v", wfs[0][entities.RootID(entities.DBSchema.Name)]) == fmt.Sprintf("%v", record[entities.RootID(entities.DBSchema.Name)]) {
			newTask[entities.RootID(entities.DBSchema.Name)] = record[entities.RootID(entities.DBSchema.Name)]
			newTask[entities.RootID("dest_table")] = record[entities.RootID("dest_table")]
		} else {
			schemas, err := s.Domain.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : wfs[0][entities.RootID(entities.DBSchema.Name)] }, false)
			if err == nil && len(schemas) > 0 {
				vals, err := s.Domain.SuperCall( tool.Params{ tool.RootTableParam: schemas[0].GetString(entities.NAMEATTR),
												 tool.RootRowsParam: tool.ReservedParam, }, tool.Record{}, tool.CREATE, "CreateOrUpdate")
				if err == nil && len(vals) > 0 {
					newTask[entities.RootID(entities.DBSchema.Name)] = wfs[0][entities.RootID(entities.DBSchema.Name)]
					newTask[entities.RootID("dest_table")] = vals[0][tool.ReservedParam]
				}
			}
		}
		params = tool.Params{ tool.RootTableParam : entities.DBTask.Name,
			tool.RootRowsParam : tool.ReservedParam, }
		_, err = s.Domain.SuperCall( params, newTask, tool.CREATE, "CreateOrUpdate")
		if err == nil {
			s.Domain.SuperCall( params, tool.Record{ 
				entities.NAMEATTR : "Task affected : " + newTask.GetString(entities.NAMEATTR), 
				"description" : "Task is affected to you and must be treated : " + newTask.GetString(entities.NAMEATTR),
				entities.RootID(entities.DBUser.Name) : newTask.GetString(entities.RootID(entities.DBUser.Name)),
				entities.RootID(entities.DBSchema.Name) : newTask.GetString(entities.RootID(entities.DBSchema.Name)),
				entities.RootID("dest_table") : record[entities.RootID("dest_table")], }, tool.CREATE, "CreateOrUpdate")
		}
	} else {
		paramsNew := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, tool.RootRowsParam: tool.ReservedParam }
		sqlFilter := "id IN (SELECT id FROM " + entities.DBHierarchy.Name + " WHERE "
		sqlFilter += entities.RootID(entities.DBUser.Name) + " IN ("
		sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")"
		sqlFilter += " OR " + entities.DBEntity.Name + "_id IN ("
		sqlFilter += "SELECT " + entities.DBEntity.Name + "_id FROM "
		sqlFilter += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
		sqlFilter += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
		sqlFilter += "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")))"
		hierarchy, _ := s.Domain.SuperCall( 
								paramsNew, 
								tool.Record{},
								tool.SELECT,
								"Get",
								sqlFilter,
							)
		for _, hierarch := range hierarchy {
			newTask := tool.Record{
				entities.RootID(entities.DBSchema.Name) : record[entities.RootID(entities.DBSchema.Name)],
				entities.RootID("dest_table") : record[entities.RootID("dest_table")],
				entities.RootID(entities.DBRequest.Name) : record[tool.SpecialIDParam],
				entities.RootID(entities.DBUser.Name) : record[entities.RootID(entities.DBUser.Name)],
				entities.RootID(entities.DBUser.Name) : hierarch["parent_" + entities.RootID(entities.DBUser.Name)],
				entities.RootID(entities.DBEntity.Name) : hierarch[entities.RootID(entities.DBEntity.Name)],
				"description" : "hierarchical verification expected by the system, workflow is currently pending.",
				"urgency" : "medium",
				"priority" : "medium",
				entities.NAMEATTR : "hierarchical verification",
			}
			params := tool.Params{ tool.RootTableParam : entities.DBTask.Name,
				tool.RootRowsParam : tool.ReservedParam, }
			_, err := s.Domain.SuperCall( params, newTask, tool.CREATE, "CreateOrUpdate")
			if err == nil {
				params = tool.Params{ tool.RootTableParam : entities.DBNotification.Name,
					tool.RootRowsParam : tool.ReservedParam,
					tool.RootRawView : "enable", }
				s.Domain.SuperCall( params, tool.Record{ 
					entities.NAMEATTR : "Task affected : " + newTask.GetString(entities.NAMEATTR), 
					"description" : "Task is affected to you and must be treated : " + newTask.GetString(entities.NAMEATTR),
					entities.RootID(entities.DBUser.Name) : newTask.GetString(entities.RootID(entities.DBUser.Name)),
					entities.RootID(entities.DBSchema.Name) : newTask.GetString(entities.RootID(entities.DBSchema.Name)),
					entities.RootID("dest_table") : record[entities.RootID("dest_table")], }, tool.CREATE, "CreateOrUpdate")
			}
		}
	}
}

func (s *RequestService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat(results, tableName) 
}
func (s *RequestService) ConfigureFilter(tableName string) (string, string) {
	rows, ok := s.Domain.GetParams()[tool.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[tool.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != tool.ReservedParam) || (ok2 && ids != "") { return s.Domain.ViewDefinition(tableName) }
	restr := entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")" 
	return s.Domain.ViewDefinition(tableName, restr)
}
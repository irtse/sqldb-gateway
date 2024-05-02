package domain

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"strconv"
	"runtime"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)

func (d *MainService) PostTreat(results utils.Results, tableName string, isWorflow bool) utils.Results {
	// retrieve all fields from schema...
	schema, err := schserv.GetSchema(tableName)
	if err != nil { return results }
	if ids, ok := d.Params[utils.SpecialIDParam]; (ok || d.Method != utils.SELECT) { 
		d.NewDataAccess(schema.ID, strings.Split(ids, ","), d.Method) 
	}
	if !d.IsShallowed() {
		schemes, id, order, cols, addAction, readonly := d.GetViewFields(tableName, false) 
		view := schserv.ViewModel{ ID: id, Name : schema.Label, Label : schema.Label, Description : tableName + " datas", Schema : schemes,
			SchemaID: id, SchemaName: tableName, ActionPath : d.BuildPath(tableName, utils.ReservedParam), Readonly : readonly,
			Order : order, Actions : addAction,  Items : []schserv.ViewItemModel{} }
		maxConcurrent := 5
		runtime.GOMAXPROCS(maxConcurrent)
		channel := make(chan schserv.ViewItemModel, len(results))
		defer close(channel)
		defer func() {
			if err := recover(); err != nil { fmt.Printf("panic occurred: %v\n", err) }
		}()
		resResults := []utils.Results{} 
		index := 0
		for index < len(results) {
			resResults = append(resResults, results[index:min(index + maxConcurrent, len(results))])
			index += maxConcurrent
		}
		count := 0
		for _, res := range resResults {
			for i, record := range res { go d.PostTreatRecord(i + count, channel, record, tableName, cols, d.Empty, isWorflow) }
			for range res { 
				rec := <-channel
				if !rec.IsEmpty { view.Items = append(view.Items, rec) }			
			}
			count += len(res)
		}
		sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Sort < view.Items[j].Sort })
		return utils.Results{ view.ToRecord() } 
	} else { 
		res := utils.Results{}
		for _, record := range results {
			if record.GetString(schserv.NAMEKEY) == "" { res = append(res, record); continue }
			label := record.GetString(schserv.NAMEKEY)
			if record.GetString(schserv.LABELKEY) == "" { label = record.GetString(schserv.LABELKEY) }
			if record[schserv.RootID(schserv.DBSchema.Name)] != nil { // SCHEMA ? 
				sch, err := schserv.GetSchemaByID(record.GetInt(schserv.RootID(schserv.DBSchema.Name)))
				if err != nil { continue }
				schema, id, order,  _, addAction, readonly := d.GetViewFields(sch.Name, false)
				res = append(res, schserv.ViewModel{ ID: record.GetInt(utils.SpecialIDParam), 
					Name : record.GetString(schserv.NAMEKEY), Label : label, Description : tableName + " shallowed datas",  
					Path: d.BuildPath(sch.Name, utils.ReservedParam),
					Schema : schema, SchemaID: id, SchemaName: tableName, Actions : addAction,
					ActionPath : d.BuildPath(sch.Name, utils.ReservedParam), Readonly : readonly,
					Order : order, Workflow: d.BuildWorkFlow(record, tableName, isWorflow) }.ToRecord())
			} else { res = append(res, schserv.ViewModel{ ID: record.GetInt(utils.SpecialIDParam), Name : record.GetString(schserv.NAMEKEY), Label : label, 
														  Workflow : d.BuildWorkFlow(record, tableName, isWorflow) }.ToRecord()) }
		}
		return res
	} 
	return results
}

func (d *MainService) PostTreatRecord(index int, channel chan schserv.ViewItemModel, record utils.Record, tableName string, cols map[string]schserv.FieldModel, shallow bool, isWorkflow bool) {
		vals := map[string]interface{}{}; shallowVals := map[string]interface{}{}; manyPathVals := map[string]string{}; manyVals := map[string]utils.Results{}
		datapath := ""; historyPath := ""
		if !shallow { 
			schema, err := schserv.GetSchema(tableName)
			if err == nil {
				historyPath = d.BuildPath(schserv.DBDataAccess.Name, utils.ReservedParam, utils.RootOrderParam + "=access_date", 
					utils.RootDirParam + "=asc", utils.RootDestTableIDParam + "=" + record.GetString(utils.SpecialIDParam), 
					schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
			}
			vals[utils.SpecialIDParam]=record.GetString(utils.SpecialIDParam) 
		}
		for _, field := range cols {
			if strings.Contains(field.Name, schserv.DBSchema.Name) { 
				dest, ok := record[schserv.RootID("dest_table")]
				id, ok2 := record[field.Name]
				if !ok2 && !ok && dest != nil && id != nil {
					schema, err := schserv.GetSchemaByID(record.GetInt(field.Name))
					if dest != nil && err == nil {
						datapath=d.BuildPath(schema.Name, fmt.Sprintf("%v", dest))
						params := utils.Params{ utils.RootTableParam : schema.Name, utils.RootRowsParam: fmt.Sprintf("%v", dest), utils.RootShallow : "enable" }
						if _, ok := d.Params[schserv.RootID("dest_table")]; ok {
							if _, err := strconv.Atoi(strings.Replace(strings.Replace(fmt.Sprintf("%v", d.Params[schserv.RootID("dest_table")]), "%25", "", -1), "%", "", -1)); err == nil {
								params[utils.SpecialIDParam] = d.Params[schserv.RootID("dest_table")]
							} else { params[schserv.NAMEKEY] = d.Params[schserv.RootID("dest_table")] }
						}
						r, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
						if _, ok := d.Params[schserv.RootID("dest_table")]; ok && (err != nil || len(r) == 0) { 
							channel <- schserv.ViewItemModel{ IsEmpty: true }; return 
						}
						if err != nil || len(r) == 0 { continue }
						ids, _ := strconv.Atoi(fmt.Sprintf("%v",r[0][utils.SpecialIDParam]))
						shallowVals["db" + utils.RootDestTableIDParam]=utils.Record{ "id": ids, "name" : fmt.Sprintf("%v",r[0][schserv.NAMEKEY]) }
					}
					continue
				}
			}
			if record.GetString(field.Name) != "" && field.Link > 0 && !shallow { 
				link := schserv.GetTablename(fmt.Sprintf("%v", field.Link))
				params := utils.Params{ utils.RootTableParam : link, utils.RootRowsParam: record.GetString(field.Name), utils.RootShallow : "enable" }
				if strings.Contains(field.Type, "many") {
					params[utils.RootRowsParam] = utils.ReservedParam
					params[schserv.RootID(tableName)] = record.GetString(utils.SpecialIDParam)
				}
				r, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
				if err != nil || len(r) == 0 { continue }
				if !strings.Contains(field.Type, "many")  { shallowVals[field.Name]=r[0]; continue 
				} else if field.Type == schserv.MANYTOMANY.String() && !d.LowerRes {
					for _, r2 := range r {
						for field2, _ := range r2 {
							if strings.Contains(field2, tableName) || field2 != "id" || !strings.Contains(field2, "_id") { continue }
							id := strings.Replace(field2, "_id", "", -1)
							params[utils.RootTableParam] = id
							sqlFilter := "id IN (SELECT " + id + "_id FROM " + link + " WHERE " + schserv.RootID(tableName) + " = " + record.GetString(utils.SpecialIDParam) + " )"
							r, err = d.Call( params, utils.Record{}, utils.SELECT, sqlFilter)
							if err != nil || len(r) == 0 { continue }
							if _, ok := manyVals[field.Name]; !ok { manyVals[field.Name] = utils.Results{} }
							manyVals[field.Name]= append(manyVals[field.Name], r...)
						}
					}
					continue	
				} else if field.Type == schserv.ONETOMANY.String() && !d.LowerRes {
					manyPathVals[field.Name] = d.BuildPath(link, utils.ReservedParam, schserv.RootID(tableName) + "=" + record.GetString(utils.SpecialIDParam))
					continue
				}
			}
			if shallow { vals[field.Name]=nil } else if v, ok:=record[field.Name]; ok { vals[field.Name]=v }
		}
		channel <- schserv.ViewItemModel{ Values : vals,  DataPaths :  datapath, ValueShallow : shallowVals, Sort: int64(index),
			HistoryPath : historyPath, ValueMany: manyVals, ValuePathMany: manyPathVals, Workflow : d.BuildWorkFlow(record, tableName, isWorkflow), }
}

func (d *MainService) BuildWorkFlow(record utils.Record, tableName string, isWorflow bool) *schserv.WorkflowModel {
	var workflow schserv.WorkflowModel
	if !isWorflow{ return nil  }
	id := ""; requestID := ""; nexts := []string{}
	if tableName == schserv.DBWorkflow.Name { id = record.GetString(utils.SpecialIDParam)
	} else if tableName == schserv.DBRequest.Name { // TODO AS SPECIALIZED
		id = record.GetString(schserv.RootID(schserv.DBWorkflow.Name))
		requestID = record.GetString(utils.SpecialIDParam)
		workflow = schserv.WorkflowModel{ IsDismiss : record.GetString("state") == "dismiss", Current : record.GetString("current_index"), 
			IsClose : record.GetString("state") == "completed" || record.GetString("state") == "dismiss" }
	} else if tableName == schserv.DBTask.Name {
		params := utils.Params { utils.RootTableParam : schserv.DBTask.Name, utils.RootRowsParam : record.GetString(utils.SpecialIDParam), }
		t, _ := d.SuperCall( params, utils.Record{}, utils.SELECT)
		if len(t) > 0 && t[0]["nexts"] != "all" && t[0]["nexts"] != "" && t[0]["nexts"] != nil { nexts = strings.Split(t[0].GetString("nexts"), ",") }
		requestID = record.GetString(schserv.RootID(schserv.DBRequest.Name))
		workflow = schserv.WorkflowModel{ CurrentDismiss : record["state"] == "dismiss", CurrentClose : record["state"] == "completed" || record["state"] == "dismiss" }
		params = utils.Params { utils.RootTableParam : schserv.DBWorkflowSchema.Name,
			utils.RootRowsParam : record.GetString(schserv.RootID(schserv.DBWorkflowSchema.Name)), }
		schemes, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
		if err != nil || len(schemes) == 0 { return nil }
		workflow.Current = schemes[0].GetString("index")
		id = fmt.Sprintf("%v", schemes[0][schserv.RootID(schserv.DBWorkflow.Name)])
	} else { return nil }

	if id == "" { return nil }
	params := utils.Params {
		utils.RootTableParam : schserv.DBWorkflowSchema.Name,
		utils.RootRowsParam : utils.ReservedParam,
		schserv.RootID(schserv.DBWorkflow.Name) : id,
	}
	steps, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
	if err == nil && len(steps) > 0 {	
		newSteps := map[string][]schserv.WorkflowStepModel{}
		for _, step := range steps {
			index := step.GetString("index")
			if workflow.Current != "" && workflow.Current == step.GetString("index") && tableName == schserv.DBTask.Name { 
				params := utils.Params { utils.RootTableParam : schserv.DBWorkflowSchema.Name,
					utils.RootRowsParam : record.GetString(schserv.RootID(schserv.DBWorkflowSchema.Name)), }
				ownSteps, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
				if hub, ok2 := ownSteps[0]["hub"]; ok2 && err == nil && len(ownSteps) > 0 { workflow.CurrentHub=hub.(bool) }
			}
			if _, ok := newSteps[index]; !ok { newSteps[index] = []schserv.WorkflowStepModel{} }
			newStep := schserv.WorkflowStepModel{ 
				ID : step.GetInt(utils.SpecialIDParam), Name: step.GetString(schserv.NAMEKEY), Optionnal : step["optionnal"].(bool), 
				IsSet : !step["optionnal"].(bool) || slices.Contains(nexts, step.GetString("wrapped_" + schserv.RootID(schserv.DBWorkflow.Name))),
			}
			if workflow.Current != "" {
				params = utils.Params { utils.RootTableParam : schserv.DBTask.Name, utils.RootRowsParam : utils.ReservedParam,
					schserv.RootID(schserv.DBWorkflowSchema.Name) : step.GetString(utils.SpecialIDParam),
					schserv.RootID(schserv.DBRequest.Name) : requestID,
				}
				tasks, err := d.SuperCall( params, utils.Record{}, utils.SELECT)
				if err == nil && len(tasks) > 0 {
					newStep.IsClose=tasks[0]["is_close"].(bool)
					newStep.IsCurrent=tasks[0]["state"] == "pending"
					newStep.IsDismiss=tasks[0]["is_dismiss"] == "dismiss"
				}
			}
			if wrapped, ok := step["wrapped_" + schserv.RootID(schserv.DBWorkflow.Name)]; ok { 
				newStep.Workflow = d.BuildWorkFlow(utils.Record{utils.SpecialIDParam : wrapped}, schserv.DBWorkflow.Name, isWorflow)
			}
			newSteps[index] = append(newSteps[index], newStep)
		}
		workflow.ID=id
		workflow.Steps = newSteps
		return &workflow
	} else { return &schserv.WorkflowModel{ Steps : map[string][]schserv.WorkflowStepModel{}, } }
}
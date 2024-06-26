package domain

import (
	"fmt"
	"sort"
	"net/url"
	"slices"
	"strings"
	"runtime"
	"runtime/debug"
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
		view := schserv.ViewModel{ ID: id, Name : schema.Label, Label : schema.Label, Description : tableName + " data", Schema : schemes,
			SchemaID: id, SchemaName: tableName, ActionPath : d.BuildPath(tableName, utils.ReservedParam), Readonly : readonly,
			Order : order, Actions : addAction, Items : []schserv.ViewItemModel{}, Shortcuts: map[string]string{} }
		shortcuts, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBView.Name + " WHERE is_shortcut = true")
		if len(shortcuts) > 0 && err == nil {
			for _, shortcut := range shortcuts {
				view.Shortcuts[utils.GetString(shortcut, schserv.NAMEKEY)] = "#" + utils.GetString(shortcut, utils.SpecialIDParam)
			}
		}
		maxConcurrent := 5
		runtime.GOMAXPROCS(maxConcurrent)
		channel := make(chan schserv.ViewItemModel, len(results))
		defer close(channel)
		defer func() {
			if err := recover(); err != nil { fmt.Printf("panic occurred: %v\n%v\n", err, string(debug.Stack())) }
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
		if len(view.Items) == 1 { view.Readonly = view.Items[0].Readonly }
		if view.Readonly { view.Actions = []string{"get"} }
		sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Sort < view.Items[j].Sort })
		return utils.Results{ view.ToRecord() } 
	} else { 
		res := utils.Results{}
		for _, record := range results {
			if record.GetString(schserv.NAMEKEY) == "" { res = append(res, record); continue }
			label := record.GetString(schserv.NAMEKEY)
			if record.GetString(schserv.LABELKEY) == "" { label = record.GetString(schserv.LABELKEY) }
			if record[schserv.RootID(schserv.DBSchema.Name)] != nil {
				sch, err := schserv.GetSchemaByID(record.GetInt(schserv.RootID(schserv.DBSchema.Name)))
				if err != nil { continue }
				schema, id, order,  _, addAction, readonly := d.GetViewFields(sch.Name, false)
				res = append(res, schserv.ViewModel{ ID: record.GetInt(utils.SpecialIDParam), 
					Name : record.GetString(schserv.NAMEKEY), Label : label, Description : tableName + " shallowed data",  
					Path: d.BuildPath(sch.Name, utils.ReservedParam), Schema : schema, SchemaID: id, 
					SchemaName: tableName, Actions : addAction, ActionPath : d.BuildPath(sch.Name, utils.ReservedParam), 
					Readonly : readonly,
					Order : order, Workflow: d.BuildWorkFlow(record, tableName, isWorflow) }.ToRecord())
			} else { res = append(res, schserv.ViewModel{ ID: record.GetInt(utils.SpecialIDParam), Name : record.GetString(schserv.NAMEKEY), Label : label, 
														  Workflow : d.BuildWorkFlow(record, tableName, isWorflow),  }.ToRecord()) }
			
		}
		return res
	} 
	return results
}

func (d *MainService) PostTreatRecord(index int, channel chan schserv.ViewItemModel, record utils.Record, tableName string, cols map[string]schserv.FieldModel, shallow bool, isWorkflow bool) {
		vals := map[string]interface{}{}; shallowVals := map[string]interface{}{}; manyPathVals := map[string]string{}; 
		manyVals := map[string]utils.Results{}
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
				if ok2 && ok && dest != nil && id != nil {
					schema, err := schserv.GetSchemaByID(int64(id.(float64)))
					if err == nil {
						datapath=d.BuildPath(schema.Name, fmt.Sprintf("%v", dest))
						shallowVals[schserv.RootID(schserv.DBSchema.Name)]=utils.Record{ "id": schema.ID, "name" : schema.Name, "label" : schema.Label }
						p := utils.AllParams(schema.Name)
						p[utils.RootRowsParam] = fmt.Sprintf("%v", dest)
						t, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schema.Name + " WHERE id=" + fmt.Sprintf("%v", dest))
						if err == nil && len(t) > 0 { 
							shallowVals[schserv.RootID("dest_table")]=utils.Record{ 
								"id":t[0][utils.SpecialIDParam], "name" : t[0][schserv.NAMEKEY], "label" : t[0][schserv.NAMEKEY], 
								"data_ref" : "@" + fmt.Sprintf("%v", schema.ID) + ":" + fmt.Sprintf("%v", t[0][utils.SpecialIDParam]) }
						}
					}
					continue
				}
			}
			if record.GetString(field.Name) != "" && field.Link > 0 && !shallow { 
				link := schserv.GetTablename(fmt.Sprintf("%v", field.Link))
				if !strings.Contains(field.Type, "many") { 
					r, err := d.Db.QueryAssociativeArray("SELECT * FROM " + link + " WHERE id=" + record.GetString(field.Name))
					if err != nil || len(r) == 0 { continue }
					if d.PermsCheck(link, "", "", utils.SELECT) {
						shallowVals[field.Name]=utils.Record{ "id": r[0][utils.SpecialIDParam], "name" : r[0][schserv.NAMEKEY],
							"data_ref" : "@" + fmt.Sprintf("%v", field.Link) + ":" + fmt.Sprintf("%v", r[0][utils.SpecialIDParam]), }
					} else { shallowVals[field.Name]=utils.Record{ "id": r[0][utils.SpecialIDParam], "name" : r[0][schserv.NAMEKEY], } }
					if _, ok := r[0]["label"]; ok { shallowVals[field.Name].(utils.Record)["label"]=r[0]["label"] }
					continue 
				} else if field.Type == schserv.MANYTOMANY.String() && !d.LowerRes { // TODO DEBUG
					lsch, _ := schserv.GetSchemaByID(field.Link)
					for _, f := range lsch.Fields {
						if strings.Contains(f.Name, tableName) || f.Name == "id" || f.Link <= 0 { continue }
						lid, _ := schserv.GetSchemaByID(f.Link)
						sqlFilter := "SELECT id,name"
						if lid.HasField("label") { sqlFilter += ",label" }
						sqlFilter += " FROM " + lid.Name + " WHERE id IN (SELECT " + f.Name + " FROM " + link + " WHERE " + schserv.RootID(tableName) + " = " + record.GetString(utils.SpecialIDParam) + " )"
						rr, err := d.Db.QueryAssociativeArray(sqlFilter)
						if err != nil || len(rr) == 0 { continue }
						if _, ok := manyVals[field.Name]; !ok { manyVals[field.Name] = utils.Results{} }
						for _, r := range rr { manyVals[field.Name] = append(manyVals[field.Name], r) }
					}
					continue	
				} 
				if field.Type == schserv.ONETOMANY.String() && !d.LowerRes && field.Link > 0 {
					schemeLink, _ := schserv.GetSchemaByID(field.Link)
					for _, f := range schemeLink.Fields {
						if strings.Contains(f.Name, tableName) && strings.Contains(f.Name, "_id") { 
							manyPathVals[field.Name] = d.BuildPath(link, utils.ReservedParam, f.Name  + "=" + record.GetString(utils.SpecialIDParam))
							break
						}
					}
					continue
				}
			}
			if shallow { vals[field.Name]=nil } else if v, ok:=record[field.Name]; ok { vals[field.Name]=v }
		}
		if cmd, ok := d.Params[utils.RootCommandRow]; ok { 
			decodedLine, _ := url.QueryUnescape(cmd)
			matches := strings.Split(decodedLine, " as ")
			if len(matches) > 1 { vals[matches[len(matches) - 1]]=record[matches[len(matches) - 1]] }
		}
		channel <- schserv.ViewItemModel{ Values : vals,  DataPaths :  datapath, ValueShallow : shallowVals, Sort: int64(index),
			HistoryPath : historyPath, ValueMany: manyVals, ValuePathMany: manyPathVals, 
			Readonly : d.IsReadonly(tableName, record),
			Workflow : d.BuildWorkFlow(record, tableName, isWorkflow), }
}

func (d *MainService) BuildWorkFlow(record utils.Record, tableName string, isWorflow bool) *schserv.WorkflowModel {
	workflow := schserv.WorkflowModel{  Position : "0", Current : "0", Steps : map[string][]schserv.WorkflowStepModel{}, }
	if !isWorflow { return nil  }
	id := ""; requestID := ""; nexts := []string{}
	if tableName == schserv.DBWorkflow.Name { id = record.GetString(utils.SpecialIDParam)
	} else if tableName == schserv.DBRequest.Name {
		t, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBRequest.Name + " WHERE id = " + record.GetString(utils.SpecialIDParam))
		if err != nil || len(t) == 0 { return nil }
		id = fmt.Sprintf("%v", t[0][schserv.RootID(schserv.DBWorkflow.Name)])
		requestID = fmt.Sprintf("%v", t[0][utils.SpecialIDParam])
		workflow = schserv.WorkflowModel{ IsDismiss : fmt.Sprintf("%v", t[0]["state"]) == "dismiss", 
			Current : fmt.Sprintf("%v", t[0]["current_index"]), Position : fmt.Sprintf("%v", t[0]["current_index"]),
			IsClose :fmt.Sprintf("%v", t[0]["state"]) == "completed" || fmt.Sprintf("%v", t[0]["state"]) == "dismiss" }
	} else if tableName == schserv.DBTask.Name {
		id = "0"
		t, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBTask.Name + " WHERE id = " + record.GetString(utils.SpecialIDParam))
		if err != nil || len(t) == 0 { return nil }
		req, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBRequest.Name + " WHERE id = " + fmt.Sprintf("%v", t[0][schserv.RootID(schserv.DBRequest.Name)]))
		if err == nil && len(req) > 0 { 
			id = fmt.Sprintf("%v", req[0][schserv.RootID(schserv.DBWorkflow.Name)])
			workflow.Position = fmt.Sprintf("%v", req[0]["current_index"])
			workflow.IsClose = req[0]["state"] == "completed" || req[0]["state"] == "dismiss"
			workflow.IsDismiss = req[0]["state"] == "dismiss"
		}
		if t[0][schserv.RootID(schserv.DBWorkflowSchema.Name)] != nil { 
			if t[0]["nexts"] != "all" && t[0]["nexts"] != "" && t[0]["nexts"] != nil { nexts = strings.Split(fmt.Sprintf("%v", t[0]["nexts"]), ",") }
			requestID = record.GetString(schserv.RootID(schserv.DBRequest.Name))
			workflow.CurrentDismiss = record["state"] == "dismiss"
			workflow.CurrentClose = record["state"] == "completed" || record["state"] == "dismiss"			
			schemes, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBWorkflowSchema.Name + " WHERE id = " + fmt.Sprintf("%v", t[0][schserv.RootID(schserv.DBWorkflowSchema.Name)]))
			if err != nil || len(schemes) == 0 { return &workflow }
			workflow.Current = utils.GetString(schemes[0], "index")
			workflow.CurrentHub=schemes[0]["hub"].(bool)
			id = fmt.Sprintf("%v", schemes[0][schserv.RootID(schserv.DBWorkflow.Name)])
		}
	} else { return nil }
	if id == "" || id == "<nil>" { return nil }
	steps, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBWorkflowSchema.Name + " WHERE " + schserv.RootID(schserv.DBWorkflow.Name) + " = " + id)
	if err == nil && len(steps) > 0 {	
		newSteps := map[string][]schserv.WorkflowStepModel{}
		for _, step := range steps {
			index :=  fmt.Sprintf("%v", step["index"])
			if _, ok := newSteps[index]; !ok { newSteps[index] = []schserv.WorkflowStepModel{} }
			newStep := schserv.WorkflowStepModel{ 
				ID : utils.GetInt(step, utils.SpecialIDParam), Name: fmt.Sprintf("%v", step[schserv.NAMEKEY]), Optionnal : step["optionnal"].(bool), 
				IsSet : !step["optionnal"].(bool) || slices.Contains(nexts, fmt.Sprintf("%v", step["wrapped_" + schserv.RootID(schserv.DBWorkflow.Name)])),
			}
			if workflow.Current != "" {
				tasks, err := d.Db.QueryAssociativeArray("SELECT * FROM " + schserv.DBTask.Name + " WHERE " + schserv.RootID(schserv.DBWorkflowSchema.Name) + " = " + fmt.Sprintf("%v", step[utils.SpecialIDParam]) + " AND " + schserv.RootID(schserv.DBRequest.Name) + " = " + requestID)
				if err == nil && len(tasks) > 0 {
					if tasks[0]["is_close"] != nil { newStep.IsClose=tasks[0]["is_close"].(bool) } else { newStep.IsClose=false }
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
	} else { return &workflow }
}

func (d *MainService) IsReadonly(tableName string, record utils.Record) bool {
	readonly := true
	for _, meth := range []utils.Method{ utils.CREATE, utils.UPDATE } {
		if d.LocalPermsCheck(tableName, "", "", meth, record.GetString(utils.SpecialIDParam)) {
			if meth == utils.CREATE && d.Empty { readonly = false; break;
			} else if meth == utils.UPDATE { readonly = false; break; }
		}
	}
	return readonly || record["state"] == "completed" || record["state"] == "dismiss"
}
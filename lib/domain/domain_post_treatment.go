package domain

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"strconv"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	"runtime"
)

type View struct {
	Name  		 	string 						`json:"name"`
	SchemaID		int64						`json:"schema_id"`
	SchemaName   	string 						`json:"schema_name"`
	Description  	string 						`json:"description"`
	Path		 	string 						`json:"link_path"`
	Order		 	[]string 					`json:"order"`
	Schema		 	tool.Record 				`json:"schema"`
	Items		 	[]tool.Record 				`json:"items"`
	Actions		 	[]map[string]interface{} 	`json:"actions"`
}

type ViewItem struct {
	Path 	   	   string					 	`json:"link_path"`
	Values 	   	   map[string]interface{} 	    `json:"values"`
	DataPaths  	   string				        `json:"data_path"`
	ValueShallow   map[string]interface{}		`json:"values_shallow"`
	ValueMany      map[string]tool.Results		`json:"values_many"`
	ValuePathMany  map[string]string			`json:"values_path_many"`
	Workflow  	   map[string]interface{}		`json:"workflow"`
}

func (d *MainService) PostTreat(results tool.Results, tableName string) tool.Results {
	// retrive all fields from schema...
	var view View
	if !d.IsShallowed() {
		schemes, id, order, cols, addAction := d.GetScheme(tableName, false) 
		if ids, ok := d.Params[tool.SpecialIDParam]; ok { d.DataAccess(id, strings.Split(ids, ","), false) }
		view = View{ Name : tableName, 
					 Description : tableName + " datas",  
					 Path : "", 
					 Schema : schemes, 
					 Order : order,
					 SchemaID: id,
					 SchemaName: tableName, 
					 Actions : []map[string]interface{}{},  Items : []tool.Record{} }	
		res := tool.Results{} 
		runtime.GOMAXPROCS(5)
		channel := make(chan tool.Record, len(results))
		defer close(channel)
		maxConcurrent := 5
		resResults := []tool.Results{} 
		index := 0
		for index < len(results) {
			resResults = append(resResults, results[index:min(index + maxConcurrent, len(results))])
			index += maxConcurrent
		}
		count := 0
		for _, res := range resResults {
			for i, record := range res { go d.PostTreatRecord(i + count, channel, record, tableName, cols, d.Empty) }
			for range res { 
				rec := <-channel
				if rec == nil { continue }
				view.Items = append(view.Items, rec) 			
			}
			count += len(res)
		}
		sort.SliceStable(view.Items, func(i, j int) bool {
			return view.Items[i]["sort"].(int) < view.Items[j]["sort"].(int)
		})
		for _, item := range view.Items { delete(item, "sort") }
		r := tool.Record{}
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &r)
		r["action_path"] = "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + tool.ReservedParam
		r["actions"]=[]string{}
		for _, meth := range []tool.Method{ tool.SELECT, tool.CREATE, tool.UPDATE, tool.DELETE } {
			if d.Empty && meth != tool.CREATE { continue }
			if d.PermsCheck(tableName, "", "", meth) || slices.Contains(addAction, meth.Method()) { 
				r["actions"]=append(r["actions"].([]string), meth.Method())
			} else if meth == tool.UPDATE { r["readonly"] = true }
		} 
		res = append(res, r)
		return res
	} else { 
		res := tool.Results{}
		for _, record := range results {
			if n, ok := record[entities.NAMEATTR]; ok {
				label := fmt.Sprintf("%v", n)
				if l, ok2 := record["label"]; ok2 { label = fmt.Sprintf("%v", l) }
				if record[entities.RootID(entities.DBSchema.Name)] != nil { // SCHEMA ? 
					schemas, err := d.Schema(record, true)
					actionPath := "/" + tool.MAIN_PREFIX + "/" + tableName + "?rows=" + tool.ReservedParam
					actions := []string{}
					readonly := false
					if err == nil || len(schemas) > 0 { 
						schema, id, order,  _, addAction := d.GetScheme(schemas[0].GetString(entities.NAMEATTR), false)
						for _, meth := range []tool.Method{ tool.SELECT, tool.CREATE, tool.UPDATE, tool.DELETE } {
							if d.PermsCheck(schemas[0].GetString(entities.NAMEATTR), "", "", meth) || slices.Contains(addAction, meth.Method()) { 
								actions=append(actions, meth.Method())
							} else if meth == tool.UPDATE { readonly = true 
							} else if meth == tool.CREATE && d.Empty { readonly = true }
						} 
						res = append(res, tool.Record{ 
							tool.SpecialIDParam : record[tool.SpecialIDParam],
							entities.NAMEATTR : n,
							"label": label, 
							"order" : order,
							"schema_id" : id,
							"actions" : actions,
							"action_path" : actionPath,
							"readonly" : readonly,
							"workflow" : d.GetWorkFlow(record, tableName),
							"link_path" : "/" + tool.MAIN_PREFIX + "/" + schemas[0].GetString(entities.NAMEATTR) + "?rows=" + tool.ReservedParam,
							"schema_name" : schemas[0].GetString(entities.NAMEATTR),
							"schema" : schema, })	
						continue
					}	
				}
				res = append(res, tool.Record{ 
					tool.SpecialIDParam : record[tool.SpecialIDParam],
					entities.NAMEATTR : n, 
					"label": label,
					"workflow" : d.GetWorkFlow(record, tableName),
				})	
			} else { res = append(res, record) }
		}
		return res
	} 
	return results
}

func (d *MainService) GetWorkFlow(record tool.Record, tableName string) tool.Record {
	id := ""; requestID := ""
	nexts := []string{}
	workflow := tool.Record{}
	if tableName == entities.DBWorkflow.Name { id = record.GetString(tool.SpecialIDParam)
	} else if tableName == entities.DBRequest.Name {
		id = record.GetString(entities.RootID(entities.DBWorkflow.Name))
		requestID = record.GetString(tool.SpecialIDParam)
		workflow["is_dismiss"]=record.GetString("state") == "dismiss"
		workflow["current"] = record.GetString("current_index")
		workflow["is_close"]=record.GetString("state") == "completed" || record.GetString("state") == "dismiss"
	} else if tableName == entities.DBTask.Name {
		params := tool.Params {
			tool.RootTableParam : entities.DBTask.Name,
			tool.RootRowsParam : record.GetString(tool.SpecialIDParam),
		}
		t, _ := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if len(t) > 0 && t[0]["nexts"] != "all" && t[0]["nexts"] != "" && t[0]["nexts"] != nil { nexts = strings.Split(t[0].GetString("nexts"), ",") }
		requestID = record.GetString(entities.RootID(entities.DBRequest.Name))
		workflow["current_dismiss"]=record["state"] == "dismiss"
		workflow["current_close"]=record["state"] == "completed" && record["state"] == "dismiss"
		params = tool.Params {
			tool.RootTableParam : entities.DBWorkflowSchema.Name,
			tool.RootRowsParam : record.GetString(entities.RootID(entities.DBWorkflowSchema.Name)),
		}
		schemes, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(schemes) == 0 { return nil }
		workflow["current"] = schemes[0].GetString("index")
		id = fmt.Sprintf("%v", schemes[0][entities.RootID(entities.DBWorkflow.Name)])
	} else { return nil }	
	if id == "" { return nil }
	params := tool.Params {
		tool.RootTableParam : entities.DBWorkflowSchema.Name,
		tool.RootRowsParam : tool.ReservedParam,
		entities.RootID(entities.DBWorkflow.Name) : id,
	}
	steps, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
	if err == nil && len(steps) > 0 {	
		newSteps := map[int][]tool.Record{}
		for _, step := range steps {
			index := step.GetInt("index")
			if workflow["current"] != "" && workflow["current"] == step.GetString("index") && tableName == entities.DBTask.Name { 
				params := tool.Params {
					tool.RootTableParam : entities.DBWorkflowSchema.Name,
					tool.RootRowsParam : record.GetString(entities.RootID(entities.DBWorkflowSchema.Name)),
				}
				ownSteps, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err == nil && len(ownSteps) > 0 {
					if hub, ok2 := ownSteps[0]["hub"]; ok2 { workflow["current_hub"]=hub.(bool) }
				}
			}
			if _, ok := newSteps[index]; !ok { newSteps[index] = []tool.Record{} }
			newStep := tool.Record{
				tool.SpecialIDParam : step.GetString(tool.SpecialIDParam),
				entities.NAMEATTR : step.GetString(entities.NAMEATTR),
				"optionnal" : step["optionnal"],
			}
			newStep["is_set"]= !step["optionnal"].(bool) || slices.Contains(nexts, step.GetString("wrapped_" + entities.RootID(entities.DBWorkflow.Name)))
			if workflow["current"] != "" {
				params = tool.Params {
					tool.RootTableParam : entities.DBTask.Name,
					tool.RootRowsParam : tool.ReservedParam,
					entities.RootID(entities.DBWorkflowSchema.Name) : step.GetString(tool.SpecialIDParam),
					entities.RootID(entities.DBRequest.Name) : requestID,
				}
				tasks, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err == nil && len(tasks) > 0 {
					newStep["is_close"]=tasks[0]["is_close"]
					newStep["is_current"]=tasks[0]["state"] == "pending"
					newStep["is_dismiss"]=tasks[0]["is_dismiss"] == "dismiss"
				}
			}
			if wrapped, ok := step["wrapped_" + entities.RootID(entities.DBWorkflow.Name)]; ok { 
				newStep["workflow"] = d.GetWorkFlow(tool.Record{tool.SpecialIDParam : wrapped}, entities.DBWorkflow.Name)
			}
			newSteps[index] = append(newSteps[index], newStep)
		}
		workflow["id"]=id
		workflow["steps"] = newSteps
		return workflow
	} else { return tool.Record{ "steps" : map[string]interface{}{}, } }
}

func (d *MainService) PostTreatRecord(index int, channel chan tool.Record, record tool.Record, tableName string,  cols map[string]entities.SchemaColumnEntity, shallow bool) {
		vals := map[string]interface{}{}
		shallowVals := map[string]interface{}{}
		manyPathVals := map[string]string{}
		manyVals := map[string]tool.Results{}
		datapath := ""
		if !shallow { vals[tool.SpecialIDParam]=fmt.Sprintf("%v", record[tool.SpecialIDParam]) }
		for _, field := range cols {
			if strings.Contains(field.Name, entities.DBSchema.Name) { 
				dest, ok := record[entities.RootID("dest_table")]
				id, ok2 := record[field.Name]
				if ok2 && ok && dest != nil && id != nil {
					schemas, err := d.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : id }, true)
					if err != nil || len(schemas) == 0 { continue }
					if dest != nil {
						datapath=d.BuildPath(fmt.Sprintf("%v",schemas[0][entities.NAMEATTR]), fmt.Sprintf("%v", dest))
						params := tool.Params{ tool.RootTableParam : fmt.Sprintf("%v",schemas[0][entities.NAMEATTR]), tool.RootRowsParam: fmt.Sprintf("%v", dest), tool.RootShallow : "enable" }
						if _, ok := d.Params[entities.RootID("dest_table")]; ok {
							if _, err := strconv.Atoi(strings.Replace(strings.Replace(fmt.Sprintf("%v", d.Params[entities.RootID("dest_table")]), "%25", "", -1), "%", "", -1)); err == nil {
								params[tool.SpecialIDParam] = d.Params[entities.RootID("dest_table")]
							} else { params[entities.NAMEATTR] = d.Params[entities.RootID("dest_table")] }
						}
						r, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
						if _, ok := d.Params[entities.RootID("dest_table")]; ok && (err != nil || len(r) == 0) { 
							channel <- nil
							return 
						}
						if err != nil || len(r) == 0 { continue }
						ids, _ := strconv.Atoi(fmt.Sprintf("%v",r[0][tool.SpecialIDParam]))
						shallowVals["db" + tool.RootDestTableIDParam]=tool.Record{ 
							"id": ids,
							"name" : fmt.Sprintf("%v",r[0][entities.NAMEATTR]),
						}
					}
				}
			}
			if f, ok:= record[field.Name]; ok && field.Link != "" && f != nil && !shallow && !strings.Contains(field.Type, "many") { 
				params := tool.Params{ tool.RootTableParam : field.Link, tool.RootRowsParam: fmt.Sprintf("%v", f), tool.RootShallow : "enable" }
				r, err := d.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(r) == 0 { continue }
				shallowVals[field.Name]=r[0]
				continue
			}
			
			if field.Link != "" && !shallow && !d.LowerRes && strings.Contains(field.Type, "manytomany") { 
				params := tool.Params{ tool.RootTableParam : field.Link, tool.RootRowsParam: tool.ReservedParam, tool.RootShallow : "enable",
									   entities.RootID(tableName) : record.GetString(tool.SpecialIDParam), }
				r, err := d.Call( params, tool.Record{}, tool.SELECT, "Get")
				if err != nil || len(r) == 0 { continue }
				ids := []string{}
				for _, r2 := range r {
					for field2, _ := range r2 {
						if !strings.Contains(field2, tableName) && field2 != "id" && strings.Contains(field2, "_id") {
							if !slices.Contains(ids, strings.Replace(field2, "_id", "", -1)) {
								ids = append(ids, strings.Replace(field2, "_id", "", -1))
							}
						}
					}
				}
				for _, id := range ids {
					params = tool.Params{ tool.RootTableParam : id, tool.RootRowsParam: tool.ReservedParam, 
						                  tool.RootShallow : "enable", tableName + "_id": record.GetString(tool.SpecialIDParam) }
					sqlFilter := "id IN (SELECT " + id + "_id FROM " + field.Link + " WHERE " + tableName + "_id = " + record.GetString(tool.SpecialIDParam) + " )"
					r, err = d.Call( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
					if err != nil || len(r) == 0 { continue }
					if _, ok := manyVals[field.Name]; !ok { manyVals[field.Name] = tool.Results{} }
					manyVals[field.Name]= append(manyVals[field.Name], r...)
				}
				continue
			}
			if field.Link != "" && !shallow && strings.Contains(field.Type, "onetomany") && !d.LowerRes { 
				manyPathVals[field.Name] = "/" + tool.MAIN_PREFIX + "/" + field.Link + "?" + tool.RootRowsParam + "=" + tool.ReservedParam + "&" + tableName + "_id=" + record.GetString(tool.SpecialIDParam)
				continue
			}
			if shallow { vals[field.Name]=nil } else if v, ok:=record[field.Name]; ok { vals[field.Name]=v }
		}
		view := ViewItem{ Values : vals, Path : "", 
			DataPaths :  datapath, ValueShallow : shallowVals, 
			ValueMany: manyVals, ValuePathMany: manyPathVals,
			Workflow : d.GetWorkFlow(record, tableName), }
		var newRec tool.Record
		b, _ := json.Marshal(view)
		json.Unmarshal(b, &newRec)
		newRec["sort"]=index
		channel <- newRec
}



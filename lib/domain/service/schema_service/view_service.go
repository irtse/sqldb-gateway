package schema_service

import (
	"fmt"
	"sort"
	"slices"
	"strings"
	"runtime"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
type ViewService struct { 
	utils.AbstractSpecializedService 
	infrastructure.InfraSpecializedService
}

func (s *ViewService) Entity() utils.SpecializedServiceInfo { return schserv.DBView }
func (s *ViewService) ConfigureFilter(tableName string, innerestr... string) (string, string, string, string) { 
	restr, _, _, _ := s.Domain.ViewDefinition(tableName, innerestr...)
	return restr, "", "", ""
}	
func (s *ViewService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results { 
	if len(results) == 0 { return results }
	res := utils.Results{}
	runtime.GOMAXPROCS(5)
	channel := make(chan utils.Record, len(results))
	for _, record := range results { go s.PostTreat(record, channel, dest_id...) }
	for range results {
		rec := <-channel
		if rec != nil { res = append(res, rec)  }
	}
	sort.SliceStable(res, func(i, j int) bool{  return int64(res[i]["index"].(float64)) <= int64(res[j]["index"].(float64))  })
	return res
}

func (s *ViewService) PostTreat(record utils.Record, channel chan utils.Record, dest_id... string) {
	id := ""
	schema, err := schserv.GetSchemaByID(utils.GetInt(record, schserv.RootID(schserv.DBSchema.Name)))
	if err != nil { channel <- nil; return }
	if record["is_empty"] != nil { s.Domain.SetEmpty(record["is_empty"].(bool)) }
	rec := utils.Record{ "id": record["id"], "name" : record["name"], "label" : record["name"], "description" : record["description"], "is_empty" : record["is_empty"],
						"index" : record["index"], "is_list" : record["is_list"], "readonly" : record["readonly"], "category" : record["category"],
						"filter_path" : "/" + utils.MAIN_PREFIX + "/" + schserv.DBFilter.Name + "?" + utils.RootRowsParam + "=" + utils.ReservedParam + "&" + utils.RootShallow + "=enable&" + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID), }	
	u, _ := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBUser.Name + " WHERE name='" + s.Domain.GetUser() +"' OR email='" + s.Domain.GetUser()  + "'")
	if len(u) > 0 { 
		rec["favorize_body"] = utils.Record{ schserv.RootID(schserv.DBUser.Name) : u[0][utils.SpecialIDParam], schserv.RootID(schserv.DBView.Name) : record[utils.SpecialIDParam] }
		rec["favorize_path"] = "/" + utils.MAIN_PREFIX + "/" + schserv.DBViewAttribution.Name + "?" + utils.RootRowsParam + "=" + utils.ReservedParam
		u, _ = s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBViewAttribution.Name + " WHERE " + schserv.RootID(schserv.DBUser.Name) + "=" + fmt.Sprintf("%v", u[0][utils.SpecialIDParam]) + " AND " + schserv.RootID(schserv.DBView.Name) + "=" + fmt.Sprintf("%v", record[utils.SpecialIDParam]))
		rec["is_favorize"] = len(u) > 0
	}
	if record["is_list"] != nil { s.Domain.SetLowerRes(record["is_list"].(bool)) } else { s.Domain.SetLowerRes(false) }
	if record["own_view"] != nil && record["own_view"].(bool) { s.Domain.SetOwn(true) }
	for _, dest := range dest_id {
		if id == "" { id = dest } else { id = "," + dest  }
	}
	path := "/" + utils.MAIN_PREFIX + "/" + schema.Name
	params := utils.AllParams(schema.Name)
	if id != "" { params[utils.RootRowsParam] = id }
	for k, v := range s.Domain.GetParams() {
		if _, ok := params[k]; !ok { 
			if k != "new" && !strings.Contains(k,"dest_table") && k != "id" {
				if k == utils.SpecialSubIDParam { params[utils.SpecialIDParam] = v } else if _, ok := params[k]; !ok { params[k] = v }
			}
		}
	}
	datas := utils.Results{utils.Record{}}
	d := utils.Results{}
	filter := ""; viewFilter := ""
	if record[schserv.RootID(schserv.DBFilter.Name)] != nil { filter = fmt.Sprintf("%v", record[schserv.RootID(schserv.DBFilter.Name)])}
	if record["view_" + schserv.RootID(schserv.DBFilter.Name)] != nil { viewFilter = fmt.Sprintf("%v", record["view_" +schserv.RootID(schserv.DBFilter.Name)])}
	sqlFilter, view, _, dir, _ := s.Domain.GetFilter(filter, viewFilter, utils.GetString(record, schserv.RootID(schserv.DBSchema.Name)))
	if view != "" { params[utils.RootColumnsParam] = view }
	if dir != "" { params[utils.RootDirParam] = dir }
	for k, p := range s.Domain.GetParams() { 
		if k == utils.RootRowsParam || k == utils.SpecialIDParam || k == utils.RootTableParam { continue }
		if k == utils.SpecialSubIDParam { params[utils.RootRowsParam] = p; continue }
		params[k]=p
	}
	if s.Domain.GetParams()[utils.RootFilterNewState] != "" { params[utils.RootFilterNewState] = s.Domain.GetParams()[utils.RootFilterNewState] }
	rec["new"] = []string{}
	if !s.Domain.GetEmpty() { d, _ = s.Domain.SpecialSuperCall( params, utils.Record{}, utils.SELECT, sqlFilter) }
	if  record["is_list"] != nil && record["is_list"].(bool) { 
		SQLrestriction, _, _, _ := s.Domain.ViewDefinition(schema.Name, sqlFilter)
		rec["new"], rec["max"] = s.Domain.CountNewDataAccess(schema.Name, SQLrestriction, params)
	}
	if !s.Domain.GetEmpty() {
		datas = utils.Results{}
		if new, ok := s.Domain.GetParams()["new"]; ok && new == "enable" {
			for _, data := range d {
				if slices.Contains(rec["new"].([]string), data.GetString("id")) { datas = append(datas, data) }
			}
		} else { datas = d }
	}
	if !s.Domain.IsShallowed() {
		treated := s.Domain.PostTreat(datas, schema.Name, false)
		if len(treated) > 0 {
			for k, v := range treated[0] { 
				if k == "items" && v != nil {
					for _, item := range v.([]interface{}) {
						values := item.(map[string]interface{})["values"]
						if list, ok := record["is_list"]; ok && list.(bool) && len(path) > 0 && path[:1] == "/" {
							nP := ""
							if strings.Contains(path, schserv.DBView.Name) { nP =  "/" + utils.MAIN_PREFIX + path + "&" + utils.RootDestTableIDParam + "=" + fmt.Sprintf("%v", values.(map[string]interface{})[utils.SpecialIDParam])
							} else { nP =  "/" + utils.MAIN_PREFIX + "/" + schema.Name + "?" + utils.RootRowsParam + "=" + fmt.Sprintf("%v", values.(map[string]interface{})[utils.SpecialIDParam]) }
							item.(map[string]interface{})["link_path"] = nP
							item.(map[string]interface{})["data_path"] = ""
						}	
					}
					rec[k]=v 
				} else if k == "schema" && v != nil { 
					newV := map[string]interface{}{}
					for fieldName, field := range v.(map[string]interface{}) {
						if  fieldName != schserv.RootID(schserv.DBWorkflow.Name) && schema.Name == schserv.DBRequest.Name && record["is_empty"].(bool) { continue }
						if view, ok := params[utils.RootColumnsParam]; !ok || view == "" || strings.Contains(view, fieldName) { 
							field.(map[string]interface{})["active"] = true
						} else { field.(map[string]interface{})["active"] = false }
						newV[fieldName] = field 
					}
					rec[k] = newV
				} else if k == "shortcuts" && v != nil { 
					shorts := map[string]interface{}{}
					for shortcut, ss := range v.(map[string]interface{}) {
						if strings.Contains(shortcut, record.GetString(schserv.NAMEKEY)) { continue }
						shorts[shortcut] = ss
					}
					rec[k]=shorts
				} else if rec[k] == nil || rec[k] == "" { rec[k]=v }
			} 
		}	
	}
	if view != "" { rec["order"] = strings.Split(view, ",") }
	rec["link_path"]=s.Domain.BuildPath(fmt.Sprintf(schserv.DBView.Name), fmt.Sprintf("%v", record[utils.SpecialIDParam]))
	channel <- rec
}
func (s *ViewService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) { 
	if s.Domain.GetMethod() != utils.DELETE {
		rec, err := s.Domain.ValidateBySchema(record, tablename)
		if err != nil && !s.Domain.GetAutoload() { return rec, err, false } else { rec = record }
		return rec, nil, false 
	}
	return record, nil, true
}
// TODO : filter service ? (not in the same service) on own
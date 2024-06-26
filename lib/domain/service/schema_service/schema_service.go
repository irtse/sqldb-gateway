package schema_service

import (
	"fmt"
	"slices"
	"errors"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)

type SchemaService struct { utils.SpecializedService }
// ADD IN THE FUTURE IN CACHE
func (s *SchemaService) Entity() utils.SpecializedServiceInfo { return schserv.DBSchema }
func (s *SchemaService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) { 
	if s.Domain.GetMethod() == utils.DELETE { 
		tableName := schserv.GetTablename(s.Domain.GetParams()[utils.RootRowsParam])
		if tableName == "" { return record, errors.New("no data name given..."), false }
		schema, err := schserv.GetSchema(tableName)
		if !schserv.IsRootDB(schema.Name) { err = errors.New("Cannot delete root schema field") }
		return record, err, false 
	}
	rec, err := s.Domain.ValidateBySchema(record, tablename)
	if err != nil && !s.Domain.GetAutoload() { return rec, err, false } else { rec = record }
	return rec, nil, false 
}
func (s *SchemaService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { 
	s.Domain.SetIsCustom(true)
	s.Domain.SuperCall( utils.Params{ utils.RootTableParam : schserv.DBPermission.Name, 
		utils.RootRowsParam : utils.ReservedParam,  schserv.NAMEKEY : "%" + tableName + "%" },  utils.Record{},  utils.DELETE)
}
func (s *SchemaService) UpdateRowAutomation(datas []map[string]interface{}, record map[string]interface{}) {
	if datas != nil { schserv.LoadCache(utils.ReservedParam, s.Domain.GetDb()) }
	for role, mainPerms := range schserv.MAIN_PERMS {
		for _, level := range []string{schserv.LEVELOWN, schserv.LEVELNORMAL} {
			rec := utils.Record{ schserv.NAMEKEY : fmt.Sprintf("%v", record[schserv.NAMEKEY]) + ":" + level + ":" + role , }
			for perms, value := range mainPerms { rec[perms]=value }
			rec[utils.SELECT.String()]=level
			s.Domain.SuperCall(utils.AllParams(schserv.DBPermission.Name), rec, utils.CREATE)
		}
	}
}
func (s *SchemaService) WriteRowAutomation(record map[string]interface{}, tableName string) { 
	s.Domain.SuperCall(utils.Params{ utils.RootTableParam : record[schserv.NAMEKEY].(string), }, 
		utils.Record{ schserv.NAMEKEY : record[schserv.NAMEKEY], "fields": []interface{}{} }, utils.CREATE)
	schema := schserv.SchemaModel{}.Deserialize(record)
	name := schema.Label
	if schema.Name != schserv.DBDataAccess.Name {
		if !slices.Contains([]string{ schserv.DBView.Name, schserv.DBRequest.Name, schserv.DBTask.Name,  schserv.DBFilter.Name, schserv.DBFilterField.Name, schserv.DBViewAttribution.Name, schserv.DBNotification.Name }, schema.Name) {
			count, err := s.Domain.GetDb().QueryAssociativeArray("SELECT COUNT(*) as count FROM " + schserv.DBView.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
			index := 2
			if err == nil && len(count) > 0  && (int(count[0]["result"].(float64)) + 1) > 1 { index = int(count[0]["result"].(float64)) + 1 }
			cat := "global data"
			if fmt.Sprintf("%v",record["name"])[:2] == "db" { cat = "technical data" }
			newView := utils.Record{ schserv.NAMEKEY : name, "indexable" : true, "description": "View description for " + name + " datas.", 
				"category" : cat, "is_empty": false, "index": index, "is_list": true, "readonly": false, schserv.RootID(schserv.DBSchema.Name) : schema.ID }
			s.Domain.SuperCall(utils.AllParams(schserv.DBView.Name), newView, utils.CREATE)
		}
		if !slices.Contains([]string{ schserv.DBTask.Name, schserv.DBRequest.Name, schserv.DBFilter.Name, schserv.DBFilterField.Name, schserv.DBViewAttribution.Name, schserv.DBNotification.Name }, schema.Name) {
			newWF := utils.Record{ schserv.NAMEKEY : "create " + name, "description": "new " + name + " workflow", schserv.RootID(schserv.DBSchema.Name) : schema.ID }
			s.Domain.SuperCall(utils.AllParams(schserv.DBWorkflow.Name), newWF, utils.CREATE)
		}
	}
	s.UpdateRowAutomation(nil, record)
}
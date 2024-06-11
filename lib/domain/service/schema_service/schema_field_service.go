package schema_service

import (
	"fmt"
	"strings"
	"strconv"
	"math/rand"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)

type SchemaFields struct { utils.SpecializedService }
func (s *SchemaFields) Entity() utils.SpecializedServiceInfo {return schserv.DBSchemaField }
func (s *SchemaFields) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, bool, bool) {
	if s.Domain.GetMethod() == utils.DELETE { 
		i, err := strconv.Atoi(s.Domain.GetParams()[utils.RootRowsParam])
		if err != nil { return record, false, false }
		schema, err := schserv.GetSchemaByFieldID(int64(i))
		if err != nil { return record, false, false }
		return record, !schserv.IsRootDB(schema.Name), false 
	}
	if typ, ok := record[schserv.TYPEKEY]; ok && strings.Contains(fmt.Sprintf("%v", typ), "enum") {
		typ2 := strings.Replace(fmt.Sprintf("%v", typ), " ", "", -1)
		typ2 = strings.Replace(typ2, "'", "", -1)
		typ2 = strings.Replace(typ2, "(", "__", -1)
		typ2 = strings.Replace(typ2, ",", "_", -1)
		typ2 = strings.Replace(typ2, ")", "", -1)
		record[schserv.TYPEKEY] = strings.ToLower(typ2)
	}
	if label, ok := record[schserv.LABELKEY]; !ok || label == "" {
		record[schserv.LABELKEY] = strings.Replace(fmt.Sprintf("%v", record[schserv.NAMEKEY]), "_", " ", -1)
	}
	rec, err := s.Domain.ValidateBySchema(record, tablename)
	if err != nil && !s.Domain.GetAutoload() { return rec, false, false } else { rec = record } 
	return rec, true, true
}
func (s *SchemaFields) WriteRowAutomation(record map[string]interface{}, tableName string) { 
	schema, err := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
	if err != nil { return }
	for role, mainPerms := range schserv.MAIN_PERMS {
			read_levels := []string{schserv.LEVELNORMAL}
			if level, ok := record["read_level"]; ok && level != "" && level != schserv.LEVELOWN {
				read_levels = append(read_levels, strings.Replace(fmt.Sprintf("%v", level), "'", "", -1))
			}
			for _, l := range read_levels {
				rec := map[string]interface{}{ 
					schserv.NAMEKEY : schema.Name + ":" + fmt.Sprintf("%v", record[schserv.NAMEKEY]) + ":" + l + ":" + role, 
				}
				for perms, value := range mainPerms { rec[perms]=value }
				rec[utils.SELECT.String()]=l
				s.Domain.SuperCall(utils.AllParams(schserv.DBPermission.Name), rec, utils.CREATE)
			}
	}
	s.Domain.SuperCall(utils.Params{ utils.RootTableParam : schema.Name, 
		utils.RootColumnsParam : fmt.Sprintf("%v", record[schserv.NAMEKEY])}, record, utils.CREATE)
	schserv.LoadCache(schema.Name, s.Domain.GetDb())
	if record[schserv.NAMEKEY] == schserv.RootID(schserv.DBUser.Name) || record[schserv.NAMEKEY] == schserv.RootID(schserv.DBEntity.Name)  {
		r := rand.New(rand.NewSource(9999999999))
		newView := utils.Record{ schserv.NAMEKEY : "my " + schema.Name, "indexable" : true, 
				"description": "View description for my " + schema.Name + " datas.", 
				"category" : "my data", "is_empty": false, "index": r.Int(), "is_list": true, "readonly": false, 
				"own_view" : true, schserv.RootID(schserv.DBSchema.Name) : schema.ID }
		s.Domain.SuperCall(utils.AllParams(schserv.DBView.Name), newView, utils.CREATE)
	}
}
func (s *SchemaFields) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {
	for _, r := range results {
		schema, err := schserv.GetSchemaByID(r[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if err != nil { continue }
		for role, mainPerms := range schserv.MAIN_PERMS {
			read_levels := []string{schserv.LEVELNORMAL}
			if level, ok := record["read_level"]; ok && level != "" && level != schserv.LEVELOWN {
				read_levels = append(read_levels,strings.Replace( fmt.Sprintf("%v", level), "'", "", -1))
			}
			for _, l := range read_levels {
				rec := map[string]interface{}{ 
					schserv.NAMEKEY : schema.Name + ":" + fmt.Sprintf("%v", record[schserv.NAMEKEY]) + ":" + l + ":" + role, 
				}
				for perms, value := range mainPerms { rec[perms]=value }
				rec[utils.SELECT.String()]=l
				s.Domain.SuperCall(utils.AllParams(schserv.DBPermission.Name), rec, utils.CREATE)
			}
		}
		newRecord := utils.Record{}
		for k, v := range record { newRecord[k] = v }
		newRecord[schserv.TYPEKEY] = r[schserv.TYPEKEY]
		newRecord[schserv.NAMEKEY] = r[schserv.NAMEKEY]
		s.Domain.SuperCall(utils.Params{ utils.RootTableParam : schema.Name, 
			utils.RootColumnsParam: r[schserv.NAMEKEY].(string) }, newRecord, utils.UPDATE)
		schserv.LoadCache(schema.Name, s.Domain.GetDb())
	}
	
}
func (s *SchemaFields) DeleteRowAutomation(results []map[string]interface{}, tableName string) { 
	for _, record := range results { 
		schema, err := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if err == nil { continue }
	    s.Domain.SuperCall( utils.Params{ utils.RootTableParam : schema.Name, 
			utils.RootColumnsParam: record[schserv.NAMEKEY].(string) }, utils.Record{},  utils.DELETE)
		s.Domain.SuperCall(utils.Params{ utils.RootTableParam : schserv.DBPermission.Name, 
			utils.RootRowsParam : utils.ReservedParam, schserv.NAMEKEY : "%" + tableName + ":" + fmt.Sprintf("%v", record[schserv.NAMEKEY]) + "%" }, 
			utils.Record{ },  utils.DELETE)
		if schema.HasField(schserv.RootID(schserv.DBUser.Name)) || schema.HasField(schserv.RootID(schserv.DBEntity.Name)) {
			p := utils.AllParams(schserv.DBView.Name)
			p[schserv.NAMEKEY] = "my " + schema.Name
			s.Domain.SuperCall(p, utils.Record{}, utils.DELETE)
		}
		schserv.LoadCache(schema.Name, s.Domain.GetDb())
	}
}	
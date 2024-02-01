package schema_service

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type ActionService struct { tool.AbstractSpecializedService }

func (s *ActionService) Entity() tool.SpecializedServiceInfo { return entities.DBAction }
func (s *ActionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	schemas, err := s.Domain.Schema(tool.Record{ entities.RootID(entities.DBSchema.Name) : record[entities.RootID("from")].(int64) })	
	if err != nil && len(schemas) == 0 { return record, false }
	if to, ok := record[entities.RootID("to")]; ok {
		schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	if link, ok := record[entities.RootID("link")]; ok {
		schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	return record, true
}
func (s *ActionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *ActionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ActionService) WriteRowAutomation(record tool.Record, tableName string) { }

func (s *ActionService) PostTreatment(results tool.Results, tablename string) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		names := []string{}
		schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : record[entities.RootID("from")]})
		if err != nil || len(schemas) == 0 { continue }
		link_path := "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		names = append(names, schemas[0][entities.NAMEATTR].(string))
		if to, ok := record[entities.RootID("to")]; ok && to != nil {
			schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			link_path += "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		}
		link_path, _ = s.Domain.GeneratePathFilter(link_path, record, nil)
		if !strings.Contains(link_path, "?") { link_path +="?rawview=enable"
	    } else { link_path +="&rawview=enable" }
		if parameter, ok := record["parameters"]; ok && parameter != nil {
			for _, par := range strings.Split(fmt.Sprintf("%v", parameter), ",") {
				link_path += "&" + par +"=%" + par + "%"
			}
		} else { record["parameters"] = "" }
		newRec := tool.Record{ "name" : fmt.Sprintf("%v", record["name"]), 
	                           "description" : fmt.Sprintf("%v",  record["description"]),
							   "method" : fmt.Sprintf("%v", record["method"]),
							   "is_view" : strings.Contains(link_path, entities.DBView.Name),
							   "parameters" : fmt.Sprintf("%v", record["parameters"]),
							   "link_path" : link_path }
		sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
		sqlFilter += entities.DBSchema.Name + " WHERE name='" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR]) + "')"
		// retrive all fields from schema...
		params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
		                       tool.RootRowsParam: tool.ReservedParam, 
						       tool.RootSQLFilterParam: sqlFilter }
		schemas, err = s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get")
		if err != nil || len(schemas) == 0 { continue }
		schemes := map[string]interface{}{}
		for _, r := range schemas {
			var scheme entities.SchemaColumnEntity
			var shallowField entities.ShallowSchemaColumnEntity
			b, _ := json.Marshal(r)
			json.Unmarshal(b, &scheme)
			json.Unmarshal(b, &shallowField)
			schemes[scheme.Name]=shallowField
		}
		newRec["schema"]=schemes
		res = append(res, newRec)
	}
	return res 
}

func (s *ActionService) ConfigureFilter(tableName string, params tool.Params) (string, string) { 
	return s.Domain.ViewDefinition(tableName, params)
}	
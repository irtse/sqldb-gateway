package schema_service

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type ActionService struct { tool.AbstractSpecializedService }

func (s *ActionService) Entity() tool.SpecializedServiceInfo { return entities.DBAction }
func (s *ActionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	schemas, err := s.Domain.Schema(tool.Record{ 
		entities.RootID(entities.DBSchema.Name) : record[entities.RootID(entities.DBSchema.Name)].(int64) }, true)	
	if err != nil && len(schemas) == 0 { return record, false, false }
	if link, ok := record[entities.RootID("link")]; ok {
		schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)}, true)
		if err != nil || len(schemas) == 0 { return record, false, false }
	}
	return record, true, false
}
func (s *ActionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *ActionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ActionService) WriteRowAutomation(record tool.Record, tableName string) { }

func (s *ActionService) PostTreatment(results tool.Results, tablename string, dest_id... string) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		schemas, err := s.Domain.Schema(tool.Record{
			entities.RootID(entities.DBSchema.Name) : record[entities.RootID(entities.DBSchema.Name)]}, true)
		if err != nil || len(schemas) == 0 { continue }
		link_path := "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
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
							   "parameters" : strings.Split(fmt.Sprintf("%v", record["parameters"]), ","),
							   "link_path" : link_path }
		sqlFilter := entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
		sqlFilter += entities.DBSchema.Name + " WHERE name=" + conn.Quote(fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])) + ")"
		// retrive all fields from schema...
		params := tool.Params{ tool.RootTableParam : entities.DBSchemaField.Name, 
		                       tool.RootRowsParam: tool.ReservedParam }
		schemas, err = s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", sqlFilter)
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
		newRec["schema_name"]=fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		newRec["schema"]=schemes
		res = append(res, newRec)
	}
	return res 
}

func (s *ActionService) ConfigureFilter(tableName string) (string, string) { 
	return s.Domain.ViewDefinition(tableName)
}	
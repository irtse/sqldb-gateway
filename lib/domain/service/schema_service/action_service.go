package schema_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type ActionService struct { tool.AbstractSpecializedService }

func (s *ActionService) Entity() tool.SpecializedServiceInfo { return entities.DBAction }
func (s *ActionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	schemas, err := tool.Schema(s.Domain, record)	
	if err != nil && len(schemas) == 0 { return record, false }
	if to, ok := record["to_schema"]; ok {
		schemas, err := tool.Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	if link, ok := record["link"]; ok {
		schemas, err := tool.Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	return record, true
}
func (s *ActionService) DeleteRowAutomation(results tool.Results) { }
func (s *ActionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ActionService) WriteRowAutomation(record tool.Record) { }

func (s *ActionService) PostTreatment(results tool.Results, tablename string) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		names := []string{}
		schemas, err := tool.Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : record["from_schema"]})
		if err != nil || len(schemas) == 0 { continue }
		delete(record, "from_schema")
		record["from"]=fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		path := "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		names = append(names, schemas[0][entities.NAMEATTR].(string))
		if to, ok := record["to_schema"]; ok && to != nil {
			schemas, err := tool.Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			delete(record, "to_schema")
			record["to"]=fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
			path += "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		}
		if link, ok := record["link"]; ok  && link != nil {
			schemas, err := tool.Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			l := fmt.Sprint("%v", schemas[0][entities.NAMEATTR])
			record["link"]=l
		}
		if p, ok := record["extra_path"]; ok  && p != nil {
			delete(record, "extra_path")
			path += "/" + fmt.Sprintf("%v", p)
		}
		record["path"] = path
		record["schemas"] = map[string]tool.Record{}
        for _, tableName := range names {
			params := tool.Params{ tool.RootTableParam : tableName, }
			schemes, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
			if err == nil && len(schemes) > 0 {
				recSchemes := map[string]tool.Record{}
				for _, scheme := range schemes { recSchemes[scheme[entities.NAMEATTR].(string)]=scheme }
				for k, v := range record["schemas"].(map[string]tool.Record) { recSchemes[k]=v }
				record["schemas"]=recSchemes
			}
		}
		
		res = append(res, record)
	}
	return tool.PostTreat(s.Domain, res, tablename) 
}

func (s *ActionService) ConfigureFilter(tableName string, params tool.Params) (string, string) { 
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
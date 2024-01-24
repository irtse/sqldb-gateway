package schema_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type ActionService struct { tool.AbstractSpecializedService }

func (s *ActionService) Entity() tool.SpecializedServiceInfo { return entities.DBAction }
func (s *ActionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	schemas, err := Schema(s.Domain, record)	
	if err != nil && len(schemas) == 0 { return record, false }
	if to, ok := record["to_schema"]; ok {
		schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	if link, ok := record["link"]; ok {
		schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
		if err != nil || len(schemas) == 0 { return record, false }
	}
	return record, true
}
func (s *ActionService) DeleteRowAutomation(results tool.Results) { }
func (s *ActionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *ActionService) WriteRowAutomation(record tool.Record) { }
func (s *ActionService) PostTreatment(results tool.Results) tool.Results { 
	res := tool.Results{}
	for _, record := range results{
		names := []string{}
		schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : record["from_schema"].(int64)})
		names = append(names, schemas[0][entities.NAMEATTR].(string))
		if err != nil || len(schemas) == 0 { continue }
		if to, ok := record["to_schema"]; ok {
			schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			names = append(names, schemas[0][entities.NAMEATTR].(string))
		}
		if link, ok := record["link"]; ok {
			schemas, err := Schema(s.Domain, tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			names = append(names, schemas[0][entities.NAMEATTR].(string))
		}
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
	return res 
}

func (s *ActionService) ConfigureFilter(tableName string, params tool.Params) (string, string) { return "", "" }	
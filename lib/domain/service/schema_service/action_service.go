package schema_service

import (
	"fmt"
	"strings"
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
		newRec := tool.Record{}
		names := []string{}
		schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : record[entities.RootID("from")]})
		if err != nil || len(schemas) == 0 { continue }
		path := "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		link_path := ""
		names = append(names, schemas[0][entities.NAMEATTR].(string))
		if to, ok := record[entities.RootID("to")]; ok && to != nil {
			schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : to.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			path += "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
		}
		if link, ok := record[entities.RootID("link")]; ok  && link != nil {
			schemas, err := s.Domain.Schema(tool.Record{entities.RootID(entities.DBSchema.Name) : link.(int64)})
			if err != nil || len(schemas) == 0 { continue }
			link_path = "/" + fmt.Sprintf("%v", schemas[0][entities.NAMEATTR])
			restr, ok2 := record["link_sql_restriction"]
			order, ok3 := record["link_sql_order"]
			dir, ok4 := record["link_sql_dir"]
			cols, ok5 := record["link_sql_columns"]
			if ok2 || ok3 || ok4 || ok5 {
				link_path += "?rows=all"
				if ok2 && restr != nil && restr != "" { 
					link_path += "&" + tool.RootSQLFilterParam + "=" + strings.Replace(restr.(string), " ", "+", -1) }
				if ok3  && order != nil && order != "" { 
					link_path += "&" + tool.RootOrderParam + "=" + strings.Replace(order.(string), " ", "+", -1) 
				}
				if ok4  && dir != nil && dir != "" { 
					link_path += "&" + tool.RootDirParam + "=" + strings.Replace(dir.(string), " ", "+", -1) 
				}
				if ok5  && cols != nil && cols != "" { 
					link_path += "&" + tool.RootColumnsParam + "=" + cols.(string)
				}
				newRec["link_path"]=link_path
			}
		}
		if p, ok := record["extra_path"]; ok  && p != nil {
			path += "/" + fmt.Sprintf("%v", p)
		}
		newRec["kind"]=record["kind"]
		newRec["method"]=record["method"]
		newRec["path"] = path
		newRec["schemas"] = map[string]tool.Record{}
        for _, tableName := range names {
			params := tool.Params{ tool.RootTableParam : tableName, }
			schemes, err := s.Domain.SuperCall(params, tool.Record{}, tool.SELECT, "Get")
			if err == nil && len(schemes) > 0 {
				recSchemes := map[string]tool.Record{}
				for _, scheme := range schemes { recSchemes[scheme[entities.NAMEATTR].(string)]=scheme }
				newRec["schemas"]=recSchemes
			}
		}
		res = append(res, newRec)
	}
	return res 
}

func (s *ActionService) ConfigureFilter(tableName string, params tool.Params) (string, string) { 
	return s.Domain.ViewDefinition(tableName, params)
}	
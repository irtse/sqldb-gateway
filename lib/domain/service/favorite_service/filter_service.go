package favorite_service
import (
	"fmt"
	"strings"
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)
type FilterService struct { 
	utils.AbstractSpecializedService 
	Fields []map[string]interface{}
}

func (s *FilterService) Entity() utils.SpecializedServiceInfo { return schserv.DBFilter }
func (s *FilterService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { }
func (s *FilterService) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {
	if id, ok := record[utils.SpecialIDParam]; ok { 
		p := utils.AllParams(schserv.DBFilterField.Name)
		p[schserv.RootID(schserv.DBFilter.Name)] = fmt.Sprintf("%v", id)
		s.Domain.SuperCall(p, utils.Record{}, utils.DELETE)
	}
	for _, field := range s.Fields {
		schema, err := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if _, ok := field["name"]; !ok || err != nil { continue }
		f, err := schema.GetField(fmt.Sprintf("%v", field["name"]))
		if err != nil { continue }
		delete(field, "name")
		field[schserv.RootID(schserv.DBSchemaField.Name)] = f.ID
		field[schserv.RootID(schserv.DBFilter.Name)]=record[utils.SpecialIDParam]
		s.Domain.Call(utils.AllParams(schserv.DBFilterField.Name), field, utils.UPDATE)
	}
}
func (s *FilterService) WriteRowAutomation(record map[string]interface{}, tableName string) {
	for _, field := range s.Fields {
		schema, err := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if _, ok := field["name"]; !ok || err != nil { continue }
		f, err := schema.GetField(fmt.Sprintf("%v", field["name"]))
		if err != nil { continue }
		delete(field, "name")
		field[schserv.RootID(schserv.DBSchemaField.Name)] = f.ID
		field[schserv.RootID(schserv.DBFilter.Name)]=record[utils.SpecialIDParam]
		t, err := s.Domain.Call(utils.AllParams(schserv.DBFilterField.Name), field, utils.CREATE)
		fmt.Println(t, err)
	}
}
func (s *FilterService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results { 
	res := s.Domain.PostTreat(results, tableName, true) 
	if s.Domain.IsShallowed() { 
		for _, rec := range res {
			p := utils.AllParams(schserv.DBFilterField.Name)
			p[schserv.RootID(schserv.DBFilter.Name)] = rec.GetString(utils.SpecialIDParam)
			schema, err := schserv.GetSchemaByID(rec.GetInt("schema_id"))
			if err != nil { continue }
			fields, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT)
			if err != nil || len(fields) == 0 { continue }
			fieldsID := []string{}
			for _, field := range fields { 
				ff, err := schema.GetFieldByID(field.GetInt(schserv.RootID(schserv.DBSchemaField.Name)))
				if err != nil { continue }
				fieldsID = append(fieldsID, ff.Name) 
			}
			rec["fields"] = fieldsID
		}
	}
	return res
}
func (s *FilterService) ConfigureFilter(tableName string) (string, string, string, string) { return s.Domain.ViewDefinition(tableName) }
func (s *FilterService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, bool, bool) {
	if s.Domain.GetMethod() != utils.DELETE {
		if _, ok := record["link"]; ok { 
			schema, err := schserv.GetSchema(fmt.Sprintf("%v", record["link"]))
			delete(record, "link")
			if err != nil { return record, false, false }
			record[schserv.RootID(schserv.DBSchema.Name)] = schema.ID
		}
		schema, _ := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if n, ok := record["name"]; ok { 
			p := utils.AllParams(schserv.DBFilter.Name)
			p[schserv.RootID(schserv.DBSchema.Name)] = fmt.Sprintf("%v", record[schserv.RootID(schserv.DBSchema.Name)])
			p["name"] = fmt.Sprintf("%v", n)
			t, err := s.Domain.SuperCall(p, utils.Record{}, utils.SELECT)
			if err == nil && len(t) > 0 { record[utils.SpecialIDParam] = t[0][utils.SpecialIDParam] }
		}
		name := fmt.Sprintf("%v", record[schserv.NAMEKEY]) + " "
		if strings.Contains(name, "<nil>") { name = "" }
		if fields, ok := record["fields"]; ok { 
			s.Fields = []map[string]interface{}{}
			for _, field := range fields.([]interface{}) { s.Fields = append(s.Fields, field.(map[string]interface{})) }
		}
		if fields, ok := record["view_fields"]; ok { 
			name += "view "
			record["is_view"]=true
			s.Fields = []map[string]interface{}{}
			for _, field := range fields.([]interface{}) { s.Fields = append(s.Fields, field.(map[string]interface{})) }
		}
		if _, ok := record[schserv.DBEntity.Name]; !ok {
			users, err := s.Domain.SuperCall(utils.AllParams(schserv.DBUser.Name), utils.Record{}, utils.SELECT, "name='"+ s.Domain.GetUser() + "' OR email='" + s.Domain.GetUser() + "'")
			if err == nil && len(users) > 0 { 
				record[schserv.RootID(schserv.DBUser.Name)]=users[0][utils.SpecialIDParam]
				record[schserv.RootID(schserv.DBUser.Name)]=users[0][utils.SpecialIDParam]
				res, err := s.Domain.SuperCall(utils.AllParams(schserv.DBFilter.Name), utils.Record{}, utils.SELECT, schserv.RootID(schserv.DBUser.Name) + "=" + users[0].GetString(utils.SpecialIDParam) + " AND " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
				count := 0
				if err == nil { count = len(res) }
				name += "filter n°" + fmt.Sprintf("%v", count + 1)
			}
			
		} else {
			res, err := s.Domain.SuperCall(utils.AllParams(schserv.DBFilter.Name), utils.Record{}, utils.SELECT, schserv.RootID(schserv.DBEntity.Name) + "=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBEntity.Name)]) + " AND " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
			count := 0
			if err == nil { count = len(res) }
			name += "filter n°" + fmt.Sprintf("%v", count + 1)
		}
		record[schserv.NAMEKEY] = name
	}
	return record, true, true
}
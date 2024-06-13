package favorite_service
import (
	"fmt"
	"sort"
	"strings"
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
)
type FilterService struct { 
	utils.AbstractSpecializedService 
	Fields []map[string]interface{}
	UpdateFields bool
}

func (s *FilterService) Entity() utils.SpecializedServiceInfo { return schserv.DBFilter }
func (s *FilterService) DeleteRowAutomation(results []map[string]interface{}, tableName string) { }
func (s *FilterService) UpdateRowAutomation(results []map[string]interface{}, record map[string]interface{}) {
	s.WriteRowAutomation(record, schserv.DBFilter.Name)
}
func (s *FilterService) WriteRowAutomation(record map[string]interface{}, tableName string) {
	for _, field := range s.Fields {
		if _, ok := record[schserv.RootID(schserv.DBSchema.Name)]; !ok { continue }
		schema, err := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
		if _, ok := field["name"]; !ok || err != nil { continue }
		f, err := schema.GetField(fmt.Sprintf("%v", field["name"]))
		if err != nil { 
			delete(field, "name")
			field[schserv.RootID(schserv.DBFilter.Name)]=record[utils.SpecialIDParam]
			_, err = s.Domain.Call(utils.AllParams(schserv.DBFilterField.Name), field, utils.CREATE)
			fmt.Printf("error: %v\n", err)
			continue 
		}
		delete(field, "name")
		field[schserv.RootID(schserv.DBSchemaField.Name)] = f.ID
		field[schserv.RootID(schserv.DBFilter.Name)]=record[utils.SpecialIDParam]
		s.Domain.Call(utils.AllParams(schserv.DBFilterField.Name), field, utils.CREATE)
	}
}
func (s *FilterService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results {
	selected := map[string]bool{}
	for _, rec := range results { 
		if b, ok := rec["is_selected"]; !ok || b == nil { rec["is_selected"] = false }
		selected[rec.GetString(utils.SpecialIDParam)] = rec["is_selected"].(bool) 
	} 
	res := s.Domain.PostTreat(results, tableName, true) 
	rr := utils.Results{}
	for _, rec := range res {
		rec["is_selected"] = selected[rec.GetString(utils.SpecialIDParam)]
		schema, err := schserv.GetSchemaByID(rec.GetInt("schema_id"))
		if err != nil { rr = append(rr, rec) }
		fields, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilterField.Name + " WHERE " + schserv.RootID(schserv.DBFilter.Name) + "=" + fmt.Sprintf("%v", rec[utils.SpecialIDParam]))
		if err != nil || len(fields) == 0 { rr = append(rr, rec); continue }
		fieldsID := []schserv.FilterModel{}
		sort.SliceStable(fields, func(i, j int) bool{ return fields[i]["index"].(float64) <= fields[j]["index"].(float64) })
		for _, field := range fields { 
			separator := ""
			if sep, ok := field["separator"]; ok && sep != nil { separator = fmt.Sprintf("%v", sep) }
			ff, err := schema.GetFieldByID(utils.GetInt(field, schserv.RootID(schserv.DBSchemaField.Name)))
			if err != nil { 
				model := schserv.FilterModel{ ID: utils.GetInt(res[0], utils.SpecialIDParam),  Name: "id",
					Index: field["index"].(float64), Label: "id", Type: "integer",
					Value: fmt.Sprintf("%v", field["value"]), Separator: separator,  Operator: fmt.Sprintf("%v", field["operator"]), Dir: fmt.Sprintf("%v", field["dir"])}
				fieldsID = append(fieldsID, model) 
				continue
			}
			model := schserv.FilterModel{ ID: utils.GetInt(res[0], utils.SpecialIDParam),  Name: ff.Name, Label: ff.Label, Index: field["index"].(float64),
				Type: ff.Type, Value: fmt.Sprintf("%v", field["value"]), Separator: fmt.Sprintf("%v", field["separator"]),  Operator: fmt.Sprintf("%v", field["operator"]), Dir: fmt.Sprintf("%v", field["dir"])}
			fieldsID = append(fieldsID, model) 
		}
		if rec["elder"] == nil { 
			fils, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilter.Name + " WHERE id=" + fmt.Sprintf("%v", rec[utils.SpecialIDParam]))
			if err == nil && len(fils) > 0 { rec["elder"] = fils[0]["elder"] } else { rec["elder"] = "all" }
		}
		rec["filter_fields"] = fieldsID
		rr = append(rr, rec)
	}
	return rr
}
func (s *FilterService) ConfigureFilter(tableName string, innerestr... string) (string, string, string, string) { 
	return s.Domain.ViewDefinition(tableName, innerestr...) 
}
func (s *FilterService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	s.UpdateFields = false
	if s.Domain.GetMethod() != utils.DELETE {
		if _, ok := record["link"]; ok { 
			schema, err := schserv.GetSchema(fmt.Sprintf("%v", record["link"]))
			delete(record, "link")
			if err != nil { return record, err, false }
			record[schserv.RootID(schserv.DBSchema.Name)] = schema.ID
		}
		if n, ok := record["name"]; ok { 
			t, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilter.Name + " WHERE " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBSchema.Name)]) + " AND name='" + fmt.Sprintf("%v", n) + "'")
			if err == nil && len(t) > 0 { record[utils.SpecialIDParam] = t[0][utils.SpecialIDParam] }
		} 
		if fields, ok := record["view_fields"]; ok {
			s.UpdateFields = true 
			s.Fields = []map[string]interface{}{}
			for _, field := range fields.([]interface{}) { s.Fields = append(s.Fields, field.(map[string]interface{})) }
		}
		if fields, ok := record["filter_fields"]; ok {
			s.UpdateFields = true
			s.Fields = []map[string]interface{}{}
			for _, field := range fields.([]interface{}) { s.Fields = append(s.Fields, field.(map[string]interface{})) }
		}
		if s.Domain.GetMethod() == utils.UPDATE && s.UpdateFields {
			p := utils.AllParams(schserv.DBFilterField.Name)
			p[schserv.RootID(schserv.DBFilter.Name)] = fmt.Sprintf("%v", record[utils.SpecialIDParam])
			s.Domain.SuperCall(p, utils.Record{}, utils.DELETE)
		}
		if s.Domain.GetMethod() == utils.CREATE {
			name := fmt.Sprintf("%v", record[schserv.NAMEKEY]) + " "
			if strings.Contains(name, "<nil>") { name = "" }
			if _, ok := record["view_fields"]; ok { 
				name += "view "
				record["is_view"]=true
			}
			if _, ok := record[schserv.DBEntity.Name]; !ok && record[schserv.RootID(schserv.DBSchema.Name)] != nil {
				schema, _ := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
				users, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBUser.Name + " WHERE name='" + s.Domain.GetUser() + "' OR email='" + s.Domain.GetUser() + "'")
				if err == nil && len(users) > 0 { 
					record[schserv.RootID(schserv.DBUser.Name)]=users[0][utils.SpecialIDParam]
					record[schserv.RootID(schserv.DBUser.Name)]=users[0][utils.SpecialIDParam]
					res, err := s.Domain.SuperCall(utils.AllParams(schserv.DBFilter.Name), utils.Record{}, utils.SELECT, schserv.RootID(schserv.DBUser.Name) + "=" + utils.GetString(users[0], utils.SpecialIDParam) + " AND " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
					count := 0
					if err == nil { count = len(res) }
					if !strings.Contains(name, "filter n°") { name += "filter n°" + fmt.Sprintf("%v", count + 1) }
				}
			} else if record[schserv.RootID(schserv.DBSchema.Name)] != nil {
				schema, _ := schserv.GetSchemaByID(record[schserv.RootID(schserv.DBSchema.Name)].(int64))
				res, err := s.Domain.GetDb().QueryAssociativeArray("SELECT * FROM " + schserv.DBFilter.Name + " WHERE " + schserv.RootID(schserv.DBEntity.Name) + "=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBEntity.Name)]) + " AND " + schserv.RootID(schserv.DBSchema.Name) + "=" + fmt.Sprintf("%v", schema.ID))
				count := 0
				if err == nil { count = len(res) }
				if !strings.Contains(name, "filter n°") { name += "filter n°" + fmt.Sprintf("%v", count + 1) }
			}
			record[schserv.NAMEKEY] = name
		}
	} else {
		p := utils.AllParams(schserv.DBFilterField.Name)
		p[schserv.RootID(schserv.DBFilter.Name)] = fmt.Sprintf("%v", record[utils.SpecialIDParam])
		s.Domain.SuperCall(p, utils.Record{}, utils.DELETE)
	}
	if sel, ok := record["is_selected"]; ok && sel.(bool) {
		s.Domain.GetDb().QueryAssociativeArray("UPDATE " + schserv.DBFilter.Name + " SET is_selected=false WHERE " + schserv.RootID(schserv.DBFilter.Name) + "=" + fmt.Sprintf("%v", record[schserv.RootID(schserv.DBFilter.Name)]))
	}
	delete(record, "filter_fields")
	return record, nil, true
}
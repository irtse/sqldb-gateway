package favorite_service

import (
	"fmt"
	"sort"
	"sqldb-ws/domain/filter"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	utils "sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	"strconv"
)

// DONE - ~ 200 LINES - PARTIALLY TESTED
type FilterService struct {
	servutils.AbstractSpecializedService
	Fields       []map[string]interface{}
	UpdateFields bool
}

func (s *FilterService) Entity() utils.SpecializedServiceInfo                                    { return ds.DBFilter }
func (s *FilterService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {}
func (s *FilterService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	s.SpecializedCreateRow(record, ds.DBFilter.Name)
}
func (s *FilterService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	for _, field := range s.Fields {
		if _, ok := record[ds.SchemaDBField]; !ok {
			continue
		}
		if schema, err := schserv.GetSchemaByID(record[ds.SchemaDBField].(int64)); err == nil && field["name"] != nil {
			delete(field, "name")
			field[ds.FilterDBField] = record[utils.SpecialIDParam]
			f, err := schema.GetField(fmt.Sprintf("%v", field["name"]))
			if err == nil {
				field[ds.SchemaDBField] = f.ID
			}
			s.Domain.Call(utils.AllParams(ds.DBFilterField.Name), field, utils.CREATE)
		}
	}
}

func (s *FilterService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) (res utils.Results) {
	selected := make(map[string]bool)
	for _, rec := range results { // memorize selected filters
		id := rec.GetString(utils.SpecialIDParam)
		selected[id] = rec["is_selected"] == nil || rec["is_selected"].(bool)
	}
	for _, rec := range view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true) { // transform to generic view
		rec["is_selected"] = selected[rec.GetString(utils.SpecialIDParam)] // restore selected filters
		schema, err := schserv.GetSchemaByID(rec.GetInt("schema_id"))
		if fields, err2 := s.Domain.GetDb().SelectQueryWithRestriction( // get filter fields
			ds.DBFilterField.Name,
			map[string]interface{}{ds.FilterDBField: rec.GetInt(utils.SpecialIDParam)},
			false,
		); err == nil && err2 == nil { // sort fields by index
			sort.SliceStable(fields, func(i, j int) bool {
				return fields[i]["index"].(int64) <= fields[j]["index"].(int64)
			})
			rec["filter_fields"] = []sm.FilterModel{} // add fields to filter
			for _, field := range fields {
				if ff, err := schema.GetFieldByID(utils.GetInt(field, ds.SchemaDBField)); err == nil {
					model := sm.FilterModel{
						ID:        utils.GetInt(rec, utils.SpecialIDParam),
						Name:      ff.Name,
						Label:     ff.Label,
						Index:     float64(field["index"].(int64)),
						Type:      ff.Type,
						Value:     fmt.Sprintf("%v", field["value"]),
						Separator: fmt.Sprintf("%v", field["separator"]),
						Operator:  fmt.Sprintf("%v", field["operator"]),
						Dir:       fmt.Sprintf("%v", field["dir"]),
					}
					if width, err := strconv.ParseFloat(fmt.Sprintf("%v", field["width"]), 64); err == nil {
						model.Width = width
					}
					rec["filter_fields"] = append(rec["filter_fields"].([]sm.FilterModel), model)
				}
			}
			if rec["elder"] == nil { // get elder filter
				if fils, _ := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilter.Name,
					map[string]interface{}{"id": rec[utils.SpecialIDParam]}, false); len(fils) > 0 {
					rec["elder"] = fils[0]["elder"]
				} else {
					rec["elder"] = "all"
				}
			}
		}
		res = append(res, rec)
	}
	return
}

func (s *FilterService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, innerestr...)
}

func (s *FilterService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	s.UpdateFields = false
	method := s.Domain.GetMethod()

	if method != utils.DELETE {
		if err := s.ProcessLink(record); err != nil {
			return record, err, false
		}
		s.ProcessName(record)
		s.ProcessFields(record, "view_fields")
		s.ProcessFields(record, "filter_fields")

		if method == utils.UPDATE && s.UpdateFields {
			s.HandleUpdate(record)
		}

		if method == utils.CREATE {
			s.HandleCreate(record)
		}
	} else {
		s.HandleDelete(record)
	}

	s.ProcessSelection(record)
	return record, nil, true
}

func (s *FilterService) ProcessLink(record map[string]interface{}) error {
	if link, ok := record["link"]; ok {
		schema, err := schserv.GetSchema(fmt.Sprintf("%v", link))
		delete(record, "link")
		if err != nil {
			return err
		}
		record[ds.SchemaDBField] = schema.ID
	}
	return nil
}

func (s *FilterService) ProcessName(record map[string]interface{}) {
	if name, ok := record["name"]; ok {
		if result, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
			ds.SchemaDBField: record[ds.SchemaDBField],
			"name":           name,
		}, false); err == nil && len(result) > 0 {
			record[utils.SpecialIDParam] = result[0][utils.SpecialIDParam]
		}
	}
}

func (s *FilterService) ProcessFields(record map[string]interface{}, fieldType string) {
	if fields, ok := record[fieldType]; ok {
		s.UpdateFields = true
		s.Fields = make([]map[string]interface{}, 0)
		for _, field := range fields.([]interface{}) {
			s.Fields = append(s.Fields, field.(map[string]interface{}))
		}
	}
}

func (s *FilterService) HandleUpdate(record map[string]interface{}) {
	params := utils.AllParams(ds.DBFilterField.Name)
	params[ds.FilterDBField] = fmt.Sprintf("%v", record[utils.SpecialIDParam])
	s.Domain.DeleteSuperCall(params)
}

func (s *FilterService) HandleCreate(record map[string]interface{}) {
	name := utils.GetString(record, sm.NAMEKEY)
	if _, ok := record["view_fields"]; ok { // is a view filter
		name += "view "
		record["is_view"] = true
	}
	if schemaID := record[ds.SchemaDBField]; schemaID != nil {
		schema, _ := schserv.GetSchemaByID(schemaID.(int64))
		if _, ok := record[ds.DBEntity.Name]; !ok {
			s.HandleUserFilterNaming(record, schema, &name)
		} else {
			s.HandleEntityFilterNaming(record, schema, &name)
		}
	}
	record[sm.NAMEKEY] = name
}

func (s *FilterService) HandleUserFilterNaming(record map[string]interface{}, schema sm.SchemaModel, name *string) {
	if user, ok := servutils.GetUserRecord(s.Domain); ok {
		record[ds.UserDBField] = user[utils.SpecialIDParam]
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
			ds.UserDBField:   utils.GetString(user, utils.SpecialIDParam),
			ds.SchemaDBField: schema.ID,
		}, false); err == nil {
			*name += fmt.Sprintf("filter n°%d", len(res)+1)
		}
	}
}

func (s *FilterService) HandleEntityFilterNaming(record map[string]interface{}, schema sm.SchemaModel, name *string) {
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
		ds.EntityDBField: utils.GetString(record, ds.EntityDBField),
		ds.SchemaDBField: schema.ID,
	}, false); err == nil {
		*name += fmt.Sprintf("filter n°%d", len(res)+1)
	}
}

func (s *FilterService) HandleDelete(record map[string]interface{}) {
	params := utils.AllParams(ds.DBFilterField.Name)
	params[ds.FilterDBField] = fmt.Sprintf("%v", record[utils.SpecialIDParam])
	s.Domain.DeleteSuperCall(params)
}

func (s *FilterService) ProcessSelection(record map[string]interface{}) {
	if sel, ok := record["is_selected"]; ok && sel.(bool) { // TODO
		s.Domain.GetDb().UpdateQuery(ds.DBFilter.Name, utils.Record{
			"is_selected": false,
		}, map[string]interface{}{
			ds.FilterDBField: record[ds.FilterDBField],
		}, true)
	}
	delete(record, "filter_fields")
	delete(record, "view_fields")
}

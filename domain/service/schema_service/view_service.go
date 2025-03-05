package schema_service

import (
	"fmt"
	"runtime"
	"slices"
	"sort"
	filterserv "sqldb-ws/domain/filter"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	infrastructure "sqldb-ws/infrastructure/service"
	"strings"
)

// DONE - ~ 200 LINES - NOT TESTED
type ViewService struct {
	servutils.AbstractSpecializedService
	infrastructure.InfraSpecializedService
}

func (s *ViewService) Entity() utils.SpecializedServiceInfo { return ds.DBView }

func (s *ViewService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	return servutils.CheckAutoLoad(tablename, record, s.Domain)
}

func (s *ViewService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	restr, _, _, _ := filterserv.NewFilterService(s.Domain).GetQueryFilter(tableName, innerestr...)
	return restr, "", "", ""
}

func (s *ViewService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) (res utils.Results) {
	runtime.GOMAXPROCS(5)
	channel := make(chan utils.Record, len(results))
	for _, record := range results {
		go s.TransformToView(record, channel, dest_id...)
	}
	for range results {
		if rec := <-channel; rec != nil {
			res = append(res, rec)
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		return int64(res[i]["index"].(float64)) <= int64(res[j]["index"].(float64))
	})
	return
}

func (s *ViewService) TransformToView(record utils.Record, channel chan utils.Record, dest_id ...string) {
	if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err != nil {
		channel <- nil
	} else {
		s.Domain.HandleRecordAttributes(record)
		rec := NewViewFromRecord(schema, record)
		if userRecord, exists := servutils.GetUserRecord(s.Domain); exists {
			s.addFavorizeInfo(userRecord, record, rec)
		}
		params := utils.GetRowTargetParameters(schema.Name, s.combineDestinations(dest_id))
		params.EnrichCondition(s.Domain.GetParams(), func(k string) bool {
			_, ok := params[k]
			return !ok && k != "new" && !strings.Contains(k, "dest_table") && k != "id"
		})
		sqlFilter, view, dir := s.getFilterDetails(record)
		params.UpdateParamsWithFilters(view, dir)
		params.EnrichCondition(s.Domain.GetParams(), func(k string) bool {
			return k != utils.RootRowsParam && k != utils.SpecialIDParam && k != utils.RootTableParam
		})
		s.Domain.GetParams().Delete(func(k string) bool {
			return k == utils.RootRowsParam || k == utils.SpecialIDParam || k == utils.RootTableParam || k == utils.SpecialSubIDParam
		})
		if datas, rec, err := s.fetchData(params, sqlFilter, record, rec, schema); err != nil {
			fmt.Errorf("error while fetching data: %v", err)
			channel <- nil
		} else {
			if record["is_list"] != nil && record["is_list"].(bool) {
				filterService := filterserv.NewFilterService(s.Domain)
				SQLrestriction, _, _, _ := filterService.GetQueryFilter(schema.Name, sqlFilter)
				rec["new"], rec["max"] = filterService.CountNewDataAccess(schema.Name, []string{SQLrestriction}, params.Anonymized())
			}
			rec = *s.processData(&rec, datas, schema, record, view, params)
			if view != "" {
				rec["order"] = strings.Split(view, ",")
			}
			rec["link_path"] = "/" + utils.MAIN_PREFIX + "/" + fmt.Sprintf(ds.DBView.Name) + "?rows=" + fmt.Sprintf("%v", record[utils.SpecialIDParam])
			channel <- rec
		}
	}
}

func (s *ViewService) addFavorizeInfo(userRecord utils.Record, record utils.Record, rec utils.Record) utils.Record {
	rec["favorize_body"] = utils.Record{
		ds.UserDBField:   userRecord[utils.SpecialIDParam],
		ds.EntityDBField: record[utils.SpecialIDParam],
	}
	rec["favorize_path"] = fmt.Sprintf("/%s/%s?%s=%s",
		utils.MAIN_PREFIX, ds.DBViewAttribution.Name, utils.RootRowsParam, utils.ReservedParam)

	attributions, _ := s.Domain.GetDb().SelectQueryWithRestriction(
		ds.DBViewAttribution.Name,
		map[string]interface{}{
			ds.UserDBField: userRecord[utils.SpecialIDParam],
			ds.ViewDBField: record[utils.SpecialIDParam],
		}, false)
	rec["is_favorize"] = len(attributions) > 0
	return rec
}

func (s *ViewService) combineDestinations(dest_id []string) string {
	return strings.Join(dest_id, ",")
}

func (s *ViewService) getFilterDetails(record utils.Record) (string, string, string) {
	filter := utils.GetString(record, ds.FilterDBField)
	viewFilter := utils.GetString(record, ds.ViewFilterDBField)
	sqlFilter, view, _, dir, _ := filterserv.NewFilterService(s.Domain).GetFilterForQuery(filter, viewFilter, utils.GetString(record, ds.SchemaDBField))
	return sqlFilter, view, dir
}

func (s *ViewService) fetchData(params utils.Params, sqlFilter string, record utils.Record, rec utils.Record, schema sm.SchemaModel) (utils.Results, utils.Record, error) {
	datas := utils.Results{}
	if !s.Domain.GetEmpty() {
		d, _ := s.Domain.SuperCall(params, utils.Record{}, utils.SELECT, true, sqlFilter)
		if record["is_list"] != nil && record["is_list"].(bool) {
			filterService := filterserv.NewFilterService(s.Domain)
			SQLrestriction, _, _, _ := filterService.GetQueryFilter(schema.Name, sqlFilter)
			rec["new"], rec["max"] = filterService.CountNewDataAccess(schema.Name, []string{SQLrestriction}, params.Anonymized())
		}
		for _, data := range d {
			if s.Domain.GetParams()["new"] != "enable" || slices.Contains(record["new"].([]string), data.GetString("id")) {
				datas = append(datas, data)
			}
		}
	}
	return datas, rec, nil
}

func (s *ViewService) processData(rec *utils.Record, datas utils.Results, schema sm.SchemaModel,
	record utils.Record, view string, params utils.Params) *utils.Record {
	if !s.Domain.IsShallowed() {
		treated := view_convertor.NewViewConvertor(s.Domain).TransformToView(datas, schema.Name, false)
		if len(treated) > 0 {
			for k, v := range treated[0] {
				if v != nil {
					switch k {
					case "items":
						(*rec)[k] = s.extractItems(v.([]interface{}), record, schema)
					case "schema":
						(*rec)[k] = s.extractSchema(v.(map[string]interface{}), record, schema, params, view)
					case "shortcuts":
						(*rec)[k] = s.extractShortcuts(v.(map[string]interface{}), record)
					}
				} else if recValue, exists := (*rec)[k]; !exists || recValue == "" {
					(*rec)[k] = v
				}
			}
		}
	}
	return rec
}

func (s *ViewService) extractSchema(value map[string]interface{}, record utils.Record, schema sm.SchemaModel, params utils.Params, view string) map[string]interface{} {
	newV := map[string]interface{}{}
	for fieldName, field := range value {
		if fieldName != ds.WorkflowDBField && schema.Name == ds.DBRequest.Name && record["is_empty"].(bool) {
			continue
		}
		col, ok := params[utils.RootColumnsParam]
		field.(map[string]interface{})["active"] = !ok || col == "" || strings.Contains(view, fieldName)
		newV[fieldName] = field
	}
	return newV
}

func (s *ViewService) extractItems(value []interface{}, record utils.Record, schema sm.SchemaModel) []interface{} {
	for _, item := range value {
		values := item.(map[string]interface{})["values"]
		if list, ok := record["is_list"]; ok && list.(bool) {
			path := utils.RootRowsParam
			if strings.Contains(path, ds.DBView.Name) {
				path = utils.RootDestTableIDParam
			}
			item.(map[string]interface{})["link_path"] = fmt.Sprintf("/%s/%s?%s=%v", utils.MAIN_PREFIX, schema.Name, utils.RootRowsParam, values.(map[string]interface{})[utils.SpecialIDParam])
			item.(map[string]interface{})["data_path"] = ""
		}
	}
	return value
}

func (s *ViewService) extractShortcuts(value map[string]interface{}, record utils.Record) map[string]interface{} {
	shorts := map[string]interface{}{}
	for shortcut, ss := range value {
		if !strings.Contains(shortcut, record.GetString(sm.NAMEKEY)) {
			shorts[shortcut] = ss
		}
	}
	return shorts
}

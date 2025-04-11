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
	"sqldb-ws/domain/task"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	"strings"
)

// DONE - ~ 200 LINES - NOT TESTED
type ViewService struct {
	servutils.AbstractSpecializedService
}

func (s *ViewService) Entity() utils.SpecializedServiceInfo { return ds.DBView }

func (s *ViewService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
	} else {
		return record, err, false
	}
}

func (s *ViewService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	restr, _, _, _ := filterserv.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
	return restr, "", "", ""
}

func (s *ViewService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) (res utils.Results) {
	runtime.GOMAXPROCS(5)
	channel := make(chan utils.Record, len(results))
	params := s.Domain.GetParams().Copy()
	for _, record := range results {
		go s.TransformToView(record, params, channel, dest_id...)
	}
	for range results {
		if rec := <-channel; rec != nil {
			res = append(res, rec)
		}
	}
	sort.SliceStable(res, func(i, j int) bool {
		return utils.ToInt64(res[i]["index"]) <= utils.ToInt64(res[j]["index"])
	})
	return
}

func (s *ViewService) TransformToView(record utils.Record, domainParams utils.Params,
	channel chan utils.Record, dest_id ...string) {
	s.Domain.SetOwn(record.GetBool("own_view"))
	if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err != nil {
		channel <- nil
	} else {
		s.Domain.HandleRecordAttributes(record)
		rec := NewViewFromRecord(schema, record)
		s.addFavorizeInfo(record, rec)
		params := utils.GetRowTargetParameters(schema.Name, s.combineDestinations(dest_id))
		params = params.EnrichCondition(domainParams.Values, func(k string) bool {
			_, ok := params.Values[k]
			return !ok && k != "new" && !strings.Contains(k, "dest_table") && k != "id"
		})
		sqlFilter, view, dir := s.getFilterDetails(record)
		params.UpdateParamsWithFilters(view, dir)
		params.EnrichCondition(domainParams.Values, func(k string) bool {
			return k != utils.RootRowsParam && k != utils.SpecialIDParam && k != utils.RootTableParam
		})
		domainParams.Delete(func(k string) bool {
			return k == utils.RootRowsParam || k == utils.SpecialIDParam || k == utils.RootTableParam || k == utils.SpecialSubIDParam
		})
		if _, ok := record["foldered"]; ok {
			if field, err := schema.GetFieldByID(record.GetInt("foldered")); err == nil {
				params.Set(utils.Rootfoldered, field.Name)
			}
		}
		if f, ok := domainParams.Get(utils.Rootfoldered); ok {
			params.Set(utils.Rootfoldered, f)
			rec["foldered"] = f
		}
		datas, rec := s.fetchData(params, domainParams, sqlFilter, record, rec, schema)
		record, rec = s.processData(rec, datas, schema, record, view, params)
		rec["link_path"] = "/" + utils.MAIN_PREFIX + "/" + fmt.Sprintf(ds.DBView.Name) + "?rows=" + utils.ToString(record[utils.SpecialIDParam])
		if _, ok := record["foldered"]; ok { // express by each column we are foldered TODO : if not in view add it
			field, err := schema.GetFieldByID(record.GetInt("foldered"))
			if err == nil {
				rec["foldered"] = field.Name
			}
		}
		if f, ok := domainParams.Get(utils.Rootfoldered); ok {
			rec["foldered"] = f
		}
		channel <- rec
	}
}

func (s *ViewService) getOrder(rec utils.Record, record utils.Record, values map[string]interface{}, view string) (utils.Record, utils.Record, string, map[string]interface{}) {
	if len(task.GetViewTask(utils.GetString(record, ds.SchemaDBField), utils.ToString(record[utils.SpecialIDParam]), s.Domain.GetUserID())) > 0 {
		vs := task.GetViewTask(utils.GetString(record, ds.SchemaDBField), utils.ToString(record[utils.SpecialIDParam]), s.Domain.GetUserID())
		if utils.GetBool(record, "is_list") {
			for _, fname := range vs {
				if val, ok := utils.ToMap(rec["schema"])[fname]; ok {
					utils.ToMap(val)["readonly"] = true
					utils.ToMap(val)["hidden"] = true
				}
				values[fname] = nil
			}
		} else {
			view = strings.Join(vs, ",")
		}
	}
	return rec, record, view, values
}

func (s *ViewService) getFilter(rec utils.Record, record utils.Record, values map[string]interface{}, schema sm.SchemaModel) (utils.Record, utils.Record, map[string]interface{}) {
	if record[ds.FilterDBField] != nil {
		if fields, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
			ds.FilterDBField: record[ds.FilterDBField],
		}, false); err == nil && len(fields) > 0 {
			for _, f := range fields {
				ff, err := schema.GetFieldByID(utils.GetInt(f, ds.SchemaFieldDBField))
				if err != nil {
					continue
				}
				if val, ok := utils.ToMap(rec["schema"])[ff.Name]; ok {
					utils.ToMap(val)["readonly"] = true
					values[ff.Name] = f["value"]
				}
			}
		}
	}
	return rec, record, values
}

func (s *ViewService) addFavorizeInfo(record utils.Record, rec utils.Record) utils.Record {
	rec["favorize_body"] = utils.Record{
		ds.ViewDBField: record.GetInt(utils.SpecialIDParam),
		ds.UserDBField: s.Domain.GetUserID(),
	}
	rec["favorize_path"] = fmt.Sprintf("/%s/%s?%s=%s",
		utils.MAIN_PREFIX, ds.DBViewAttribution.Name, utils.RootRowsParam, utils.ReservedParam)

	attributions, _ := s.Domain.GetDb().SelectQueryWithRestriction(
		ds.DBViewAttribution.Name,
		map[string]interface{}{
			ds.UserDBField: s.Domain.GetUserID(),
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
	sqlFilter, view, _, dir, _ := filterserv.NewFilterService(s.Domain).GetFilterForQuery(
		filter, viewFilter, utils.GetString(record, ds.SchemaDBField), s.Domain.GetParams())
	return sqlFilter, view, dir
}

func (s *ViewService) fetchData(params utils.Params, domainParams utils.Params, sqlFilter string, record utils.Record, rec utils.Record, schema sm.SchemaModel) (utils.Results, utils.Record) {
	datas := utils.Results{}
	if !s.Domain.GetEmpty() {
		d, _ := s.Domain.Call(params.RootRaw(), utils.Record{}, utils.SELECT, sqlFilter)
		if utils.Compare(record["is_list"], true) {
			params.SimpleDelete("limit")
			params.SimpleDelete("offset")
			filterService := filterserv.NewFilterService(s.Domain)
			restriction, _, _, _ := filterService.GetQueryFilter(schema.Name, params, sqlFilter)
			rec["new"], rec["max"] = filterService.CountNewDataAccess(schema.Name, []interface{}{restriction})
			s.Domain.GetDb().ClearQueryFilter()
		}
		for _, data := range d {
			if p, _ := domainParams.Get("new"); p != "enable" || slices.Contains(record["new"].([]string), data.GetString("id")) {
				datas = append(datas, data)
			}
		}
	}
	return datas, rec
}

func (s *ViewService) processData(rec utils.Record, datas utils.Results, schema sm.SchemaModel,
	record utils.Record, view string, params utils.Params) (utils.Record, utils.Record) {
	if utils.Compare(record["is_empty"], true) {
		datas = append(datas, utils.Record{})
	}
	if !s.Domain.IsShallowed() {
		treated := view_convertor.NewViewConvertor(s.Domain).TransformToView(datas, schema.Name, false, params)

		if len(treated) > 0 {
			rec["schema"] = s.extractSchema(utils.ToMap(treated[0]["schema"]), record, schema, params, view)
			for k, v := range treated[0] {
				if v != nil {
					switch k {
					case "items":
						rec, view = s.extractItems(utils.ToList(v), k, rec, record, schema, view)
					case "shortcuts":
						rec[k] = s.extractShortcuts(utils.ToMap(v), record)
					default:
						if recValue, exists := rec[k]; !exists || recValue == "" {
							rec[k] = v
						}
					}
				}
			}
		}
	}
	return record, rec
}

func (s *ViewService) extractSchema(value map[string]interface{}, record utils.Record, schema sm.SchemaModel, params utils.Params, view string) map[string]interface{} {
	newV := map[string]interface{}{}
	for fieldName, field := range value {
		if fieldName != ds.WorkflowDBField && schema.Name == ds.DBRequest.Name && utils.Compare(record["is_empty"], true) {
			continue
		}
		col, ok := params.Get(utils.RootColumnsParam)
		utils.ToMap(field)["active"] = !ok || col == "" || strings.Contains(view, fieldName) || strings.Contains(col, fieldName)
		newV[fieldName] = field
	}
	return newV
}

func (s *ViewService) extractItems(value []interface{}, key string, rec utils.Record, record utils.Record, schema sm.SchemaModel, view string) (utils.Record, string) {
	for _, item := range value {
		values := utils.ToMap(item)["values"]
		if utils.Compare(rec["is_list"], true) {
			path := utils.RootRowsParam
			if strings.Contains(path, ds.DBView.Name) {
				path = utils.RootDestTableIDParam
			}
			utils.ToMap(item)["link_path"] = fmt.Sprintf("/%s/%s?%s=%v", utils.MAIN_PREFIX, schema.Name,
				utils.RootRowsParam, utils.ToMap(values)[utils.SpecialIDParam])
		}
		rec, record, values = s.getFilter(rec, record, utils.ToMap(values), schema)
		rec, record, view, values = s.getOrder(rec, record, utils.ToMap(values), view)
		// here its to format by filter depending on task running about this document of viewing, if enable.
	}
	rec[key] = value
	return rec, view
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

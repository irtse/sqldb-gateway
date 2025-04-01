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
	"strings"
)

// DONE - ~ 200 LINES - NOT TESTED
type ViewService struct {
	servutils.AbstractSpecializedService
}

func (s *ViewService) Entity() utils.SpecializedServiceInfo { return ds.DBView }

func (s *ViewService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	return servutils.CheckAutoLoad(tablename, record, s.Domain)
}

func (s *ViewService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	restr, _, _, _ := filterserv.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
	return restr, "", "", ""
}

func (s *ViewService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) (res utils.Results) {
	runtime.GOMAXPROCS(5)
	userRecord, _ := servutils.GetUserRecord(s.Domain)
	channel := make(chan utils.Record, len(results))
	params := s.Domain.GetParams().Copy()
	for _, record := range results {
		go s.TransformToView(record, userRecord, params, channel, dest_id...)
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

func (s *ViewService) TransformToView(record utils.Record, userRecord utils.Record, domainParams utils.Params,
	channel chan utils.Record, dest_id ...string) {
	if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err != nil {
		channel <- nil
	} else {
		s.Domain.HandleRecordAttributes(record)
		rec := NewViewFromRecord(schema, record)
		if userRecord != nil {
			s.addFavorizeInfo(userRecord, record, rec)
		}
		params := utils.GetRowTargetParameters(schema.Name, s.combineDestinations(dest_id))
		params.EnrichCondition(domainParams, func(k string) bool {
			_, ok := params[k]
			return !ok && k != "new" && !strings.Contains(k, "dest_table") && k != "id"
		})
		sqlFilter, view, dir := s.getFilterDetails(record)
		params.UpdateParamsWithFilters(view, dir)
		params.EnrichCondition(domainParams, func(k string) bool {
			return k != utils.RootRowsParam && k != utils.SpecialIDParam && k != utils.RootTableParam
		})
		domainParams.Delete(func(k string) bool {
			return k == utils.RootRowsParam || k == utils.SpecialIDParam || k == utils.RootTableParam || k == utils.SpecialSubIDParam
		})
		datas, rec := s.fetchData(params, domainParams, sqlFilter, record, rec, schema)
		if utils.Compare(record["is_list"], true) {
			filterService := filterserv.NewFilterService(s.Domain)
			delete(params, "limit")
			delete(params, "offset")
			s.Domain.GetDb().ClearQueryFilter()
			SQLrestriction, _, _, _ := filterService.GetQueryFilter(schema.Name, params, sqlFilter)
			rec["new"], rec["max"] = filterService.CountNewDataAccess(schema.Name, []string{SQLrestriction})
		}
		rec = *s.processData(&rec, datas, schema, record, view, params)
		if view != "" {
			rec["order"] = strings.Split(view, ",")
		}
		rec["link_path"] = "/" + utils.MAIN_PREFIX + "/" + fmt.Sprintf(ds.DBView.Name) + "?rows=" + utils.ToString(record[utils.SpecialIDParam])
		channel <- rec
	}
}

func (s *ViewService) addFavorizeInfo(userRecord utils.Record, record utils.Record, rec utils.Record) utils.Record {
	fmt.Println(s.Domain.GetParams(), record, rec)
	rec["favorize_body"] = utils.Record{
		ds.ViewDBField: record.GetInt(utils.SpecialIDParam),
		ds.UserDBField: userRecord[utils.SpecialIDParam],
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
	sqlFilter, view, _, dir, _ := filterserv.NewFilterService(s.Domain).GetFilterForQuery(
		filter, viewFilter, utils.GetString(record, ds.SchemaDBField), s.Domain.GetParams())
	return sqlFilter, view, dir
}

func (s *ViewService) fetchData(params utils.Params, domainParams utils.Params, sqlFilter string, record utils.Record, rec utils.Record, schema sm.SchemaModel) (utils.Results, utils.Record) {
	datas := utils.Results{}
	if !s.Domain.GetEmpty() {
		d, _ := s.Domain.SuperCall(params, utils.Record{}, utils.SELECT, true, sqlFilter)
		if utils.Compare(record["is_list"], true) {
			delete(params, "limit")
			delete(params, "offset")
			filterService := filterserv.NewFilterService(s.Domain)
			s.Domain.GetDb().ClearQueryFilter()
			rec["new"], rec["max"] = filterService.CountNewDataAccess(schema.Name, []string{sqlFilter})
		}
		for _, data := range d {
			if domainParams["new"] != "enable" || slices.Contains(record["new"].([]string), data.GetString("id")) {
				datas = append(datas, data)
			}
		}
	}
	return datas, rec
}

func (s *ViewService) processData(rec *utils.Record, datas utils.Results, schema sm.SchemaModel,
	record utils.Record, view string, params utils.Params) *utils.Record {
	if utils.Compare(record["is_empty"], true) {
		datas = append(datas, utils.Record{})
	}
	if !s.Domain.IsShallowed() {
		treated := view_convertor.NewViewConvertor(s.Domain).TransformToView(datas, schema.Name, false)
		if len(treated) > 0 {
			for k, v := range treated[0] {
				if v != nil {
					switch k {
					case "items":
						(*rec)[k] = s.extractItems(utils.ToList(v), record, schema)
					case "schema":
						(*rec)[k] = s.extractSchema(utils.ToMap(v), record, schema, params, view)
					case "shortcuts":
						(*rec)[k] = s.extractShortcuts(utils.ToMap(v), record)
					default:
						if recValue, exists := (*rec)[k]; !exists || recValue == "" {
							(*rec)[k] = v
						}
					}
				}
			}
		}
	}
	return rec
}

func (s *ViewService) extractSchema(value map[string]interface{}, record utils.Record, schema sm.SchemaModel, params utils.Params, view string) map[string]interface{} {
	newV := map[string]interface{}{}
	for fieldName, field := range value {
		if fieldName != ds.WorkflowDBField && schema.Name == ds.DBRequest.Name && utils.Compare(record["is_empty"], true) {
			continue
		}
		col, ok := params[utils.RootColumnsParam]
		utils.ToMap(field)["active"] = !ok || col == "" || strings.Contains(view, fieldName)
		newV[fieldName] = field
	}
	return newV
}

func (s *ViewService) extractItems(value []interface{}, record utils.Record, schema sm.SchemaModel) []interface{} {
	for _, item := range value {
		values := utils.ToMap(item)["values"]
		if utils.Compare(record["is_list"], true) {
			path := utils.RootRowsParam
			if strings.Contains(path, ds.DBView.Name) {
				path = utils.RootDestTableIDParam
			}
			if !(len(schema.Fields) == 1 && schema.Fields[0].Name == "name") {
				fmt.Println("ERROR: schema fields not equal to 1 or name", schema.Name)
				utils.ToMap(item)["link_path"] = fmt.Sprintf("/%s/%s?%s=%v", utils.MAIN_PREFIX, schema.Name,
					utils.RootRowsParam, utils.ToMap(values)[utils.SpecialIDParam])
			}

			utils.ToMap(item)["data_path"] = ""
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

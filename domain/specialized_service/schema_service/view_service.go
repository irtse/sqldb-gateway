package schema_service

import (
	"fmt"
	"runtime"
	"sort"
	filterserv "sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/domain_service/view_convertor"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/schema/models"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
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
	if !s.Domain.IsSuperAdmin() {
		innerestr = append(innerestr, "only_super_admin=false")
	}
	restr, _, _, _ := filterserv.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
	return restr, "", "", ""
}

func (s *ViewService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) (res utils.Results) {
	runtime.GOMAXPROCS(5)
	channel := make(chan utils.Record, len(results))
	params := s.Domain.GetParams().Copy()
	schemas := []models.SchemaModel{}
	if len(results) == 1 {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBViewSchema.Name, map[string]interface{}{
			ds.ViewDBField: results[0][utils.SpecialIDParam],
		}, false); err == nil {
			for _, r := range res {
				if sch, err := schserv.GetSchemaByID(utils.GetInt(r, ds.SchemaDBField)); err == nil {
					schemas = append(schemas, sch)
				}
			}
		}
	}
	for _, record := range results {
		go s.TransformToView(record, schemas, params, channel, dest_id...)
	}
	for range results {
		if rec := <-channel; rec != nil {
			res = append(res, rec)
		}
	}
	for _, r := range res {
		if utils.GetBool(r, "is_empty") {
			continue
		}

		filterService := filterserv.NewFilterService(s.Domain)
		f := []string{}
		if strings.Trim(utils.GetString(r, "sqlfilter"), " ") != "" {
			f = append(f, utils.GetString(r, "sqlfilter"))
		}
		news, maxs := filterService.CountNewDataAccess(utils.GetString(r, "schema_name"), f)
		for _, scheme := range schemas {
			news1, maxs1 := filterService.CountNewDataAccess(scheme.Name, f)
			news += news1
			maxs += maxs1
		}
		r["new"] = news
		r["max"] = maxs
		delete(r, "sqlfilter")
	}

	sort.SliceStable(res, func(i, j int) bool {
		return utils.ToInt64(res[i]["index"]) <= utils.ToInt64(res[j]["index"])
	})
	return
}

func (s *ViewService) TransformToView(record utils.Record, schemas []models.SchemaModel, domainParams utils.Params,
	channel chan utils.Record, dest_id ...string) {
	s.Domain.SetOwn(record.GetBool("own_view"))
	if schema, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err != nil {
		channel <- nil
	} else {
		// retrive additionnal view to combine to the main... add a type can be filtered by a filter line
		// add type onto order and schema plus verify if filter not implied.
		// may regenerate to get limits... for file... for type and for dest_table_id if needed.
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
		if _, ok := record["group_by"]; ok {
			if field, err := schema.GetFieldByID(record.GetInt("group_by")); err == nil {
				params.Set(utils.RootGroupBy, field.Name)
			}
		}
		if f, ok := domainParams.Get(utils.RootGroupBy); ok {
			params.Set(utils.RootGroupBy, f)
			rec["group_by"] = f
		}
		datas := utils.Results{}
		rec["sqlfilter"] = sqlFilter
		if shal, ok := s.Domain.GetParams().Get(utils.RootShallow); !ok || shal != "enable" {
			datas, rec = s.fetchData(params, sqlFilter, rec)
		}
		record, rec = s.processData(rec, datas, schema, record, view, params)
		if !s.Domain.IsShallowed() {
			for _, scheme := range schemas {
				if line, ok := params.Get(utils.RootFilterLine); ok {
					if val, operator := connector.GetFieldInInjection(line, "type"); val != "" {
						if strings.Contains(operator, "LIKE") {
							if strings.Contains(operator, "NOT") && strings.Contains(scheme.Label, val) {
								continue
							} else if !strings.Contains(scheme.Label, val) {
								continue
							}
						} else if scheme.Label != val {
							continue
						}
					}
				}
				treated := view_convertor.NewViewConvertor(s.Domain).TransformToView(datas, scheme.Name, false, params)
				if len(treated) > 0 {
					items := treated[0]["items"].([]interface{})
					rec, view = s.extractItems(utils.ToList(items), "items", rec, record,
						schema, view, params, true)
					newSchema := map[string]interface{}{}
					for k, v := range rec["schema"].(map[string]interface{}) {
						if scheme.HasField(k) {
							newSchema[k] = v
						}
					}
					if rec["order"] != nil {
						order := []string{"type"}
						for _, o := range rec["order"].([]string) {
							if newSchema[o] != nil {
								order = append(order, o)
							}
						}
						rec["order"] = order
					}
					newSchema["type"] = models.ViewFieldModel{
						Label:    "type",
						Type:     "varchar",
						Index:    2,
						Readonly: true,
						Active:   true,
					}
					rec["schema"] = newSchema
				}
			}

		}
		rec["link_path"] = "/" + utils.MAIN_PREFIX + "/" + fmt.Sprintf(ds.DBView.Name) + "?rows=" + utils.ToString(record[utils.SpecialIDParam])
		if _, ok := record["group_by"]; ok { // express by each column we are foldered TODO : if not in view add it
			field, err := schema.GetFieldByID(record.GetInt("group_by"))
			if err == nil {
				rec["group_by"] = field.Name
			}
		}
		if f, ok := domainParams.Get(utils.RootGroupBy); ok {
			rec["group_by"] = f
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

// this filter a view only with its property
func (s *ViewService) getFilter(rec utils.Record, record utils.Record, values map[string]interface{}, schema sm.SchemaModel) (utils.Record, utils.Record, map[string]interface{}) {
	if record[ds.FilterDBField] != nil && s.Domain.GetEmpty() {
		if fields, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
			ds.FilterDBField + "_1": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
				"is_view":              false,
				"dashboard_restricted": false,
			}, false, utils.SpecialIDParam),
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

	attributions, _ := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(
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
	if sch, err := schserv.GetSchemaByID(utils.GetInt(record, ds.SchemaDBField)); err == nil {
		sqlFilter, view, _, dir, _ := filterserv.NewFilterService(s.Domain).GetFilterForQuery(
			filter, viewFilter, sch, s.Domain.GetParams())
		return sqlFilter, view, dir
	}
	return "", "", ""
}
func (s *ViewService) fetchData(params utils.Params, sqlFilter string, rec utils.Record) (utils.Results, utils.Record) {
	datas := utils.Results{}
	if !s.Domain.GetEmpty() {
		datas, _ = s.Domain.Call(params.RootRaw(), utils.Record{}, utils.SELECT, []interface{}{sqlFilter}...)
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
						rec, view = s.extractItems(utils.ToList(v), k, rec, record, schema,
							view, params, false)
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

func (s *ViewService) extractItems(value []interface{}, key string, rec utils.Record, record utils.Record,
	schema sm.SchemaModel, view string, params utils.Params, main bool) (utils.Record, string) {
	for _, item := range value {
		values := utils.ToMap(item)["values"]
		if len(s.Domain.DetectFileToSearchIn()) > 0 {
			for search, field := range s.Domain.DetectFileToSearchIn() {
				if utils.ToMap(values)[field] == nil || !utils.SearchInFile(utils.GetString(utils.ToMap(values), field), search) {
					continue
				}
			}
		}
		if line, ok := params.Get(utils.RootFilterLine); ok {
			if val, operator := connector.GetFieldInInjection(line, ds.DestTableDBField); val != "" && utils.GetString(utils.ToMap(values), ds.DestTableDBField) != "" {
				if schemaDest, err := schserv.GetSchemaByID(utils.GetInt(utils.ToMap(values), ds.SchemaDBField)); err == nil {
					cmd := "name" + operator + val
					if strings.Contains(operator, "LIKE") {
						cmd = "name::text " + operator + val
					}
					if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schemaDest.Name, []string{
						"id = " + utils.GetString(utils.ToMap(values), ds.DestTableDBField), cmd,
					}, false); err == nil && len(res) == 0 {
						continue
					}
				}
			}
		}
		if !main {
			utils.ToMap(values)["type"] = schema.Label
		}
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
	if rec[key] == nil {
		rec[key] = value
	} else {
		rec[key] = append(rec[key].([]interface{}), value...)
	}
	return rec, view
}

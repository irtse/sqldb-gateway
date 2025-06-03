package view_convertor

import (
	"fmt"
	"net/url"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/schema"
	scheme "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"strconv"
	"strings"
)

type ViewConvertor struct {
	Domain     utils.DomainITF
	SchemaSeen map[string]map[string]interface{}
}

func NewViewConvertor(domain utils.DomainITF) *ViewConvertor {
	return &ViewConvertor{Domain: domain, SchemaSeen: map[string]map[string]interface{}{}}
}

func (v *ViewConvertor) TransformToView(results utils.Results, tableName string, isWorkflow bool, params utils.Params) utils.Results {
	schema, err := scheme.GetSchema(tableName)
	if err != nil {
		if results == nil {
			return utils.Results{}
		}
		return results
	}
	if ids, ok := params.Get(utils.SpecialIDParam); ok || v.Domain.GetMethod() != utils.SELECT {
		if len(ids) == 0 {
			for _, r := range results {
				ids += r.GetString(utils.SpecialIDParam) + ","
			}
			ids = connector.RemoveLastChar(ids)
		}
		v.NewDataAccess(schema.GetID(), strings.Split(ids, ","), v.Domain.GetMethod()) // FOUND IT !
	}
	if v.Domain.IsShallowed() {
		return v.transformShallowedView(results, tableName, isWorkflow)
	}
	return v.transformFullView(results, schema, tableName, isWorkflow, params)
}

func (v *ViewConvertor) transformFullView(results utils.Results, schema sm.SchemaModel, tableName string, isWorkflow bool, params utils.Params) utils.Results {
	schemes, id, order, cols, addAction, _ := v.GetViewFields(tableName, false, results)
	commentBody := map[string]interface{}{}
	if len(results) == 1 {
		commentBody = map[string]interface{}{
			ds.UserDBField:      utils.ToInt64(v.Domain.GetUserID()),
			ds.SchemaDBField:    utils.ToInt64(schema.ID),
			ds.DestTableDBField: utils.GetInt(results[0], utils.SpecialIDParam),
		}
	}
	otherOrder := []string{}
	if v.Domain.GetEmpty() {
		if res, err := v.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
			ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
				ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
					"is_view":              true,
					"dashboard_restricted": false,
				}, false, utils.SpecialIDParam),
				ds.FilterDBField + "_1": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
					ds.SchemaDBField: schema.ID,
				}, false, "view_"+ds.FilterDBField),
			}, false, ds.FilterDBField),
		}, false); err == nil {
			for _, r := range res {
				if f, err := schema.GetFieldByID(utils.GetInt(r, ds.SchemaFieldDBField)); err == nil {
					otherOrder = append(otherOrder, f.Name)
				}
			}
		}
	} else if res, err := v.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
		ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
			"is_view":              true,
			"dashboard_restricted": false,
		}, false, utils.SpecialIDParam),
		ds.FilterDBField + "_1": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
			utils.SpecialIDParam: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.UserDBField: v.Domain.GetUserID(),
				ds.EntityDBField: v.Domain.GetDb().BuildSelectQueryWithRestriction(
					ds.DBEntityUser.Name,
					map[string]interface{}{
						ds.UserDBField: v.Domain.GetUserID(),
					}, true, ds.EntityDBField),
			}, true, ds.WorkflowSchemaDBField),
		}, false, "view_"+ds.FilterDBField),
		ds.FilterDBField + "_2": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
			utils.SpecialIDParam: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.SchemaDBField: schema.ID,
				"is_close":       false,
			}, false, ds.WorkflowSchemaDBField),
		}, false, "view_"+ds.FilterDBField),
	}, false); err == nil {
		for _, r := range res {
			if f, err := schema.GetFieldByID(utils.GetInt(r, ds.SchemaFieldDBField)); err == nil {
				otherOrder = append(otherOrder, f.Name)
			}
		}
	}
	o := []string{}
	for _, or := range order {
		if len(otherOrder) == 0 || slices.Contains(otherOrder, or) {
			o = append(o, or)
		}
	}
	view := sm.ViewModel{
		ID:          id,
		Name:        schema.Name,
		Label:       schema.Label,
		Description: fmt.Sprintf("%s data", tableName),
		Schema:      schemes,
		IsWrapper:   tableName == ds.DBTask.Name || tableName == ds.DBRequest.Name,
		SchemaID:    id,
		SchemaName:  tableName,
		ActionPath:  utils.BuildPath(tableName, utils.ReservedParam),
		Order:       o,
		Actions:     addAction,
		CommentBody: commentBody,
		Items:       []sm.ViewItemModel{},
		Shortcuts:   v.GetShortcuts(schema.ID, addAction),
		Redirection: v.getRedirection(),
		Triggers:    []sm.ManualTriggerModel{},
		Consents:    v.getConsent(schema.ID, results),
	}
	v.ProcessResultsConcurrently(results, tableName, cols, isWorkflow, &view, params)
	// if there is only one item in the view, we can set the view readonly to the item readonly
	if len(view.Items) == 1 {
		view.Readonly = view.Items[0].Readonly
	}
	idParamsOk := len(v.Domain.GetParams().GetAsArgs(utils.SpecialSubIDParam)) > 0
	if idParamsOk && slices.Contains(ds.PUPERMISSIONEXCEPTION, schema.Name) {
		view.Readonly = true
		for _, sch := range schemes {
			utils.ToMap(sch)["active"] = true
		}
	}
	if view.Readonly { // if the view is readonly, we remove the actions
		view.Actions = []string{"get"}
	}
	sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Sort < view.Items[j].Sort })
	return utils.Results{view.ToRecord()}
}

func (v *ViewConvertor) transformShallowedView(results utils.Results, tableName string, isWorkflow bool) utils.Results {
	res := utils.Results{}
	max := int64(0)
	if sch, err := scheme.GetSchema(tableName); err == nil {
		max, _ = filter.NewFilterService(v.Domain).CountMaxDataAccess(sch.Name, []string{})
	}
	for _, record := range results {
		if _, ok := record["is_draft"]; ok && record.GetBool("is_draft") && !v.Domain.IsOwn(false, false, utils.SELECT) {
			continue
		}
		if record.GetString(sm.NAMEKEY) == "" {
			res = append(res, record)
			continue
		}
		res = append(res, v.createShallowedViewItem(record, tableName, isWorkflow, max))
	}
	return res
}

func (v *ViewConvertor) ProcessResultsConcurrently(results utils.Results, tableName string,
	cols map[string]sm.FieldModel, isWorkflow bool, view *sm.ViewModel, params utils.Params) {
	const maxConcurrent = 5
	runtime.GOMAXPROCS(maxConcurrent)
	channel := make(chan sm.ViewItemModel, len(results))
	defer close(channel)
	go func() {
		if err := recover(); err != nil {
			fmt.Printf("panic occurred: %v\n%v\n", err, string(debug.Stack()))
		}
	}()
	createdIds := []string{}
	sch, err := scheme.GetSchema(tableName)
	if err == nil {
		createdIds = filter.NewFilterService(v.Domain).GetCreatedAccessData(sch.ID)
	}
	for index, record := range results {
		if !utils.GetBool(record, "is_draft") {
			view.Triggers = append(view.Triggers, v.getTriggers(
				record.Copy(), v.Domain.GetMethod(), sch,
				utils.GetInt(record, ds.SchemaDBField),
				utils.GetInt(record, ds.DestTableDBField))...,
			)
		}
		go v.ConvertRecordToView(index, view, channel, record, tableName, cols, v.Domain.GetEmpty(), isWorkflow, params, createdIds)
	}
	for range results {
		rec := <-channel
		if !rec.IsEmpty {
			rec = v.getSharing(sch.ID, rec, v.Domain.GetUserID())
			view.Items = append(view.Items, rec)
		}
	}
}

func (s *ViewConvertor) getRedirection() string {
	if triggers.HasRedirection(s.Domain.GetDomainID()) {
		s, _ := triggers.GetRedirection(s.Domain.GetDomainID())
		return s
	}
	return ""
}

func (s *ViewConvertor) getSharing(schemaID string, rec sm.ViewItemModel, userID string) sm.ViewItemModel {
	id := rec.Values[utils.SpecialIDParam]
	m := map[string]interface{}{
		ds.UserDBField:      userID,
		ds.SchemaDBField:    schemaID,
		ds.DestTableDBField: id,
	}
	m["read_access"] = true
	m["update_access"] = true
	m["delete_access"] = true
	rec.Sharing = sm.SharingModel{
		SharedWithPath: fmt.Sprintf("/%s/%s?%s=%s&%s=disable", utils.MAIN_PREFIX, ds.DBUser.Name, utils.RootRowsParam,
			utils.ReservedParam, utils.RootScope),
		Body: m,
		ShallowPath: map[string]string{
			"shared_" + ds.UserDBField: fmt.Sprintf("/%s/%s?%s=%s&%s=enable&%s=enable", utils.MAIN_PREFIX, ds.DBUser.Name,
				utils.RootRowsParam, utils.ReservedParam, utils.RootShallow, utils.RootScope),
		},
		Path: fmt.Sprintf("/%s/%s?%s=%s&%s=enable", utils.MAIN_PREFIX, ds.DBShare.Name, utils.RootRowsParam, utils.ReservedParam, utils.RootShallow),
	}
	return rec
}

func (s *ViewConvertor) getFieldFill(sch sm.SchemaModel, key string) interface{} {
	if !sch.HasField(key) {
		return nil
	}
	var value interface{}
	f, _ := sch.GetField(key)

	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFieldAutoFill.Name, map[string]interface{}{
		ds.SchemaFieldDBField: f.ID,
	}, true); err == nil && len(res) > 0 {
		r := res[0]
		if val, ok := r["value"]; ok && val != nil {
			value = s.fromITF(val)
		} else if dest, ok := r["from_"+ds.DestTableDBField]; ok && dest != nil {
			if schID, err := strconv.Atoi(utils.ToString(r["from_"+ds.SchemaDBField])); err == nil && schID >= 0 {
				if schFrom, err := scheme.GetSchemaByID(utils.ToInt64(r["from_"+ds.SchemaDBField])); err == nil {
					if ff, err2 := schFrom.GetFieldByID(utils.GetInt(r, "from_"+ds.SchemaFieldDBField)); err2 == nil {
						if ress, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name, map[string]interface{}{
							utils.SpecialIDParam: dest,
						}, true); err == nil && len(ress) > 0 {
							value = s.fromITF(ress[0][ff.Name])
						}
					} else {
						if ress, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name, map[string]interface{}{
							utils.SpecialIDParam: dest,
						}, true); err == nil && len(ress) > 0 {
							value = s.fromITF(ress[0][ff.ID])
						}
					}
				}
			}
		} else if utils.GetBool(r, "first_own") {
			if schFrom, err := scheme.GetSchemaByID(utils.ToInt64(r["from_"+ds.SchemaDBField])); err == nil {
				if ff, err2 := schFrom.GetFieldByID(utils.GetInt(r, "from_"+ds.SchemaFieldDBField)); err2 == nil {
					if schFrom.Name == ds.DBUser.Name && ff.Name == utils.SpecialIDParam {
						value = s.Domain.GetUserID()
					} else {
						restr := filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name, utils.ToListAnonymized(restr), false); err == nil && len(rr) > 0 {
							value = s.fromITF(rr[0][ff.Name])
						}
					}
				} else {
					if schFrom.Name == ds.DBUser.Name {
						value = s.Domain.GetUserID()
					} else {
						restr := filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name, utils.ToListAnonymized(restr), false); err == nil && len(rr) > 0 {
							value = s.fromITF(rr[0][utils.SpecialIDParam])
						}
					}
				}
			}
		}
	}
	return value
}

func (s *ViewConvertor) getFieldsFill(sch sm.SchemaModel, values map[string]interface{}) map[string]interface{} {
	if !s.Domain.GetEmpty() {
		return values
	}
	for k := range values {
		values[k] = s.getFieldFill(sch, k)
	}
	return values
}

func (v *ViewConvertor) createShallowedViewItem(record utils.Record, tableName string, isWorkflow bool, max int64) utils.Record {
	ts := []sm.ManualTriggerModel{}
	label := record.GetString(sm.NAMEKEY)
	if record.GetString(sm.LABELKEY) != "" {
		label = record.GetString(sm.LABELKEY)
	}
	otherOrder := []string{}
	translatable := true
	if sch, err := scheme.GetSchema(tableName); err == nil {
		if f, err := sch.GetField("label"); err == nil {
			translatable = f.Translatable
		} else if f, err := sch.GetField("name"); err == nil {
			translatable = f.Translatable
		}
		if !utils.GetBool(record, "is_draft") {
			ts = v.getTriggers(record, v.Domain.GetMethod(), sch, utils.GetInt(record, ds.SchemaDBField), utils.GetInt(record, ds.DestTableDBField))
		}
		_, ok := v.Domain.GetParams().Get(utils.RootShallow)
		if ok {
			otherOrder := []string{}
			entity := v.Domain.GetDb().BuildSelectQueryWithRestriction(
				ds.DBEntityUser.Name,
				map[string]interface{}{
					ds.UserDBField: v.Domain.GetUserID(),
				}, true, ds.EntityDBField)
			fmt.Println(entity)
			if v.Domain.GetEmpty() {
				if res, err := v.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
					ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
						ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
							"is_view":              true,
							"dashboard_restricted": false,
						}, false, utils.SpecialIDParam),
						ds.FilterDBField + "_1": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
							ds.SchemaDBField: sch.ID,
						}, false, "view_"+ds.FilterDBField),
					}, false, ds.FilterDBField),
				}, false); err == nil {
					for _, r := range res {
						if f, err := schema.GetFieldByID(utils.GetInt(r, ds.SchemaFieldDBField)); err == nil {
							otherOrder = append(otherOrder, f.Name)
						}
					}
				}
			} else if res, err := v.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
				ds.FilterDBField: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBFilter.Name, map[string]interface{}{
					"is_view":              true,
					"dashboard_restricted": false,
				}, false, utils.SpecialIDParam),
				ds.FilterDBField + "_1": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
					utils.SpecialIDParam: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
						ds.UserDBField:   v.Domain.GetUserID(),
						ds.EntityDBField: entity,
					}, true, ds.WorkflowSchemaDBField),
				}, false, "view_"+ds.FilterDBField),
				ds.FilterDBField + "_2": v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
					utils.SpecialIDParam: v.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
						ds.SchemaDBField: sch.ID,
						"is_close":       false,
					}, false, ds.WorkflowSchemaDBField),
				}, false, "view_"+ds.FilterDBField),
			}, false); err == nil {
				for _, r := range res {
					if f, err := schema.GetFieldByID(utils.GetInt(r, ds.SchemaFieldDBField)); err == nil {
						otherOrder = append(otherOrder, f.Name)
					}
				}
			}
		}
	}
	view := sm.ViewModel{
		ID:           record.GetInt(utils.SpecialIDParam),
		Name:         record.GetString(sm.NAMEKEY),
		Label:        label,
		Workflow:     v.EnrichWithWorkFlowView(record, tableName, isWorkflow),
		Redirection:  v.getRedirection(),
		Translatable: translatable,
		Triggers:     ts,
		Max:          max,
	}

	if _, ok := v.Domain.GetParams().Get(utils.SpecialIDParam); ok && record[ds.SchemaDBField] != nil {
		if sch, err := scheme.GetSchemaByID(record.GetInt(ds.SchemaDBField)); err != nil {
			return nil
		} else {
			schema, id, order, _, addAction, _ := v.GetViewFields(sch.Name, false, utils.Results{record}) // FOUND IT
			o := []string{}
			for _, ord := range order {
				if len(otherOrder) == 0 || slices.Contains(otherOrder, ord) {
					o = append(o, ord)
				}
			}
			if _, ok := record["is_draft"]; ok && record.GetBool("is_draft") && !slices.Contains(addAction, "put") && v.Domain.IsOwn(false, false, utils.SELECT) {
				addAction = append(addAction, "put")
			}
			view.Description = fmt.Sprintf("%s shallowed data", tableName)
			view.IsWrapper = tableName == ds.DBTask.Name || tableName == ds.DBRequest.Name
			view.Path = utils.BuildPath(sch.Name, utils.ReservedParam)
			view.Schema = schema

			view.SchemaID = id
			view.SchemaName = tableName
			view.Actions = addAction
			view.ActionPath = utils.BuildPath(sch.Name, utils.ReservedParam)
			view.Order = o
			view.Consents = v.getConsent(utils.ToString(id), []utils.Record{record})
		}
	}
	return view.ToRecord()
}

func (d *ViewConvertor) ConvertRecordToView(index int, view *sm.ViewModel, channel chan sm.ViewItemModel,
	record utils.Record, tableName string, cols map[string]sm.FieldModel, isEmpty bool, isWorkflow bool, params utils.Params,
	createdIds []string) {

	vals, shallowVals, manyPathVals := make(map[string]interface{}), make(map[string]interface{}), make(map[string]string)
	manyVals := make(map[string]utils.Results)
	var datapath, historyPath, commentPath, synthesisPath string = "", "", "", ""
	schema, err := scheme.GetSchema(tableName)
	if !isEmpty {
		if err == nil {
			synthesisPath = d.getSynthesis(record, schema)
			historyPath = utils.BuildPath(ds.DBDataAccess.Name, utils.ReservedParam, utils.RootOrderParam+"=access_date", utils.RootDirParam+"=asc", utils.RootDestTableIDParam+"="+record.GetString(utils.SpecialIDParam), ds.RootID(ds.DBSchema.Name)+"="+utils.ToString(schema.ID))
			commentPath = utils.BuildPath(ds.DBComment.Name, utils.ReservedParam, utils.RootDestTableIDParam+"="+record.GetString(utils.SpecialIDParam), ds.RootID(ds.DBSchema.Name)+"="+utils.ToString(schema.ID))
		}
		vals[utils.SpecialIDParam] = record.GetString(utils.SpecialIDParam)
	}
	for _, field := range cols {
		if d, s, ok := d.HandleDBSchemaField(record, field, tableName, shallowVals); ok && d != "" {
			datapath = d
			shallowVals = s
			continue
		} else {
			shallowVals = s
		}
		shallowVals, manyVals, manyPathVals = d.HandleLinkField(record, field, tableName, isEmpty, shallowVals, manyVals, manyPathVals)

		if isEmpty {
			vals[field.Name] = nil
		} else if v, ok := record[field.Name]; ok {
			vals[field.Name] = v
		}
	}

	d.ApplyCommandRow(record, vals, params)
	d.getFilterByWFSchema(view, schema, record)
	vals = d.getOrder(view, record, vals)
	vals = d.getFieldsFill(schema, vals)
	channel <- sm.ViewItemModel{
		Values:        vals,
		DataPaths:     datapath,
		ValueShallow:  shallowVals,
		Sort:          int64(index),
		CommentsPath:  commentPath,
		HistoryPath:   historyPath,
		ValueMany:     manyVals,
		ValuePathMany: manyPathVals,
		Readonly:      IsReadonly(tableName, record, createdIds, d.Domain),
		Workflow:      d.EnrichWithWorkFlowView(record, tableName, isWorkflow),
		Draft:         utils.GetBool(record, "is_draft"),
		Synthesis:     synthesisPath,
		New:           d.GetNew(utils.GetString(record, utils.SpecialIDParam), schema.ID),
	}
}

func (s *ViewConvertor) GetNew(id string, schemaID string) bool {
	if id == "" {
		return false
	}
	if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBDataAccess.Name, map[string]interface{}{
		"write":             false,
		"update":            false,
		ds.DestTableDBField: id,
		ds.SchemaDBField:    schemaID,
		ds.UserDBField:      s.Domain.GetUserID(),
	}, false); err == nil && len(res) > 0 {
		return false
	}
	return true
}

func (s *ViewConvertor) fromITF(val interface{}) interface{} {
	if slices.Contains([]string{"true", "false"}, utils.ToString(val)) {
		return val == "true" // should set type
	} else if i, err := strconv.Atoi(utils.ToString(val)); err == nil && i >= 0 {
		return i // should set type
	} else {
		return utils.ToString(val) // should set type
	}
}

func (s *ViewConvertor) getOrder(view *sm.ViewModel, record utils.Record, values map[string]interface{}) map[string]interface{} {
	newOrder := ""
	if len(task.GetViewTask(utils.GetString(record, ds.SchemaDBField), utils.ToString(record[utils.SpecialIDParam]), s.Domain.GetUserID())) > 0 {
		vs := task.GetViewTask(utils.GetString(record, ds.SchemaDBField), utils.ToString(record[utils.SpecialIDParam]), s.Domain.GetUserID())
		if utils.GetBool(record, "is_list") {
			for _, fname := range vs {
				if val, ok := view.Schema[fname]; ok {
					utils.ToMap(val)["readonly"] = true
					utils.ToMap(val)["hidden"] = true
				}
				values[fname] = nil
			}
		} else {
			newOrder = strings.Join(vs, ",")
		}
	}
	if newOrder != "" {
		view.Order = strings.Split(newOrder, ",")
	}
	return values
}
func (s *ViewConvertor) getFilterByWFSchema(view *sm.ViewModel, schema sm.SchemaModel, record utils.Record) {
	tasks := task.GetTasks(schema.ID, utils.GetString(record, utils.SpecialIDParam))
	if tasks != nil {
		for _, task := range *tasks {
			if fields, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBFilterField.Name, map[string]interface{}{
				ds.FilterDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
					ds.WorkflowDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
						utils.SpecialIDParam: task.TaskID,
					}, false, ds.WorkflowDBField),
					ds.WorkflowDBField + "_1": s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
						ds.SchemaDBField: schema.ID,
					}, false, utils.SpecialIDParam),
				}, false, ds.FilterDBField),
			}, false); err == nil && len(fields) > 0 {
				for _, f := range schema.Fields {
					ok := false
					for _, ff := range fields {
						if f.ID == utils.GetString(ff, ds.SchemaFieldDBField) {
							ok = true
							break
						}
					}
					if !ok {
						if val, ok := view.Schema[f.Name]; ok {
							newOrder := []string{}
							for _, o := range view.Order {
								if o != f.Name {
									newOrder = append(newOrder, o)
								}
							}
							view.Order = newOrder
							m := utils.ToMap(val)
							m["readonly"] = true
						}
					}
				}
			}
		}
	}
}

func (s *ViewConvertor) getConsent(schemaID string, results utils.Results) []map[string]interface{} {
	if !s.Domain.GetEmpty() && len(results) != 1 {
		return []map[string]interface{}{}
	}
	if consents, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBConsent.Name, map[string]interface{}{
		ds.SchemaDBField: schemaID,
	}, false); err == nil && len(consents) > 0 {
		if len(results) > 0 {
			if consentsResp, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(
				ds.DBConsentResponse.Name,
				map[string]interface{}{
					ds.SchemaDBField:    schemaID,
					ds.DestTableDBField: results[0][utils.SpecialIDParam],
					ds.ConsentDBField:   utils.GetString(consents[0], utils.SpecialIDParam),
				}, false); err == nil && len(consentsResp) > 0 {
				return []map[string]interface{}{}
			}
		}
		cst := []map[string]interface{}{}
		for _, r := range consents {
			c := map[string]interface{}{}
			c["name"] = utils.GetString(r, "name")
			c["optionnal"] = utils.GetBool(r, "optionnal")
			c["body"] = map[string]interface{}{
				ds.SchemaDBField:  r[ds.SchemaDBField],
				ds.ConsentDBField: r[utils.SpecialIDParam],
			}
			c["action_path"] = fmt.Sprintf("/%s/%s?%s=%s", utils.MAIN_PREFIX, ds.DBConsentResponse.Name, utils.RootRowsParam, utils.ReservedParam)
			cst = append(cst, c)
		}
		return cst
	}
	return []map[string]interface{}{}
}

func (s *ViewConvertor) getSynthesis(record utils.Record, schema sm.SchemaModel) string {
	taskIDs := ""
	if schema.Name == ds.DBTask.Name {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.RequestDBField: record[ds.RequestDBField],
		}, false); err == nil && len(res) > 0 {
			is := []string{}
			for _, r := range res {
				is = append(is, utils.GetString(r, utils.SpecialIDParam))
			}
			if len(is) > 0 {
				taskIDs = strings.Join(is, ",")
			}
		}
	} else if schema.Name == ds.DBRequest.Name {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.RequestDBField: record[utils.SpecialIDParam],
		}, false); err == nil && len(res) > 0 {
			is := []string{}
			for _, r := range res {
				is = append(is, utils.GetString(r, utils.SpecialIDParam))
			}
			if len(is) > 0 {
				taskIDs = strings.Join(is, ",")
			}
		}
	} else {
		if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			utils.SpecialIDParam: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.DestTableDBField: record[utils.SpecialIDParam],
				ds.SchemaDBField:    schema.ID,
			}, false, utils.SpecialIDParam),
			utils.SpecialIDParam + "_1": s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.RequestDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
					ds.DestTableDBField: record[utils.SpecialIDParam],
					ds.SchemaDBField:    schema.ID,
				}, false, utils.SpecialIDParam),
			}, false, utils.SpecialIDParam),
		}, true); err == nil && len(res) > 0 {
			is := []string{}
			for _, r := range res {
				is = append(is, utils.GetString(r, utils.SpecialIDParam))
			}
			if len(is) > 0 {
				taskIDs = strings.Join(is, ",")
			}
		}
	}
	if taskIDs != "" { // means there is actually running task effective on these data
		return fmt.Sprintf("/%s/%s?%s=%s&scope=enable&%s=%s",
			utils.MAIN_PREFIX, ds.DBTask.Name,
			utils.RootRowsParam, taskIDs,
			utils.RootColumnsParam, "name,state,dbuser_id,dbentity_id,binded_to_email",
		)
	}
	return ""
}

func (d *ViewConvertor) HandleDBSchemaField(record utils.Record, field sm.FieldModel, tableName string, shallowVals map[string]interface{}) (string, map[string]interface{}, bool) {
	datapath := ""
	id, idOk := record[field.Name]
	dest, destOk := record[ds.DestTableDBField]
	if !strings.Contains(field.Name, ds.DBSchema.Name) || !idOk || id == nil {
		return datapath, shallowVals, false
	}
	schema, err := scheme.GetSchemaByID(utils.ToInt64(id))
	if err != nil {
		return datapath, shallowVals, false
	}
	shallowVals[ds.SchemaDBField] = utils.Record{"id": utils.ToString(schema.ID), "name": utils.ToString(schema.Name), "label": utils.ToString(schema.Label)}
	if destOk && dest != nil {
		datapath = utils.BuildPath(schema.Name, utils.ToString(dest))
		if t, err := d.Domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
			utils.SpecialIDParam: dest,
		}, false); err == nil && len(t) > 0 {
			shallowVals[ds.DestTableDBField] = utils.Record{
				utils.SpecialIDParam: utils.ToString(t[0][utils.SpecialIDParam]),
				sm.NAMEKEY:           utils.ToString(t[0][sm.NAMEKEY]),
				sm.LABELKEY:          utils.ToString(t[0][sm.NAMEKEY]),
				"data_ref":           "@" + utils.ToString(schema.ID) + ":" + utils.ToString(t[0][utils.SpecialIDParam]),
				"values_path":        utils.BuildPath(utils.ToString(schema.ID), utils.ToString(t[0][utils.SpecialIDParam]), utils.RootShallow+"=enable"),
			}
		}
	}
	return datapath, shallowVals, true
}

func (d *ViewConvertor) HandleLinkField(record utils.Record, field sm.FieldModel, tableName string, shallow bool,
	shallowVals map[string]interface{}, manyVals map[string]utils.Results, manyPathVals map[string]string) (map[string]interface{}, map[string]utils.Results, map[string]string) {
	if record.GetString(field.Name) == "" || field.GetLink() <= 0 || shallow {
		return shallowVals, manyVals, manyPathVals
	}
	link := scheme.GetTablename(utils.ToString(field.Link))

	if strings.Contains(field.Type, "many") {
		manyVals, manyPathVals = d.HandleManyField(record, field, tableName, link, manyVals, manyPathVals)
		return shallowVals, manyVals, manyPathVals
	}
	shallowVals = d.HandleOneField(record, field, link, shallowVals)
	return shallowVals, manyVals, manyPathVals
}

func (d *ViewConvertor) HandleManyField(record utils.Record, field sm.FieldModel, tableName, link string,
	manyVals map[string]utils.Results, manyPathVals map[string]string) (map[string]utils.Results, map[string]string) {
	if !d.Domain.IsShallowed() {
		l, _ := scheme.GetSchemaByID(field.GetLink())
		for _, f := range l.Fields {
			if field.Type == sm.ONETOMANY.String() && field.GetLink() > 0 {
				if strings.Contains(f.Name, tableName) && strings.Contains(f.Name, "_id") {
					manyPathVals[field.Name] = utils.BuildPath(
						link, utils.ReservedParam,
						f.Name+"="+record.GetString(utils.SpecialIDParam))
					break
				}
				continue
			}
			if strings.Contains(f.Name, tableName) || f.Name == utils.SpecialIDParam || f.GetLink() <= 0 {
				continue
			}
			lid, _ := scheme.GetSchemaByID(f.GetLink())
			if _, ok := manyVals[field.Name]; !ok {
				manyVals[field.Name] = utils.Results{}
			}
			if res, err := d.Domain.GetDb().SelectQueryWithRestriction(lid.Name, map[string]interface{}{
				utils.SpecialIDParam: d.Domain.GetDb().BuildSelectQueryWithRestriction(link, map[string]interface{}{
					ds.RootID(tableName): record.GetString(utils.SpecialIDParam),
				}, false)}, false); err == nil {
				for _, r := range res {
					manyVals[field.Name] = append(manyVals[field.Name], r)
				}
			}
		}
	}
	return manyVals, manyPathVals
}

func (d *ViewConvertor) HandleOneField(record utils.Record, field sm.FieldModel, link string, shallowVals map[string]interface{}) map[string]interface{} {
	v := record.GetString(field.Name)
	if r, err := d.Domain.GetDb().SelectQueryWithRestriction(link, map[string]interface{}{
		utils.SpecialIDParam: v,
	}, false); err == nil && len(r) > 0 {
		ref := fmt.Sprintf("@%v:%v", field.Link, r[0][utils.SpecialIDParam])
		shallowVals[field.Name] = utils.Record{
			utils.SpecialIDParam: r[0][utils.SpecialIDParam],
			sm.NAMEKEY:           r[0][sm.NAMEKEY],
			"data_ref":           ref,
		}
		if _, ok := r[0][sm.LABELKEY]; ok {
			shallowVals[field.Name].(utils.Record)[sm.LABELKEY] = r[0][sm.LABELKEY]
		}
	}
	return shallowVals
}

func (d *ViewConvertor) ApplyCommandRow(record utils.Record, vals map[string]interface{}, params utils.Params) {
	if cmd, ok := params.Get(utils.RootCommandRow); ok {
		decodedLine, _ := url.QueryUnescape(cmd)
		matches := strings.Split(decodedLine, " as ")
		if len(matches) > 1 {
			vals[matches[len(matches)-1]] = record[matches[len(matches)-1]]
		}
	}
}

func (d *ViewConvertor) getTriggers(record utils.Record, method utils.Method, fromSchema sm.SchemaModel, toSchemaID, destID int64) []sm.ManualTriggerModel {
	if _, ok := d.Domain.GetParams().Get(utils.SpecialIDParam); method == utils.DELETE || (!ok && method == utils.SELECT) {
		return []sm.ManualTriggerModel{}
	}
	if utils.UPDATE == method && d.Domain.GetIsDraftToPublished() {
		method = utils.CREATE
	}
	mt := []sm.ManualTriggerModel{}
	triggerService := triggers.NewTrigger(d.Domain)
	if res, err := triggerService.GetTriggers("manual", method, fromSchema.ID); err == nil {
		for _, r := range res {
			typ := utils.GetString(r, "type")
			switch typ {
			case "mail":
				if t, err := d.getMailTriggers(record, fromSchema, utils.GetString(r, "description"), utils.GetString(r, "name"),
					utils.GetInt(r, utils.SpecialIDParam), toSchemaID, destID); err == nil {
					mt = append(mt, t...)
				}
			}
		}
	}
	return mt
}

func (d *ViewConvertor) getMailTriggers(record utils.Record, fromSchema sm.SchemaModel, triggerDesc string, triggerName string, triggerID, toSchemaID, destID int64) ([]sm.ManualTriggerModel, error) {
	if sch, err := schema.GetSchema(ds.DBEmailSended.Name); err != nil {
		return nil, err
	} else {
		triggerService := triggers.NewTrigger(d.Domain)
		mails := triggerService.TriggerManualMail("manual", record, fromSchema, triggerID, toSchemaID, destID)
		bodies := []sm.ManualTriggerModel{}
		s := sch.ToMapRecord()
		for _, f := range sch.Fields {
			if f.GetLink() > 0 {
				if sch2, err := schema.GetSchemaByID(f.GetLink()); err == nil {
					s[f.Name].(map[string]interface{})["action_path"] = utils.BuildPath(sch2.Name, utils.ReservedParam, utils.RootShallow+"=enable")
					for _, f2 := range sch2.Fields {
						if f2.GetLink() > 0 && strings.Contains(f2.Name, "_id") && !strings.Contains(f2.Name, sch2.Name) {
							if sch3, err := schema.GetSchemaByID(f2.GetLink()); err == nil {
								s[f.Name].(map[string]interface{})["data_schema"] = sch2.ToMapRecord()
								s[f.Name].(map[string]interface{})["values_path"] = utils.BuildPath(sch3.Name, utils.ReservedParam, utils.RootShallow+"=enable")
							}
						}
					}
				}
			}
			if strings.Contains(f.Type, "upload") {
				s[f.Name].(map[string]interface{})["action_path"] = fmt.Sprintf("/%s/%s/import?rows=all&columns=%s", utils.MAIN_PREFIX, sch.Name, f.Name)
				s[f.Name].(map[string]interface{})["values_path"] = fmt.Sprintf("/%s/%s/import?rows=all&columns=%s", utils.MAIN_PREFIX, sch.Name, f.Name)
			}
		}
		for _, m := range mails {
			bodies = append(bodies, sm.ManualTriggerModel{
				Name:        triggerName,
				Description: triggerDesc,
				Type:        "mail",
				Schema:      s,
				Body:        m,
				ActionPath:  utils.BuildPath(sch.Name, utils.ReservedParam),
			})
		}
		return bodies, nil
	}

}

func IsReadonly(tableName string, record utils.Record, createdIds []string, d utils.DomainITF) bool {
	if d.GetEmpty() || utils.GetBool(record, "is_draft") {
		return false
	}
	readonly := true
	for _, meth := range []utils.Method{utils.CREATE, utils.UPDATE} {
		if d.VerifyAuth(tableName, "", "", meth, record.GetString(utils.SpecialIDParam)) {
			if (meth == utils.CREATE && d.GetEmpty()) || meth == utils.UPDATE {
				readonly = false
				break
			}
		}
	}
	if sch, err := scheme.GetSchema(tableName); err == nil {
		m := map[string]interface{}{
			"is_close":     false,
			ds.UserDBField: d.GetUserID(),
		}
		if tableName == ds.DBTask.Name {
			delete(m, ds.UserDBField)
			m[utils.SpecialIDParam+"_1"] = d.GetDb().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.EntityDBField: d.GetDb().BuildSelectQueryWithRestriction(
					ds.DBEntityUser.Name,
					map[string]interface{}{
						ds.UserDBField: d.GetUserID(),
					}, true, ds.EntityDBField),
				ds.UserDBField: d.GetUserID(),
			}, true, utils.SpecialIDParam)
			m[utils.SpecialIDParam] = record[utils.SpecialIDParam]
			m[ds.WorkflowSchemaDBField] = d.GetDb().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
				utils.SpecialIDParam: record[ds.WorkflowSchemaDBField],
			}, false, utils.SpecialIDParam)

			if res, err := d.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, m, false); err != nil || len(res) == 0 {
				return true
			} else if slices.Contains(createdIds, record.GetString(utils.SpecialIDParam)) {
				return false
			}
		} else {
			m[ds.DestTableDBField] = record[utils.SpecialIDParam]
			m[ds.SchemaDBField] = sch.ID
			if res, err := d.GetDb().SelectQueryWithRestriction(ds.DBRequest.Name, m, false); err != nil || len(res) == 0 {
				return true
			} else if slices.Contains(createdIds, record.GetString(utils.SpecialIDParam)) {
				return false
			}
		}
	}
	return readonly || record["state"] == "completed" || record["state"] == "dismiss" || record["state"] == "refused" || record["state"] == "canceled"
}

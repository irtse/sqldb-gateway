package view_convertor

import (
	"fmt"
	"net/url"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/history"
	"sqldb-ws/domain/domain_service/triggers"
	scheme "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	connector "sqldb-ws/infrastructure/connector/db"
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
		return utils.Results{}
	}
	if ids, ok := params.Get(utils.SpecialIDParam); ok || v.Domain.GetMethod() != utils.SELECT {
		if len(ids) == 0 {
			for _, r := range results {
				ids += r.GetString(utils.SpecialIDParam) + ","
			}
			ids = connector.RemoveLastChar(ids)
		}
		history.NewDataAccess(schema.GetID(), strings.Split(ids, ","), v.Domain)
	}
	if v.Domain.IsShallowed() {
		return v.transformShallowedView(results, tableName, isWorkflow)
	}
	return v.transformFullView(results, &schema, isWorkflow, params)
}

func (v *ViewConvertor) transformFullView(results utils.Results, schema *sm.SchemaModel, isWorkflow bool, params utils.Params) utils.Results {
	schemes, id, order, _, addAction, _ := v.GetViewFields(schema.Name, false, results)
	commentBody := map[string]interface{}{}
	if len(results) == 1 {
		commentBody = map[string]interface{}{
			ds.UserDBField:      utils.ToInt64(v.Domain.GetUserID()),
			ds.SchemaDBField:    utils.ToInt64(schema.ID),
			ds.DestTableDBField: utils.GetInt(results[0], utils.SpecialIDParam),
		}
	}
	max, _ := history.CountMaxDataAccess(schema, []string{}, v.Domain)

	view := sm.NewView(id, schema.Name, schema.Label, schema, max, []sm.ManualTriggerModel{})
	view.Schema = schemes
	view.Redirection = getRedirection(v.Domain.GetDomainID())
	view.Order = CompareOrder(schema, order, v.Domain)
	view.Actions = addAction
	view.CommentBody = commentBody
	view.Shortcuts = v.GetShortcuts(schema.ID, addAction)
	view.Consents = v.getConsent(schema.ID, results)

	v.ProcessResultsConcurrently(results, schema, isWorkflow, &view, params)
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

func (v *ViewConvertor) TransformMultipleSchema(results utils.Results, schema *sm.SchemaModel, isWorkflow bool, params utils.Params) utils.Results {
	max, _ := history.CountMaxDataAccess(schema, []string{}, v.Domain)
	view := sm.ViewModel{
		Items: []sm.ViewItemModel{},
		Max:   max,
	}
	v.ProcessResultsConcurrently(results, schema, isWorkflow, &view, params)
	// if there is only one item in the view, we can set the view readonly to the item readonly
	sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Sort < view.Items[j].Sort })
	return utils.Results{view.ToRecord()}
}

func (v *ViewConvertor) ProcessResultsConcurrently(results utils.Results, schema *sm.SchemaModel,
	isWorkflow bool, view *sm.ViewModel, params utils.Params) {
	const maxConcurrent = 5
	runtime.GOMAXPROCS(maxConcurrent)
	channel := make(chan sm.ViewItemModel, len(results))
	defer close(channel)
	go func() {
		if err := recover(); err != nil {
			fmt.Printf("panic occurred: %v\n%v\n", err, string(debug.Stack()))
		}
	}()
	createdIds := history.GetCreatedAccessData(schema.ID, v.Domain)
	for index, record := range results {
		if !utils.GetBool(record, "is_draft") {
			view.Triggers = append(view.Triggers, triggers.NewTrigger(v.Domain).GetViewTriggers(
				record.Copy(), v.Domain.GetMethod(), schema,
				utils.GetInt(record, ds.SchemaDBField),
				utils.GetInt(record, ds.DestTableDBField))...,
			)
		}
		go v.ConvertRecordToView(index, view, channel, record, schema, v.Domain.GetEmpty(), isWorkflow, params, createdIds)
	}
	for range results {
		rec := <-channel
		if !rec.IsEmpty {
			rec = GetSharing(schema.ID, rec, v.Domain)
			view.Items = append(view.Items, rec)
		}
	}
}

func (v *ViewConvertor) transformShallowedView(results utils.Results, tableName string, isWorkflow bool) utils.Results {
	res := utils.Results{}
	max := int64(0)
	sch, err := scheme.GetSchema(tableName)
	if err == nil {
		return res
	}
	max, _ = history.CountMaxDataAccess(&sch, []string{}, v.Domain)
	scheme, id, order, _, addAction, _ := v.GetViewFields(tableName, false, utils.Results{})
	for _, record := range results {
		if _, ok := record["is_draft"]; ok && record.GetBool("is_draft") && !v.Domain.IsOwn(false, false, utils.SELECT) {
			continue
		}
		if record.GetString(sm.NAMEKEY) == "" {
			res = append(res, record)
			continue
		}
		newView := v.createShallowedViewItem(record, &sch, isWorkflow, max)
		if _, ok := record["is_draft"]; ok && record.GetBool("is_draft") && !slices.Contains(addAction, "put") && v.Domain.IsOwn(false, false, utils.SELECT) {
			addAction = append(addAction, "put")
		}
		newView.Schema = scheme
		newView.SchemaID = id
		newView.Actions = addAction
		newView.Order = CompareOrder(&sch, order, v.Domain)
		newView.Consents = v.getConsent(utils.ToString(id), []utils.Record{record})
		res = append(res, newView.ToRecord())
	}
	return res
}

func (v *ViewConvertor) createShallowedViewItem(record utils.Record, schema *sm.SchemaModel, isWorkflow bool, max int64) sm.ViewModel {
	ts := []sm.ManualTriggerModel{}
	label := record.GetString(sm.NAMEKEY)
	if record.GetString(sm.LABELKEY) != "" {
		label = record.GetString(sm.LABELKEY)
	}
	translatable := true
	if f, err := schema.GetField("label"); err == nil {
		translatable = f.Translatable
	} else if f, err := schema.GetField("name"); err == nil {
		translatable = f.Translatable
	}
	if !utils.GetBool(record, "is_draft") {
		ts = triggers.NewTrigger(v.Domain).GetViewTriggers(
			record, v.Domain.GetMethod(), schema, utils.GetInt(record, ds.SchemaDBField), utils.GetInt(record, ds.DestTableDBField))
	}
	view := sm.NewView(record.GetInt(utils.SpecialIDParam), record.GetString(sm.NAMEKEY), label, schema, max, ts)
	view.Path = utils.BuildPath(schema.Name, utils.ReservedParam)
	view.Redirection = getRedirection(v.Domain.GetDomainID())
	view.Workflow = v.EnrichWithWorkFlowView(record, schema.Name, isWorkflow)
	view.Translatable = translatable
	return view
}

func (d *ViewConvertor) ConvertRecordToView(index int, view *sm.ViewModel, channel chan sm.ViewItemModel,
	record utils.Record, schema *sm.SchemaModel, isEmpty bool, isWorkflow bool, params utils.Params,
	createdIds []string) {

	vals, shallowVals, manyPathVals := make(map[string]interface{}), make(map[string]interface{}), make(map[string]string)
	manyVals := make(map[string]utils.Results)
	var datapath, historyPath, commentPath, synthesisPath string = "", "", "", ""
	if !isEmpty {
		synthesisPath = d.getSynthesis(record, schema)
		historyPath = utils.BuildPath(ds.DBDataAccess.Name, utils.ReservedParam, utils.RootOrderParam+"=access_date", utils.RootDirParam+"=asc", utils.RootDestTableIDParam+"="+record.GetString(utils.SpecialIDParam), ds.RootID(ds.DBSchema.Name)+"="+utils.ToString(schema.ID))
		commentPath = utils.BuildPath(ds.DBComment.Name, utils.ReservedParam, utils.RootDestTableIDParam+"="+record.GetString(utils.SpecialIDParam), ds.RootID(ds.DBSchema.Name)+"="+utils.ToString(schema.ID))
		vals[utils.SpecialIDParam] = record.GetString(utils.SpecialIDParam)
	}
	for _, field := range schema.Fields {
		if d, s, ok := d.HandleDBSchemaField(record, field, shallowVals); ok && d != "" {
			datapath = d
			shallowVals = s
			continue
		} else {
			shallowVals = s
		}
		shallowVals, manyVals, manyPathVals = d.HandleLinkField(record, field, schema, isEmpty, shallowVals, manyVals, manyPathVals)

		if isEmpty {
			vals[field.Name] = nil
		} else if v, ok := record[field.Name]; ok {
			vals[field.Name] = v
		}
	}

	d.ApplyCommandRow(record, vals, params)
	newOrder, vals := GetOrder(schema, record, vals, []string{}, d.Domain)
	if len(newOrder) > 0 {
		view.Order = newOrder
		vals = d.getFieldsFill(schema, vals)
		channel <- sm.ViewItemModel{
			Values:        vals,
			DataPaths:     datapath,
			ValueShallow:  shallowVals,
			Sort:          int64(index),
			DataRef:       d.getLinkPath(record, schema), // to redirect
			CommentsPath:  commentPath,
			HistoryPath:   historyPath,
			ValueMany:     manyVals,
			ValuePathMany: manyPathVals,
			Readonly:      IsReadonly(schema.Name, record, createdIds, d.Domain),
			Workflow:      d.EnrichWithWorkFlowView(record, schema.Name, isWorkflow),
			Draft:         utils.GetBool(record, "is_draft"),
			Synthesis:     synthesisPath,
			New:           history.GetNew(utils.GetString(record, utils.SpecialIDParam), schema.ID, d.Domain),
		}
	}
}

func (s *ViewConvertor) getLinkPath(record utils.Record, sch *sm.SchemaModel) string {
	if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
		utils.SpecialIDParam: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.DestTableDBField: utils.GetString(record, utils.SpecialIDParam),
			ds.SchemaDBField:    sch.GetID(),
		}, false, utils.SpecialIDParam),
		ds.RequestDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
			ds.DestTableDBField: utils.GetString(record, utils.SpecialIDParam),
			ds.SchemaDBField:    sch.GetID(),
		}, false, utils.SpecialIDParam),
	}, true); err == nil && len(res) > 0 {
		firstTaskToWrap := res[0]
		if s, err := scheme.GetSchema(ds.DBTask.Name); err == nil {
			return "@" + s.ID + ":" + utils.GetString(firstTaskToWrap, utils.SpecialIDParam)
		}
	}
	return "@" + sch.ID + ":" + utils.GetString(record, utils.SpecialIDParam)
}

func (s *ViewConvertor) getFieldsFill(sch *sm.SchemaModel, values map[string]interface{}) map[string]interface{} {
	if !s.Domain.GetEmpty() {
		return values
	}
	for k := range values {
		values[k] = s.getFieldFill(sch, k)
	}
	return values
}

func (s *ViewConvertor) getFieldFill(sch *sm.SchemaModel, key string) interface{} {
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
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name,
							utils.ToListAnonymized(filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)), false); err == nil && len(rr) > 0 {
							value = s.fromITF(rr[0][ff.Name])
						}
					}
				} else {
					if schFrom.Name == ds.DBUser.Name {
						value = s.Domain.GetUserID()
					} else {
						if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schFrom.Name,
							utils.ToListAnonymized(filter.NewFilterService(s.Domain).RestrictionByEntityUser(schFrom, []string{}, true)),
							false); err == nil && len(rr) > 0 {
							value = s.fromITF(rr[0][utils.SpecialIDParam])
						}
					}
				}
			}
		}
	}
	return value
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

func (s *ViewConvertor) getSynthesis(record utils.Record, schema *sm.SchemaModel) string {
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
			utils.RootColumnsParam, "name,state,dbuser_id,dbentity_id,binded_to_email,closing_date",
		)
	}
	return ""
}

func (d *ViewConvertor) HandleDBSchemaField(record utils.Record, field sm.FieldModel, shallowVals map[string]interface{}) (string, map[string]interface{}, bool) {
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
		if t, err := d.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
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

func (d *ViewConvertor) HandleLinkField(record utils.Record, field sm.FieldModel, schema *sm.SchemaModel, shallow bool,
	shallowVals map[string]interface{}, manyVals map[string]utils.Results, manyPathVals map[string]string) (map[string]interface{}, map[string]utils.Results, map[string]string) {
	if (record.GetString(field.Name) == "" && !strings.Contains(field.Type, "many")) || field.GetLink() <= 0 || shallow {
		return shallowVals, manyVals, manyPathVals
	}
	link := scheme.GetTablename(utils.ToString(field.Link))

	if strings.Contains(field.Type, "many") {
		manyVals, manyPathVals = d.HandleManyField(record, field, schema, link, manyVals, manyPathVals)
		return shallowVals, manyVals, manyPathVals
	}
	shallowVals = d.HandleOneField(record, field, link, shallowVals)
	return shallowVals, manyVals, manyPathVals
}

func (d *ViewConvertor) HandleManyField(record utils.Record, field sm.FieldModel, schema *sm.SchemaModel, link string,
	manyVals map[string]utils.Results, manyPathVals map[string]string) (map[string]utils.Results, map[string]string) {
	if !d.Domain.IsShallowed() {
		l, _ := scheme.GetSchemaByID(field.GetLink())
		for _, f := range l.Fields {
			if field.Type == sm.ONETOMANY.String() && field.GetLink() > 0 {
				if strings.Contains(f.Name, schema.Name) && strings.Contains(f.Name, "_id") {
					manyPathVals[field.Name] = utils.BuildPath(
						link, utils.ReservedParam,
						f.Name+"="+record.GetString(utils.SpecialIDParam))
					break
				}
				continue
			}
			if strings.Contains(f.Name, schema.Name) || f.Name == utils.SpecialIDParam || f.GetLink() <= 0 {
				continue
			}
			lid, _ := scheme.GetSchemaByID(f.GetLink())
			if _, ok := manyVals[field.Name]; !ok {
				manyVals[field.Name] = utils.Results{}
			}
			// field link is a many to many... such as authors
			// link is related tableName : demo_authors
			// f is the field from some_authors that not correspond to the schema.Name _ id : exemple demo_id -> demo
			// lid is the link of this field for exemple : user_id

			// on veut former une requête comme suit : SELECT * FROM dbuser WHERE id IN (SELECT dbuser_id FROM demo_authors WHERE dbdemo_id = ?)
			fmt.Println(lid.Name, link, schema.Name)
			// HERE IS REGULARY MALFORMED REQUEST FOR AUTHORS
			if res, err := d.Domain.GetDb().SelectQueryWithRestriction(lid.Name, map[string]interface{}{
				utils.SpecialIDParam: d.Domain.GetDb().BuildSelectQueryWithRestriction(link, map[string]interface{}{
					"!" + ds.RootID(lid.Name): nil,
					ds.RootID(schema.Name):    record.GetString(utils.SpecialIDParam),
				}, false, ds.RootID(lid.Name))}, false); err == nil {
				for _, r := range res {
					manyVals[field.Name] = append(manyVals[field.Name], r)
				}
			}
			if res, err := d.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(link, map[string]interface{}{
				ds.RootID(lid.Name):    nil,
				ds.RootID(schema.Name): record.GetString(utils.SpecialIDParam),
			}, false); err == nil {
				for _, r := range res {
					manyVals[field.Name] = append(manyVals[field.Name], utils.Record{"name": utils.GetString(r, "name")})
				}
			}
		}
	}
	return manyVals, manyPathVals
}

func (d *ViewConvertor) HandleOneField(record utils.Record, field sm.FieldModel, link string, shallowVals map[string]interface{}) map[string]interface{} {
	v := record.GetString(field.Name)
	if r, err := d.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(link, map[string]interface{}{
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
			m[utils.SpecialIDParam+"_1"] = d.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.EntityDBField: d.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(
					ds.DBEntityUser.Name,
					map[string]interface{}{
						ds.UserDBField: d.GetUserID(),
					}, true, ds.EntityDBField),
				ds.UserDBField: d.GetUserID(),
			}, true, utils.SpecialIDParam)
			m[utils.SpecialIDParam] = record[utils.SpecialIDParam]
			m[ds.WorkflowSchemaDBField] = d.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
				utils.SpecialIDParam: record[ds.WorkflowSchemaDBField],
			}, false, utils.SpecialIDParam)

			if res, err := d.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, m, false); err != nil || len(res) == 0 {
				return true
			} else if slices.Contains(createdIds, record.GetString(utils.SpecialIDParam)) {
				return false
			}
		} else {
			m[ds.DestTableDBField] = record[utils.SpecialIDParam]
			m[ds.SchemaDBField] = sch.ID
			if res, err := d.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, m, false); err != nil || len(res) == 0 {
				return true
			} else if slices.Contains(createdIds, record.GetString(utils.SpecialIDParam)) {
				return false
			}
		}
	}
	return readonly || record["state"] == "completed" || record["state"] == "dismiss" || record["state"] == "refused" || record["state"] == "canceled"
}

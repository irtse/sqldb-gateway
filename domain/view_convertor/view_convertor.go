package view_convertor

import (
	"fmt"
	"net/url"
	"runtime"
	"runtime/debug"
	"sort"
	scheme "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"strings"
)

type ViewConvertor struct {
	Domain utils.DomainITF
}

func NewViewConvertor(domain utils.DomainITF) *ViewConvertor {
	return &ViewConvertor{Domain: domain}
}

func (v *ViewConvertor) TransformToView(results utils.Results, tableName string, isWorkflow bool) utils.Results {
	schema, err := scheme.GetSchema(tableName)
	if err != nil {
		if results == nil {
			return utils.Results{}
		}
		return results
	}
	if ids, ok := v.Domain.GetParams()[utils.SpecialIDParam]; ok || v.Domain.GetMethod() != utils.SELECT {
		v.NewDataAccess(schema.GetID(), strings.Split(ids, ","), v.Domain.GetMethod()) // FOUND IT !
	}
	if v.Domain.IsShallowed() {
		return v.transformShallowedView(results, tableName, isWorkflow)
	}
	return v.transformFullView(results, schema, tableName, isWorkflow)
}

func (v *ViewConvertor) transformFullView(results utils.Results, schema sm.SchemaModel, tableName string, isWorkflow bool) utils.Results {
	schemes, id, order, cols, addAction, readonly := v.GetViewFields(tableName, false)
	view := sm.ViewModel{
		ID:          id,
		Name:        schema.Label,
		Label:       schema.Label,
		Description: fmt.Sprintf("%s data", tableName),
		Schema:      schemes,
		IsWrapper:   tableName == ds.DBTask.Name || tableName == ds.DBRequest.Name,
		SchemaID:    id,
		SchemaName:  tableName,
		ActionPath:  v.BuildPath(tableName, utils.ReservedParam),
		Readonly:    readonly,
		Order:       order,
		Actions:     addAction,
		Items:       []sm.ViewItemModel{},
		Shortcuts:   v.GetShortcuts(),
	}

	v.processResultsConcurrently(results, tableName, cols, isWorkflow, &view)
	// if there is only one item in the view, we can set the view readonly to the item readonly
	if len(view.Items) == 1 {
		view.Readonly = view.Items[0].Readonly
	}
	if view.Readonly { // if the view is readonly, we remove the actions
		view.Actions = []string{"get"}
	}
	sort.SliceStable(view.Items, func(i, j int) bool { return view.Items[i].Sort < view.Items[j].Sort })
	return utils.Results{view.ToRecord()}
}

func (v *ViewConvertor) transformShallowedView(results utils.Results, tableName string, isWorkflow bool) utils.Results {
	res := utils.Results{}
	for _, record := range results {
		if record.GetString(sm.NAMEKEY) == "" {
			res = append(res, record)
			continue
		}
		res = append(res, v.createShallowedViewItem(record, tableName, isWorkflow))
	}
	return res
}

func (v *ViewConvertor) processResultsConcurrently(results utils.Results, tableName string,
	cols map[string]sm.FieldModel, isWorkflow bool, view *sm.ViewModel) {
	const maxConcurrent = 5
	runtime.GOMAXPROCS(maxConcurrent)
	channel := make(chan sm.ViewItemModel, len(results))
	defer close(channel)
	go func() {
		if err := recover(); err != nil {
			fmt.Printf("panic occurred: %v\n%v\n", err, string(debug.Stack()))
		}
	}()
	for index, record := range results {
		go v.ConvertRecordToView(index, channel, record, tableName, cols, v.Domain.GetEmpty(), isWorkflow)
	}
	for range results {
		rec := <-channel
		if !rec.IsEmpty {
			view.Items = append(view.Items, rec)
		}
	}
}

func (v *ViewConvertor) createShallowedViewItem(record utils.Record, tableName string, isWorkflow bool) utils.Record {
	label := record.GetString(sm.NAMEKEY)
	if record.GetString(sm.LABELKEY) != "" {
		label = record.GetString(sm.LABELKEY)
	}
	view := sm.ViewModel{
		ID:       record.GetInt(utils.SpecialIDParam),
		Name:     record.GetString(sm.NAMEKEY),
		Label:    label,
		Workflow: v.EnrichWithWorkFlowView(record, tableName, isWorkflow),
	}
	if record[ds.SchemaDBField] != nil {
		if sch, err := scheme.GetSchemaByID(record.GetInt(ds.SchemaDBField)); err != nil {
			return nil
		} else {
			schema, id, order, _, addAction, readonly := v.GetViewFields(sch.Name, false) // FOUND IT
			view.Description = fmt.Sprintf("%s shallowed data", tableName)
			view.IsWrapper = tableName == ds.DBTask.Name || tableName == ds.DBRequest.Name
			view.Path = v.BuildPath(sch.Name, utils.ReservedParam)
			view.Schema = schema
			view.SchemaID = id
			view.SchemaName = tableName
			view.Actions = addAction
			view.ActionPath = v.BuildPath(sch.Name, utils.ReservedParam)
			view.Readonly = readonly
			view.Order = order
		}

	}
	return view.ToRecord()
}

func (d *ViewConvertor) ConvertRecordToView(index int, channel chan sm.ViewItemModel,
	record utils.Record, tableName string, cols map[string]sm.FieldModel, isEmpty bool, isWorkflow bool) {
	vals, shallowVals, manyPathVals := make(map[string]interface{}), make(map[string]interface{}), make(map[string]string)
	manyVals := make(map[string]utils.Results)
	var ok bool = false
	var datapath, historyPath string = "", ""

	if !isEmpty {
		schema, err := scheme.GetSchema(tableName)
		if err == nil {
			historyPath = d.BuildPath(ds.DBDataAccess.Name, utils.ReservedParam, utils.RootOrderParam+"=access_date", utils.RootDirParam+"=asc", utils.RootDestTableIDParam+"="+record.GetString(utils.SpecialIDParam), ds.RootID(ds.DBSchema.Name)+"="+utils.ToString(schema.ID))
		}
		vals[utils.SpecialIDParam] = record.GetString(utils.SpecialIDParam)
	}
	for _, field := range cols {
		if datapath, ok = d.HandleDBSchemaField(record, field, tableName, shallowVals); ok {
			continue
		}
		d.HandleLinkField(record, field, tableName, isEmpty, shallowVals, manyVals, manyPathVals)
		if isEmpty {
			vals[field.Name] = nil
		} else if v, ok := record[field.Name]; ok {
			vals[field.Name] = v
		}
	}
	d.ApplyCommandRow(record, vals)
	channel <- sm.ViewItemModel{
		Values:        vals,
		DataPaths:     datapath,
		ValueShallow:  shallowVals,
		Sort:          int64(index),
		HistoryPath:   historyPath,
		ValueMany:     manyVals,
		ValuePathMany: manyPathVals,
		Readonly:      d.IsReadonly(tableName, record),
		Workflow:      d.EnrichWithWorkFlowView(record, tableName, isWorkflow),
	}
}

func (d *ViewConvertor) HandleDBSchemaField(record utils.Record, field sm.FieldModel, tableName string, shallowVals map[string]interface{}) (string, bool) {
	datapath := ""
	id, idOk := record[field.Name]
	dest, destOk := record[ds.DestTableDBField]
	if !strings.Contains(field.Name, ds.DBSchema.Name) || !idOk || id == nil {
		return datapath, false
	}
	schema, err := scheme.GetSchemaByID(utils.ToInt64(id))
	if err != nil {
		return datapath, false
	}
	shallowVals[ds.SchemaDBField] = utils.Record{"id": utils.ToString(schema.ID), "name": utils.ToString(schema.Name), "label": utils.ToString(schema.Label)}
	if destOk && dest != nil {
		datapath = d.BuildPath(schema.Name, utils.ToString(dest))
		if t, err := d.Domain.GetDb().SelectQueryWithRestriction(schema.Name, map[string]interface{}{
			utils.SpecialIDParam: dest,
		}, false); err == nil && len(t) > 0 {
			shallowVals[ds.DestTableDBField] = utils.Record{
				utils.SpecialIDParam: utils.ToString(t[0][utils.SpecialIDParam]),
				sm.NAMEKEY:           utils.ToString(t[0][sm.NAMEKEY]),
				sm.LABELKEY:          utils.ToString(t[0][sm.NAMEKEY]),
				"data_ref":           "@" + utils.ToString(schema.ID) + ":" + utils.ToString(t[0][utils.SpecialIDParam])}
		}
	}
	return datapath, true
}

func (d *ViewConvertor) HandleLinkField(record utils.Record, field sm.FieldModel, tableName string, shallow bool,
	shallowVals map[string]interface{}, manyVals map[string]utils.Results, manyPathVals map[string]string) {
	if record.GetString(field.Name) == "" || field.GetLink() <= 0 || shallow {
		return
	}
	link := scheme.GetTablename(utils.ToString(field.Link))
	if strings.Contains(field.Type, "many") {
		d.HandleManyField(record, field, tableName, link, manyVals, manyPathVals)
		return
	}
	d.HandleOneField(record, field, link, shallowVals)
}

func (d *ViewConvertor) HandleManyField(record utils.Record, field sm.FieldModel, tableName, link string, manyVals map[string]utils.Results, manyPathVals map[string]string) {
	if !d.Domain.IsShallowed() {
		l, _ := scheme.GetSchemaByID(field.GetLink())
		for _, f := range l.Fields {
			if field.Type == sm.ONETOMANY.String() && field.GetLink() > 0 {
				if strings.Contains(f.Name, tableName) && strings.Contains(f.Name, "_id") {
					manyPathVals[field.Name] = d.BuildPath(
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
			views := []string{utils.SpecialIDParam, sm.NAMEKEY}
			if lid.HasField(sm.LABELKEY) {
				views = append(views, sm.LABELKEY)
			}
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
}

func (d *ViewConvertor) HandleOneField(record utils.Record, field sm.FieldModel, link string, shallowVals map[string]interface{}) {
	if r, err := d.Domain.GetDb().SelectQueryWithRestriction(link, map[string]interface{}{
		utils.SpecialIDParam: record.GetString(field.Name),
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
}

func (d *ViewConvertor) ApplyCommandRow(record utils.Record, vals map[string]interface{}) {
	if cmd, ok := d.Domain.GetParams()[utils.RootCommandRow]; ok {
		decodedLine, _ := url.QueryUnescape(cmd)
		matches := strings.Split(decodedLine, " as ")
		if len(matches) > 1 {
			vals[matches[len(matches)-1]] = record[matches[len(matches)-1]]
		}
	}
}

func (d *ViewConvertor) IsReadonly(tableName string, record utils.Record) bool {
	readonly := true
	for _, meth := range []utils.Method{utils.CREATE, utils.UPDATE} {
		if d.Domain.VerifyAuth(tableName, "", "", meth, record.GetString(utils.SpecialIDParam)) {
			if (meth == utils.CREATE && d.Domain.GetEmpty()) || meth == utils.UPDATE {
				readonly = false
				break
			}
		}
	}
	return readonly || record["state"] == "completed" || record["state"] == "dismiss" || record["state"] == "refused" || record["state"] == "canceled"
}

package schema_service

import (
	"fmt"
	"math/rand"
	"slices"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
	"strings"
)

type SchemaService struct {
	servutils.SpecializedService
	Fields []interface{}
}

// DONE - UNDER 100 LINES - NOT TESTED
func (s *SchemaService) Entity() utils.SpecializedServiceInfo { return ds.DBSchema }

func (s *SchemaService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.DELETE {
		if s.Domain.IsSuperAdmin() {
			return record, nil, false
		}
		return record, fmt.Errorf("cannot delete schema field on schemaDB"), false
	}
	s.Fields = []interface{}{}
	if fields, ok := record["fields"]; ok && fields != nil {
		s.Fields = utils.ToList(fields)
		delete(record, "fields")
	}
	return s.SpecializedService.VerifyDataIntegrity(record, tablename)
}

func (s *SchemaService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for _, res := range results {
		schema, err := schserv.GetSchema(utils.ToString(res[sm.NAMEKEY]))
		if err != nil {
			return
		}
		s.Domain.HandleRecordAttributes(utils.Record{"is_custom": true})
		s.Domain.DeleteSuperCall(utils.AllParams(ds.DBSchemaField.Name).Enrich(map[string]interface{}{
			ds.SchemaDBField: schema.ID,
		}))
		s.Domain.DeleteSuperCall(utils.AllParams(ds.DBPermission.Name).Enrich(map[string]interface{}{
			sm.NAMEKEY: "%" + utils.ToString(res[sm.NAMEKEY]) + "%",
		}))
		s.Domain.DeleteSuperCall(utils.AllParams(ds.DBView.Name).Enrich(map[string]interface{}{
			sm.NAMEKEY: "%" + schema.Name + "%",
		}))
		schserv.DeleteSchema(utils.ToString(res[sm.NAMEKEY]))
	}
}

func (s *SchemaService) SpecializedCreateRow(record map[string]interface{}, tableName string) {

	schema := sm.SchemaModel{}.Deserialize(record)
	res, err := s.Domain.CreateSuperCall(utils.GetTableTargetParameters(record[sm.NAMEKEY]), record)
	if err != nil || len(res) == 0 {
		return
	}
	schema, err = schserv.SetSchema(record)
	if err != nil {
		return
	}
	if MissingField[utils.GetString(record, "name")] != nil {
		l := MissingField[utils.GetString(record, "name")]
		for _, r := range l {
			if ft, err := schserv.GetSchema(utils.GetString(r, "foreign_table")); err == nil {
				r["link_id"] = utils.ToInt64(ft.ID)
				delete(r, "foreign_table")
				s.Domain.CreateSuperCall(utils.AllParams(ds.DBSchemaField.Name), r)
			}
		}
	}
	for _, field := range s.Fields {
		f := utils.ToMap(field)
		f[ds.SchemaDBField] = schema.ID
		field, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
		if err != nil || len(field) == 0 {
			continue
		}
		schema = schema.SetField(field[0])
	}
	if schema.Name != ds.DBDataAccess.Name {
		if !slices.Contains([]string{ds.DBView.Name, ds.DBRequest.Name, ds.DBTask.Name,
			ds.DBFilter.Name, ds.DBFilterField.Name, ds.DBViewAttribution.Name, ds.DBNotification.Name}, schema.Name) {
			var index int64 = 2
			if count, err := s.Domain.GetDb().SimpleMathQuery(
				"COUNT", ds.DBView.Name, map[string]interface{}{ds.SchemaDBField: utils.ToString(schema.ID)},
				false); err == nil && len(count) > 0 && (utils.ToInt64(count[0]["result"])+1) > 1 {
				index = utils.ToInt64(count[0]["result"]) + 1
			}
			cat := "data"
			if utils.ToString(record["name"])[:2] == "db" {
				cat = "technical data"
			}
			wfs := utils.Record{}
			// create workflow except for the following schemas
			if !slices.Contains([]string{
				ds.DBTask.Name,
				ds.DBRequest.Name,
				ds.DBFilter.Name,
				ds.DBFilterField.Name,
				ds.DBViewAttribution.Name,
				ds.DBNotification.Name}, schema.Name) {
				if w, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBWorkflow.Name),
					NewWorkflow(
						"create "+schema.Label,
						"new "+schema.Label+" workflow",
						schema.GetID()),
				); err == nil && len(w) > 0 {
					wfs = w[0]
				}
			}
			if schema.Category != "" {
				// EMPTY SUBMIT FORM WITH A FILTER on request
				if resquestSchema, err := schserv.GetSchema(ds.DBRequest.Name); err == nil {
					filter := "Submit " + strings.ReplaceAll(strings.ReplaceAll(schema.Name, "_", ""), "db", "") + " datas."
					body := map[string]interface{}{
						ds.SchemaDBField: resquestSchema.ID,
						"name":           "filter " + filter,
					}
					if f, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBFilter.Name), body); err == nil && len(f) > 0 {
						body["name"] = "view " + utils.ToString(body["name"])
						body["is_view"] = true
						if vf, err := s.Domain.CreateSuperCall(utils.AllParams(ds.DBFilter.Name), body); err == nil && len(f) > 0 {
							wf, _ := resquestSchema.GetField(ds.WorkflowDBField)
							m := map[string]interface{}{
								ds.FilterDBField:      vf[0][utils.SpecialIDParam],
								ds.SchemaFieldDBField: wf.ID,
							}
							s.Domain.CreateSuperCall(utils.AllParams(ds.DBFilterField.Name), m)
							if wfs[utils.SpecialIDParam] != nil {
								m[ds.FilterDBField] = f[0][utils.SpecialIDParam]
								m["value"] = wfs[utils.SpecialIDParam]
								s.Domain.CreateSuperCall(utils.AllParams(ds.DBFilterField.Name), m)
								newViewSubmit := NewView("create a "+schema.Label,
									"create "+schema.Label,
									filter, "", resquestSchema.GetID(), index, false, true, false, false,
									vf[0][utils.SpecialIDParam], f[0][utils.SpecialIDParam], record[utils.SpecialIDParam])
								s.Domain.CreateSuperCall(utils.AllParams(ds.DBView.Name), newViewSubmit)
							}
						}
					}
				}
				newView := NewView(schema.Label, schema.Label, "View description for "+strings.ReplaceAll(strings.ReplaceAll(schema.Name, "_", " "), "db", "")+" datas.",
					cat, schema.GetID(), index, true, false, true, false, nil, nil, nil)

				s.Domain.CreateSuperCall(utils.AllParams(ds.DBView.Name), newView)
				if schema.CanOwned {
					r := rand.New(rand.NewSource(9999999999))
					newView = NewView("my "+schema.Label, "my "+schema.Label,
						"View description for my "+schema.Label+" datas.",
						"my data", schema.GetID(), int64(r.Int()), true, false, true, true, nil, nil, nil)
					s.Domain.CreateSuperCall(utils.AllParams(ds.DBView.Name), newView)
				}

			}
		}

	}
	UpdatePermissions(utils.Record{}, utils.ToString(record[sm.NAMEKEY]), []string{sm.LEVELOWN, sm.LEVELNORMAL}, s.Domain)
	s.AbstractSpecializedService.SpecializedCreateRow(record, tableName)
}

func (s *SchemaService) SpecializedUpdateRow(datas []map[string]interface{}, record map[string]interface{}) {
	schema, err := schserv.GetSchema(utils.ToString(record[sm.NAMEKEY]))
	if err != nil {
		res, err := s.Domain.UpdateSuperCall(utils.GetTableTargetParameters(record[sm.NAMEKEY]), record)
		if err != nil || len(res) == 0 {
			return
		}
		schema, err = schserv.SetSchema(res[0])
		if err != nil {
			return
		}
		for _, field := range s.Fields {
			f := utils.ToMap(field)
			f[ds.SchemaDBField] = schema.ID
			if schema.HasField(utils.ToString(f[sm.NAMEKEY])) {
				s.Domain.UpdateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
			} else {
				s.Domain.CreateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
			}
			schema = schema.SetField(f)
		}
	}
	UpdatePermissions(utils.Record{}, utils.ToString(record[sm.NAMEKEY]), []string{sm.LEVELOWN, sm.LEVELNORMAL}, s.Domain)
	s.AbstractSpecializedService.SpecializedUpdateRow(datas, record)
}

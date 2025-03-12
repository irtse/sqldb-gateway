package schema_service

import (
	"fmt"
	"slices"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"strings"
)

type SchemaService struct {
	servutils.SpecializedService
	Fields []interface{}
}

// DONE - UNDER 100 LINES - NOT TESTED
func (s *SchemaService) Entity() utils.SpecializedServiceInfo { return ds.DBSchema }

func (s *SchemaService) ShouldVerify() bool { return false }

func (s *SchemaService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.DELETE {
		return record, fmt.Errorf("cannot delete schema field on schemaDB"), false
	}
	if fields, ok := record["fields"]; ok && fields != nil {
		s.Fields = utils.ToList(fields)
		delete(record, "fields")
	}
	return record, nil, true
}

func (s *SchemaService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	s.Domain.HandleRecordAttributes(utils.Record{"is_custom": true})
	s.Domain.DeleteSuperCall(
		utils.GetRowTargetParameters(ds.DBPermission.Name, sm.NAMEKEY).Enrich(
			map[string]interface{}{sm.NAMEKEY: "%" + tableName + "%"}))
}

func (s *SchemaService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	schema := sm.SchemaModel{}.Deserialize(record)
	res, err := s.Domain.CreateSuperCall(utils.GetTableTargetParameters(record[sm.NAMEKEY]), record)
	if err != nil || len(res) == 0 {
		return
	}
	schserv.LoadCache(utils.ReservedParam, s.Domain.GetDb())
	for _, field := range s.Fields {
		f := utils.ToMap(field)
		if typ, ok := f[sm.TYPEKEY]; !ok || typ == "" || strings.Contains(utils.ToString(typ), "many") {
			continue
		}
		f[ds.SchemaDBField] = schema.ID
		s.Domain.CreateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
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
			cat := "global data"
			if utils.ToString(record["name"])[:2] == "db" {
				cat = "technical data"
			}
			newView := NewView("view "+schema.Label, "View description for "+schema.Label+" datas.",
				cat, schema.GetID(), index, true, false, true, false, false,
			)
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBView.Name), newView)
		}
		// create workflow except for the following schemas
		if !slices.Contains([]string{
			ds.DBTask.Name,
			ds.DBRequest.Name,
			ds.DBFilter.Name,
			ds.DBFilterField.Name,
			ds.DBViewAttribution.Name,
			ds.DBNotification.Name}, schema.Name) {
			s.Domain.CreateSuperCall(utils.AllParams(ds.DBWorkflow.Name),
				NewWorkflow(
					"create "+schema.Label,
					"new "+schema.Label+" workflow",
					schema.GetID()),
			)
		}
	}
	UpdatePermissions(utils.Record{}, utils.ToString(record[sm.NAMEKEY]), []string{sm.LEVELOWN, sm.LEVELNORMAL}, s.Domain)
}

func (s *SchemaService) SpecializedUpdateRow(datas []map[string]interface{}, record map[string]interface{}) {
	schema, err := schserv.GetSchema(utils.ToString(record[sm.NAMEKEY]))
	if err != nil {
		s.Domain.UpdateSuperCall(utils.GetTableTargetParameters(record[sm.NAMEKEY]), record)
		for _, field := range s.Fields {
			f := utils.ToMap(field)
			if typ, ok := f[sm.TYPEKEY]; !ok || typ == "" || strings.Contains(utils.ToString(typ), "many") {
				continue
			}
			f[ds.SchemaDBField] = schema.ID
			if schema.HasField(utils.ToString(f[sm.NAMEKEY])) {
				s.Domain.UpdateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
			} else {
				s.Domain.CreateSuperCall(utils.AllParams(ds.DBSchemaField.Name), f)
			}
		}
	}
	if datas != nil {
		schserv.LoadCache(utils.ReservedParam, s.Domain.GetDb())
	}
	UpdatePermissions(utils.Record{}, utils.ToString(record[sm.NAMEKEY]), []string{sm.LEVELOWN, sm.LEVELNORMAL}, s.Domain)
}

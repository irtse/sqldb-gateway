package schema_service

import (
	"fmt"
	"slices"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
)

type SchemaService struct{ servutils.SpecializedService }

// DONE - UNDER 100 LINES - NOT TESTED
func (s *SchemaService) Entity() utils.SpecializedServiceInfo { return ds.DBSchema }
func (s *SchemaService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.DELETE {
		if schema, err := schserv.GetSchema(tablename); err != nil || !ds.IsRootDB(schema.Name) {
			return record, fmt.Errorf("cannot delete schema field %v", err), false
		}
		return record, nil, false
	}
	return servutils.CheckAutoLoad(tablename, record, s.Domain)
}

func (s *SchemaService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	s.Domain.HandleRecordAttributes(utils.Record{"is_custom": true})
	s.Domain.DeleteSuperCall(
		utils.GetRowTargetParameters(ds.DBPermission.Name, sm.NAMEKEY).Enrich(
			map[string]interface{}{sm.NAMEKEY: "%" + tableName + "%"}))
}

func (s *SchemaService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	s.Domain.CreateSuperCall(utils.GetTableTargetParameters(record[sm.NAMEKEY]),
		utils.Record{
			sm.NAMEKEY: record[sm.NAMEKEY],
			"fields":   []interface{}{},
		})
	schema := sm.SchemaModel{}.Deserialize(record)
	if schema.Name != ds.DBDataAccess.Name {
		if !slices.Contains([]string{ds.DBView.Name, ds.DBRequest.Name, ds.DBTask.Name,
			ds.DBFilter.Name, ds.DBFilterField.Name, ds.DBViewAttribution.Name, ds.DBNotification.Name}, schema.Name) {
			index := 2
			if count, err := s.Domain.GetDb().SimpleMathQuery(
				"COUNT", ds.DBView.Name, map[string]interface{}{ds.SchemaDBField: fmt.Sprintf("%v", schema.ID)},
				false); err == nil && len(count) > 0 && (int(count[0]["result"].(float64))+1) > 1 {
				index = int(count[0]["result"].(float64)) + 1
			}
			cat := "global data"
			if fmt.Sprintf("%v", record["name"])[:2] == "db" {
				cat = "technical data"
			}
			newView := NewView(
				"view "+schema.Label,
				"View description for "+schema.Label+" datas.",
				cat, schema.ID, index, true, false, true, false, false,
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
					schema.ID),
			)
		}
	}
	s.SpecializedUpdateRow(nil, record)
}

func (s *SchemaService) SpecializedUpdateRow(datas []map[string]interface{}, record map[string]interface{}) {
	if datas != nil {
		schserv.LoadCache(utils.ReservedParam, s.Domain.GetDb())
	}
	UpdatePermissions(utils.Record{}, fmt.Sprintf("%v", record[sm.NAMEKEY]),
		[]string{sm.LEVELOWN, sm.LEVELNORMAL}, s.Domain)
}

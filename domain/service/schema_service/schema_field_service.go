package schema_service

import (
	"fmt"
	"math/rand"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"strconv"
	"strings"
)

// DONE - UNDER 100 LINES - NOT TESTED
type SchemaFields struct{ servutils.SpecializedService }

func (s *SchemaFields) Entity() utils.SpecializedServiceInfo { return ds.DBSchemaField }

func (s *SchemaFields) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.DELETE { // delete root schema field
		if i, err := strconv.Atoi(s.Domain.GetParams()[utils.RootRowsParam]); err != nil {
			return record, err, false
		} else if schema, err := sch.GetSchemaByFieldID(int64(i)); err != nil || !ds.IsRootDB(schema.Name) {
			return record, fmt.Errorf("cannot delete root schema field  %v", err), false
		}
		return record, nil, false
	}
	// validate schema field
	utils.Add(record, sm.TYPEKEY, record[sm.TYPEKEY],
		func(i interface{}) bool { return i != nil && i != "" },
		func(i interface{}) interface{} { return utils.PrepareEnum(fmt.Sprintf("%v", i)) })
	utils.Add(record, sm.LABELKEY, record[sm.LABELKEY],
		func(i interface{}) bool { return i != nil && i != "" },
		func(i interface{}) interface{} { return strings.Replace(fmt.Sprintf("%v", i), "_", " ", -1) })
	if rec, err := sch.ValidateBySchema(record, tablename, s.Domain.GetMethod(), s.Domain.VerifyAuth); err != nil && !s.Domain.GetAutoload() {
		return rec, err, false
	}
	return record, nil, true
}

func (s *SchemaFields) SpecializedCreateRow(record map[string]interface{}, tableName string) { // THERE
	if schema, err := s.write(record, record, false); err != nil || schema == nil {
		return
	} else if record[sm.NAMEKEY] == ds.UserDBField || record[sm.NAMEKEY] == ds.EntityDBField { // create view
		r := rand.New(rand.NewSource(9999999999))
		newView := NewView("my"+schema.Name,
			"View description for my "+schema.Name+" datas.",
			"my data", schema.ID, r.Int(), true, false, true, false, true)
		s.Domain.CreateSuperCall(utils.AllParams(ds.DBView.Name), newView)
	}
}

func (s *SchemaFields) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	for _, r := range results {
		s.write(r, record, true)
	}
}

func (s *SchemaFields) write(r map[string]interface{}, record map[string]interface{}, isUpdate bool) (*sm.SchemaModel, error) {
	schema, err := sch.GetSchemaByID(r[ds.SchemaDBField].(int64))
	if err != nil {
		return nil, err
	}
	readLevels := []string{sm.LEVELNORMAL}
	if level, ok := record["read_level"]; ok && level != "" && level != sm.LEVELOWN {
		readLevels = append(readLevels, strings.Replace(fmt.Sprintf("%v", level), "'", "", -1))
	}
	UpdatePermissions(record, schema.Name, readLevels, s.Domain)
	if isUpdate {
		newRecord := utils.ToRecord(record, map[string]interface{}{
			sm.TYPEKEY: r[sm.TYPEKEY],
			sm.NAMEKEY: r[sm.NAMEKEY],
		})
		s.Domain.UpdateSuperCall(utils.GetColumnTargetParameters(schema.Name, r[sm.NAMEKEY]), newRecord)
	} else {
		s.Domain.CreateSuperCall(utils.GetColumnTargetParameters(schema.Name, r[sm.NAMEKEY]), record)
	}
	sch.LoadCache(schema.Name, s.Domain.GetDb())
	return &schema, nil
}

func (s *SchemaFields) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for _, record := range results { // delete all columns
		schema, err := sch.GetSchemaByID(record[ds.SchemaDBField].(int64))
		if err != nil { // schema not found
			s.Domain.DeleteSuperCall(utils.GetColumnTargetParameters(schema.Name, record[sm.NAMEKEY]))
			s.Domain.DeleteSuperCall(
				utils.AllParams(schema.Name).Enrich(map[string]interface{}{
					sm.NAMEKEY: "%" + record[sm.NAMEKEY].(string) + "%",
				}),
			)
			if schema.HasField(ds.UserDBField) || schema.HasField(ds.EntityDBField) { // delete view
				s.Domain.DeleteSuperCall(utils.AllParams(ds.DBView.Name).Enrich(map[string]interface{}{
					sm.NAMEKEY: "my " + schema.Name,
				}))
			}
			sch.LoadCache(schema.Name, s.Domain.GetDb()) // reload schema
		}
	}
}

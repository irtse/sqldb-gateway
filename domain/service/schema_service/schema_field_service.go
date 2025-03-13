package schema_service

import (
	"fmt"
	"slices"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
	"strings"
)

// DONE - UNDER 100 LINES - NOT TESTED
type SchemaFields struct{ servutils.SpecializedService }

func (s *SchemaFields) ShouldVerify() bool { return true }

func (s *SchemaFields) Entity() utils.SpecializedServiceInfo { return ds.DBSchemaField }

func (s *SchemaFields) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetMethod() == utils.DELETE { // delete root schema field
		return record, fmt.Errorf("cannot delete root schema field"), false
	}
	utils.Add(record, sm.TYPEKEY, record[sm.TYPEKEY],
		func(i interface{}) bool { return i != nil && i != "" },
		func(i interface{}) interface{} { return utils.PrepareEnum(utils.ToString(i)) })
	if !strings.Contains(sm.DataTypeToEnum(), utils.ToString(record[sm.TYPEKEY])) {
		return record, fmt.Errorf("invalid type"), false
	}
	utils.Add(record, sm.LABELKEY, record[sm.LABELKEY],
		func(i interface{}) bool { return true },
		func(i interface{}) interface{} {
			if i == nil || i == "" {
				i = utils.ToString(record[sm.NAMEKEY])
			}
			return strings.Replace(utils.ToString(i), "_", " ", -1)
		})
	if !slices.Contains(ds.NOAUTOLOADROOTTABLESSTR, tablename) {
		if rec, err := sch.ValidateBySchema(record, tablename, s.Domain.GetMethod(), s.Domain.VerifyAuth); err != nil && !s.Domain.GetAutoload() {
			return rec, err, false
		}
	}
	return record, nil, true
}

func (s *SchemaFields) SpecializedCreateRow(record map[string]interface{}, tableName string) { // THERE
	s.Write(record, record, false)
}

func (s *SchemaFields) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	for _, r := range results {
		s.Write(r, record, true)
	}
}

func (s *SchemaFields) Write(r map[string]interface{}, record map[string]interface{}, isUpdate bool) (*sm.SchemaModel, error) {
	schema, err := sch.GetSchemaByID(utils.ToInt64(r[ds.SchemaDBField]))
	if err != nil {
		return nil, err
	}
	if typ, ok := record[sm.TYPEKEY]; !ok || strings.Contains(utils.ToString(typ), "many") || schema.HasField(utils.ToString(record[sm.NAMEKEY])) {
		return nil, fmt.Errorf("field already exists")
	}
	readLevels := []string{sm.LEVELNORMAL}
	if level, ok := record["read_level"]; ok && level != "" && level != sm.LEVELOWN && slices.Contains(sm.READLEVELACCESS, utils.ToString(level)) {
		readLevels = append(readLevels, strings.Replace(utils.ToString(level), "'", "", -1))
	}
	UpdatePermissions(record, schema.Name, readLevels, s.Domain)
	if !slices.Contains(ds.NOAUTOLOADROOTTABLESSTR, utils.ToString(record[sm.NAMEKEY])) {
		if isUpdate {
			newRecord := utils.ToRecord(record, map[string]interface{}{
				sm.TYPEKEY: r[sm.TYPEKEY],
				sm.NAMEKEY: r[sm.NAMEKEY],
			})
			s.Domain.UpdateSuperCall(utils.GetColumnTargetParameters(schema.Name, r[sm.NAMEKEY]), newRecord)
		} else {
			s.Domain.CreateSuperCall(utils.GetColumnTargetParameters(schema.Name, r[sm.NAMEKEY]), record)
		}
	}
	schema = schema.SetField(r)
	return &schema, nil
}

func (s *SchemaFields) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for _, record := range results { // delete all columns
		schema, err := sch.GetSchemaByID(utils.ToInt64(record[ds.SchemaDBField]))
		if err != nil { // schema not found
			s.Domain.DeleteSuperCall(utils.GetColumnTargetParameters(schema.Name, record[sm.NAMEKEY]))
			s.Domain.DeleteSuperCall(
				utils.AllParams(schema.Name).Enrich(map[string]interface{}{
					sm.NAMEKEY: "%" + utils.ToString(record[sm.NAMEKEY]) + "%",
				}),
			)
			if schema.HasField(ds.UserDBField) || schema.HasField(ds.EntityDBField) { // delete view
				s.Domain.DeleteSuperCall(utils.AllParams(ds.DBView.Name).Enrich(map[string]interface{}{
					sm.NAMEKEY: "my " + schema.Name,
				}))
			}
			sch.DeleteSchemaField(tableName, utils.ToString(record[sm.NAMEKEY]))
		}
	}
}

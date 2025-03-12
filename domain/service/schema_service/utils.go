package schema_service

import (
	"fmt"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
)

func NewView(name string, desc string, category string, schemaDB int64, index int64,
	indexable bool, empty bool, isList bool, readonly bool, ownView bool) utils.Record {
	return utils.Record{
		sm.NAMEKEY:       name,
		ds.SchemaDBField: schemaDB,
		"description":    desc,
		"category":       category,
		"index":          index,
		"indexable":      indexable,
		"is_empty":       empty,
		"is_list":        isList,
		"readonly":       readonly,
		"own_view":       ownView,
	}
}

func NewViewFromRecord(schema sm.SchemaModel, record utils.Record) utils.Record {
	return utils.Record{
		"id":          record["id"],
		"new":         []string{},
		"name":        record["name"],
		"label":       record["name"],
		"description": record["description"],
		"is_empty":    record["is_empty"],
		"index":       record["index"],
		"is_list":     record["is_list"],
		"readonly":    record["readonly"],
		"category":    record["category"],
		"filter_path": fmt.Sprintf("/%s/%s?%s=%s&%s=enable&%s=%v",
			utils.MAIN_PREFIX, ds.DBFilter.Name, utils.RootRowsParam,
			utils.ReservedParam, utils.RootShallow, ds.SchemaDBField, schema.ID),
	}
}

func NewWorkflow(name string, desc string, schemaDB int64) utils.Record {
	return utils.Record{
		sm.NAMEKEY:       name,
		"description":    desc,
		ds.SchemaDBField: schemaDB,
	}
}

func UpdatePermissions(record utils.Record, schemaName string, readLevels []string, domain utils.DomainITF) {
	for role, mainPerms := range sm.MAIN_PERMS { // update permissions
		for _, l := range readLevels {
			rec := utils.ToRecord(
				map[string]interface{}{
					utils.SELECT.String(): l,
					sm.NAMEKEY:            schemaName + ":" + l + ":" + role,
				}, mainPerms.Anonymized(),
			)
			if col, ok := record[sm.NAMEKEY]; ok {
				rec[sm.NAMEKEY] = schemaName + ":" + utils.ToString(col) + ":" + l + ":" + role
			}
			domain.CreateSuperCall(utils.AllParams(ds.DBPermission.Name), rec)
		}
	}
}

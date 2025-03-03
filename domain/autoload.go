package domain

import (
	"fmt"
	"os"

	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"sqldb-ws/infrastructure/service"
)

func Load() {
	database := connector.Open(nil)
	defer database.Conn.Close()

	initializeTables(database)
	domainInstance := initializeDomain(database)
	initializeRootTables(database, domainInstance)
	createSuperAdmin(domainInstance)
}

func initializeTables(database *connector.Database) {
	for _, table := range []sm.SchemaModel{ds.DBSchema, ds.DBSchemaField, ds.DBPermission, ds.DBView, ds.DBWorkflow} {
		service := service.NewTableService(database, true, "", table.Name, table.ToRecord())
		service.NoLog = true
		service.Update()
	}
}

func initializeDomain(database *connector.Database) *SpecializedDomain {
	domainInstance := Domain(true, os.Getenv("SUPERADMIN_NAME"), false, nil)
	domainInstance.AutoLoad = true
	schserv.LoadCache(utils.ReservedParam, database)
	return domainInstance
}

func initializeRootTables(database *connector.Database, domainInstance *SpecializedDomain) {
	var wfNew, viewNew bool
	rootTables := append(ds.ROOTTABLES, ds.DEMOROOTTABLES...)

	for _, table := range rootTables {
		if _, err := schserv.GetSchema(table.Name); err != nil {
			if !createRootTable(domainInstance, table, table.ToRecord()) {
				continue
			}
			schserv.LoadCache(table.Name, database)
			if table.Name == ds.DBWorkflow.Name {
				wfNew = true
			}
			if table.Name == ds.DBView.Name {
				viewNew = true
			}
		}
	}

	updateRootTables(rootTables, domainInstance, wfNew, viewNew)
}

func createRootTable(domainInstance *SpecializedDomain, table sm.SchemaModel, record utils.Record) bool {
	params := utils.AllParams(ds.DBSchema.Name)
	params[utils.RootRawView] = "enable"
	res, err := domainInstance.Call(params, record, utils.CREATE)
	if err != nil || len(res) == 0 {
		return false
	}
	addFields(table, domainInstance, sm.SchemaModel{}.Deserialize(res[0]), false)
	return true
}

func updateRootTables(rootTables []sm.SchemaModel, domainInstance *SpecializedDomain, wfNew, viewNew bool) {
	for _, table := range rootTables {
		schema, err := schserv.GetSchema(table.Name)
		if err != nil {
			continue
		}
		if table.Name == ds.DBWorkflow.Name && wfNew {
			createWorkflowView(domainInstance, schema)
		}
		if table.Name == ds.DBView.Name && viewNew {
			createView(domainInstance, schema)
		}
		addFields(table, domainInstance, schema, true)
	}
}

func createWorkflowView(domainInstance *SpecializedDomain, schema sm.SchemaModel) {
	params := utils.Params{
		utils.RootTableParam: ds.DBView.Name,
		utils.RootRowsParam:  utils.ReservedParam,
		utils.RootRawView:    "enable",
	}
	newWorkflow := utils.Record{
		sm.NAMEKEY:       "workflow",
		"indexable":      true,
		"description":    fmt.Sprintf("View description for %s datas.", ds.DBWorkflow.Name),
		"category":       "workflow",
		"is_empty":       false,
		"index":          0,
		"is_list":        true,
		"readonly":       false,
		ds.SchemaDBField: schema.ID,
	}
	domainInstance.Call(params, newWorkflow, utils.CREATE)
}

func createView(domainInstance *SpecializedDomain, schema sm.SchemaModel) {
	params := utils.AllParams(ds.DBWorkflow.Name).RootRaw()
	newView := utils.Record{
		sm.NAMEKEY:       fmt.Sprintf("create %s", ds.DBView.Name),
		"description":    fmt.Sprintf("new %s workflow", ds.DBView.Name),
		ds.SchemaDBField: schema.ID,
	}
	domainInstance.Call(params, newView, utils.CREATE)
}

func createSuperAdmin(domainInstance *SpecializedDomain) {
	params := utils.AllParams(ds.DBUser.Name)
	params[utils.RootRawView] = "enable"
	domainInstance.Call(params, utils.Record{
		"name":        os.Getenv("SUPERADMIN_NAME"),
		"email":       os.Getenv("SUPERADMIN_EMAIL"),
		"super_admin": true,
		"password":    os.Getenv("SUPERADMIN_PASSWORD"),
	}, utils.CREATE)
}

func addFields(table sm.SchemaModel, domainInstance *SpecializedDomain, schema sm.SchemaModel, enforce bool) {
	for _, col := range table.Fields {
		if schema.HasField(col.Name) && enforce {
			continue
		}
		colData := col.ToRecord()
		if col.ForeignTable != "" {
			foreignSchema, err := schserv.GetSchema(col.ForeignTable)
			if err == nil {
				colData["link_id"] = foreignSchema.ID
			}
		}
		colData[ds.SchemaDBField] = schema.ID
		domainInstance.AutoLoad = true
		domainInstance.Call(utils.AllParams(ds.SchemaFieldDBField), colData, utils.CREATE)
	}
}

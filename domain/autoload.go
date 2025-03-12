package domain

import (
	"fmt"
	"os"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"

	"github.com/schollz/progressbar/v3"
)

func Load() {
	database := connector.Open(nil)
	defer database.Conn.Close()

	domainInstance := initializeDomain()
	bar := progressbar.Default(int64(len(ds.NOAUTOLOADROOTTABLES) + len(ds.ROOTTABLES) + len(ds.DEMOROOTTABLES) + len(ds.DBRootViews)))
	initializeTables(domainInstance, bar)               // Create tables if they don't exist, needed for the next step
	initializeRootTables(database, domainInstance, bar) // Create root tables if they don't exist, needed for the next step
	createSuperAdmin(domainInstance, bar)
	createRootView(domainInstance, bar)
	schserv.LoadCache(utils.ReservedParam, database)
}

func initializeTables(domainInstance *SpecializedDomain, bar *progressbar.ProgressBar) {
	for _, table := range ds.NOAUTOLOADROOTTABLES {
		domainInstance.CreateSuperCall(utils.GetTableTargetParameters(table.Name), table.ToRecord())
		bar.Add(1)
	}
}

func initializeDomain() *SpecializedDomain {
	domainInstance := Domain(true, os.Getenv("SUPERADMIN_NAME"), false, nil)
	domainInstance.AutoLoad = true
	return domainInstance
}

func initializeRootTables(database *connector.Database, domainInstance *SpecializedDomain, bar *progressbar.ProgressBar) {
	var wfNew, viewNew bool
	rootTables := append(ds.ROOTTABLES, ds.DEMOROOTTABLES...)
	for _, table := range rootTables {
		if _, err := schserv.GetSchema(table.Name); err != nil {
			if createRootTable(domainInstance, table.ToRecord()) {
				schserv.LoadCache(table.Name, database)
				if table.Name == ds.DBWorkflow.Name {
					wfNew = true
				}
				if table.Name == ds.DBView.Name {
					viewNew = true
				}
				schema, err := schserv.GetSchema(table.Name)
				if err == nil {
					if table.Name == ds.DBWorkflow.Name && wfNew {
						createWorkflowView(domainInstance, schema)
					}
					if table.Name == ds.DBView.Name && viewNew {
						createView(domainInstance, schema)
					}
				}
			}
		}
		bar.Add(1)
	}
}

func createRootTable(domainInstance *SpecializedDomain, record utils.Record) bool {
	params := utils.AllParams(ds.DBSchema.Name).RootRaw()
	res, err := domainInstance.CreateSuperCall(params, record)
	return !(err != nil || len(res) == 0)
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
	domainInstance.CreateSuperCall(params, newWorkflow)
}

func createRootView(domainInstance *SpecializedDomain, bar *progressbar.ProgressBar) {
	for _, view := range ds.DBRootViews {
		params := utils.AllParams(ds.DBView.Name).RootRaw()
		domainInstance.CreateSuperCall(params, view)
		bar.Add(1)
	}
}

func createView(domainInstance *SpecializedDomain, schema sm.SchemaModel) {
	params := utils.AllParams(ds.DBWorkflow.Name).RootRaw()
	newView := utils.Record{
		sm.NAMEKEY:       fmt.Sprintf("create %s", ds.DBView.Name),
		"description":    fmt.Sprintf("new %s workflow", ds.DBView.Name),
		ds.SchemaDBField: schema.ID,
	}
	domainInstance.CreateSuperCall(params, newView)
}

func createSuperAdmin(domainInstance *SpecializedDomain, bar *progressbar.ProgressBar) {
	params := utils.AllParams(ds.DBUser.Name)
	params[utils.RootRawView] = "enable"
	domainInstance.CreateSuperCall(params, utils.Record{
		"name":        os.Getenv("SUPERADMIN_NAME"),
		"email":       os.Getenv("SUPERADMIN_EMAIL"),
		"super_admin": true,
		"password":    os.Getenv("SUPERADMIN_PASSWORD"),
	})
	bar.Add(1)
}

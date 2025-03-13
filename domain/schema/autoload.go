package schema

import (
	"fmt"
	"os"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"time"

	"github.com/schollz/progressbar/v3"
)

func Load(domainInstance utils.DomainITF) {
	db := connector.Open(nil)
	defer db.Close()
	progressbar.OptionSetMaxDetailRow(1)
	bar := progressbar.NewOptions64(
		int64(len(ds.NOAUTOLOADROOTTABLES)+len(ds.ROOTTABLES)+len(ds.DEMOROOTTABLES)+len(ds.DBRootViews)),
		progressbar.OptionSetDescription("Setup root DB"),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		progressbar.OptionSetMaxDetailRow(1),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)
	LoadCache(utils.ReservedParam, db)
	initializeTables(domainInstance, bar)     // Create tables if they don't exist, needed for the next step
	initializeRootTables(domainInstance, bar) // Create root tables if they don't exist, needed for the next step
	createSuperAdmin(domainInstance, bar)
	createRootView(domainInstance, bar)
}

func initializeTables(domainInstance utils.DomainITF, bar *progressbar.ProgressBar) {
	for _, table := range ds.NOAUTOLOADROOTTABLES {
		if _, err := GetSchema(table.Name); err != nil {
			bar.AddDetail("Creating table " + table.Name)
			domainInstance.CreateSuperCall(utils.GetTableTargetParameters(table.Name), table.ToSchemaRecord())
		}
		bar.Add(1)
	}
}

func initializeRootTables(domainInstance utils.DomainITF, bar *progressbar.ProgressBar) {
	var wfNew, viewNew bool
	rootTables := append(ds.ROOTTABLES, ds.DEMOROOTTABLES...)
	for _, table := range rootTables {
		if _, err := GetSchema(table.Name); err != nil {
			bar.AddDetail("Creating Schema " + table.Name)
			if createRootTable(domainInstance, table.ToRecord()) {
				wfNew = table.Name == ds.DBWorkflow.Name
				viewNew = table.Name == ds.DBView.Name
				if schema, err := GetSchema(table.Name); err == nil {
					if wfNew {
						createWorkflowView(domainInstance, schema, bar)
					}
					if viewNew {
						createView(domainInstance, schema, bar)
					}
				}
			}
		}
		bar.Add(1)
	}
}

func createRootTable(domainInstance utils.DomainITF, record utils.Record) bool {
	params := utils.AllParams(ds.DBSchema.Name).RootRaw()
	res, err := domainInstance.CreateSuperCall(params, record)
	return !(err != nil || len(res) == 0)
}

func createWorkflowView(domainInstance utils.DomainITF, schema sm.SchemaModel, bar *progressbar.ProgressBar) {
	bar.AddDetail("Creating Integration Workflow for Schema " + schema.Name)
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

func createRootView(domainInstance utils.DomainITF, bar *progressbar.ProgressBar) {
	for _, view := range ds.DBRootViews {
		bar.AddDetail("Creating Root View " + utils.ToString(utils.ToMap(view)[sm.NAMEKEY]))
		params := utils.AllParams(ds.DBView.Name).RootRaw()
		domainInstance.CreateSuperCall(params, view)
		bar.Add(1)
	}
}

func createView(domainInstance utils.DomainITF, schema sm.SchemaModel, bar *progressbar.ProgressBar) {
	bar.AddDetail("Create View for Schema " + schema.Name)
	params := utils.AllParams(ds.DBWorkflow.Name).RootRaw()
	newView := utils.Record{
		sm.NAMEKEY:       fmt.Sprintf("create %s", ds.DBView.Name),
		"description":    fmt.Sprintf("new %s workflow", ds.DBView.Name),
		ds.SchemaDBField: schema.ID,
	}
	domainInstance.CreateSuperCall(params, newView)
}

func createSuperAdmin(domainInstance utils.DomainITF, bar *progressbar.ProgressBar) {
	bar.AddDetail("Create SuperAdmin profile user ")
	domainInstance.CreateSuperCall(utils.AllParams(ds.DBUser.Name).RootRaw(), utils.Record{
		"name":        os.Getenv("SUPERADMIN_NAME"),
		"email":       os.Getenv("SUPERADMIN_EMAIL"),
		"super_admin": true,
		"password":    os.Getenv("SUPERADMIN_PASSWORD"),
	})
	bar.Add(1)
}

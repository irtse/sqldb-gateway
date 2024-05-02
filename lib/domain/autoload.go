package domain

import (
	"os"
	"fmt"
	"encoding/json"
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)

func Load() {
	database := conn.Open()
	defer database.Conn.Close()
	for _, table := range []schserv.SchemaModel{ schserv.DBSchema, schserv.DBSchemaField, schserv.DBPermission, schserv.DBView, schserv.DBWorkflow } {
			rec := map[string]interface{}{}
			data, _:= json.Marshal(table)
			json.Unmarshal(data, &rec)
			service := infrastructure.Table(database, true, "", table.Name, rec)
			service.NoLog = true
			service.CreateOrUpdate()
	}
	d := Domain(true, "superadmin", false)
	d.AutoLoad = true
	schserv.LoadCache(utils.ReservedParam, database)
	roots := schserv.ROOTTABLES
	roots = append(roots, schserv.DEMOROOTTABLES...)
	for _, t := range roots {
		rec := utils.Record{}
		data, _:= json.Marshal(t)
		json.Unmarshal(data, &rec) 
		_, err := schserv.GetSchema(t.Name)
		if err != nil {
			p := utils.AllParams(schserv.DBSchema.Name) 
			p[utils.RootRawView] = "enable"
			res, err := d.Call(p, rec, utils.CREATE)
			if err != nil && len(res) == 0 { continue }
			if t.Name == schserv.DBView.Name || t.Name == schserv.DBWorkflow.Name {
				addFields(t, d, schserv.SchemaModel{}.Deserialize(res[0]), false)
			}
			schserv.LoadCache(t.Name, database)
		}
	}
	for _, t := range roots {
		schema, err := schserv.GetSchema(t.Name)
		if err != nil { continue }
		if t.Name == schserv.DBWorkflow.Name {
			params := utils.Params{ utils.RootTableParam: schserv.DBView.Name, utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable" }
			newWF := utils.Record{ schserv.NAMEKEY : "workflow", 
				"indexable" : true, "description": "View description for " + t.Name + " datas.", "category" : "workflow", 
				"is_empty": false, "index": 0, "is_list": true, "readonly": false, schserv.RootID(schserv.DBSchema.Name) : schema.ID }
			d.Call(params, newWF, utils.CREATE)
			continue
		}
		if t.Name == schserv.DBView.Name {
			params := utils.Params{ utils.RootTableParam: schserv.DBWorkflow.Name, utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable" }
			newView := utils.Record{ schserv.NAMEKEY : "create " + t.Name, "description": "new " + t.Name + " workflow", schserv.RootID(schserv.DBSchema.Name) : schema.ID }
			d.Call(params, newView, utils.CREATE)
		}
		addFields(t, d, schema, true)
	}
	// Generate an root superadmin (ready to use...)
	p := utils.AllParams(schserv.DBUser.Name)
	p[utils.RootRawView] = "enable"
	d.Call(p, utils.Record{ "name" : os.Getenv("SUPERADMIN_NAME"), "email" : os.Getenv("SUPERADMIN_EMAIL"), "super_admin" : true, "password" : os.Getenv("SUPERADMIN_PASSWORD") }, utils.CREATE)
	addRootDatas(DBRootViews, schserv.DBView.Name)
	for name, datas := range schserv.DEMODATASENUM {
		for _, data := range datas {
			params := utils.Params{ utils.RootTableParam: name, utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable" }
			d.Call(params, utils.Record{ "name" : data }, utils.CREATE)
		}
	}
}
func addFields(t schserv.SchemaModel, d *MainService, schema schserv.SchemaModel, ok bool) {
	for _, col := range t.Fields {
		if schema.HasField(col.Name) && ok { continue }
		c := utils.Record{}
		b, _:= json.Marshal(col)
		json.Unmarshal(b, &c)
		if col.ForeignTable != "" {
			foreign, err := schserv.GetSchema(col.ForeignTable)
			if err == nil { c["link_id"]=foreign.ID }
		}
		c[schserv.RootID(schserv.DBSchema.Name)] = schema.ID
		d.AutoLoad = true
		d.Call(utils.AllParams(schserv.DBSchemaField.Name), c, utils.CREATE)
		
	}
}

func addRootDatas(flattenedSubArray []map[string]interface{}, name string) {
	d := Domain(true, "superadmin", false)
	for _, root := range flattenedSubArray {
		if _, ok := root["link"]; ok {
			schema, err := schserv.GetSchema(fmt.Sprintf("%v", root["link"]))
			if err != nil { continue }
			root[schserv.RootID(schserv.DBSchema.Name)] = schema.ID
			delete(root, "link")
			if filter, ok := root["filter"]; ok {
				params := utils.Params{ utils.RootTableParam: schserv.DBFilter.Name, utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable" }
				filter.(map[string]interface{})["link"] = schema.Name
				d.Call(params, filter.(map[string]interface{}), utils.CREATE)
			}
			params := utils.Params{ utils.RootTableParam: name, utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable" }
			d.Call(params, root, utils.CREATE)
		}
	}
}

var DBRootViews = []map[string]interface{}{ 
	map[string]interface{} { schserv.NAMEKEY : "submit a data",
	"is_list" : false,
	"indexable" : true,
	"description" : "select a form to submit an entry.",
	"readonly" : false,
	"index" : 0,
	"link" : schserv.DBRequest.Name,
	"is_empty" : true, 
	"category" : "request",
	"filter" : map[string]interface{}{
		"name" : "submit form",
		"view_fields" : []interface{}{
			map[string]interface{}{ "name" : "dbworkflow_id", "index" : 0 },
		},
	},
	},
	map[string]interface{} { schserv.NAMEKEY : "my unvalidated datas",
	"is_list" : true,
	"indexable" : true,
	"description" : nil,
	"readonly" : true,
	"index" : 1,
	"category" : "request",
	"link" : schserv.DBRequest.Name,
	"is_empty" : false,
	"filter" : map[string]interface{}{
		"name" : "unvalidated requests",
		"fields" : []interface{}{
			map[string]interface{}{ "name" : "state", "value" : "completed", "dir" : "ASC", "index" : 0, "operator" : "!=", "separator" : "and" },
		},
	}, },
	map[string]interface{} { schserv.NAMEKEY : "my validated datas",
	"is_list" : true,
	"indexable" : true,
	"description" : nil,
	"readonly" : true,
	"index" : 1,
	"link" : schserv.DBRequest.Name,
	"is_empty" : false,
	"filter" : map[string]interface{}{
		"name" : "validated requests",
		"fields" : []interface{}{
			map[string]interface{}{ "name" : "state", "value" : "completed", "dir" : "ASC", "index" : 0, "operator" : "=", "separator" : "and" },
		},
	}, },
	map[string]interface{} { schserv.NAMEKEY : "assigned activity",
	"is_list" : true,
	"indexable" : true,
	"description" : nil,
	"readonly" : false,
	"index" : 1,
	"link" : schserv.DBTask.Name,
	"category" : "my activity",
	"is_empty" : false,
	"filter" : map[string]interface{}{
		"name" : "unvalidated tasks",
		"fields" : []interface{}{
			map[string]interface{}{ "name" : "state", "value" : "completed", "dir" : "ASC", "index" : 0, "operator" : "!=", "separator" : "and" },
		},
	}, },
	map[string]interface{} { schserv.NAMEKEY : "archived activity",
	"is_list" : true,
	"indexable" : true,
	"description" : nil,
	"readonly" : true,
	"index" : 1,
	"category" : "my activity",
	"link" : schserv.DBTask.Name,
	"is_empty" : false,
	"filter" : map[string]interface{}{
		"name" : "validated tasks",
		"fields" : []interface{}{
			map[string]interface{}{ "name" : "state", "value" : "completed", "dir" : "ASC", "index" : 0, "operator" : "=", "separator" : "and" },
		},
	}},
}
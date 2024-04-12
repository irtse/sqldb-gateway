package domain

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)

func Load() {
	database := conn.Open()
	defer database.Conn.Close()
	for _, table := range []entities.TableEntity{ entities.DBSchema, entities.DBSchemaField, entities.DBPermission } {
			rec := tool.Record{}
			data, _:= json.Marshal(table)
			json.Unmarshal(data, &rec)
			if typ, ok := rec[entities.TYPEATTR]; ok && strings.Contains(strings.ToLower(fmt.Sprintf("%v", typ)), "many") {
				continue
			}
			service := infrastructure.Table(database, true, "", table.Name, tool.Params{}, rec, tool.CREATE)
			service.NoLog = true
			service.CreateOrUpdate()
	}
	d := Domain(true, "superadmin", false)
	for _, table := range entities.ROOTTABLES {
			rec := tool.Record{}
			data, _:= json.Marshal(table)
			json.Unmarshal(data, &rec) 
			res, err := d.SuperCall(tool.Params{ tool.RootTableParam: entities.DBSchema.Name,
									tool.RootRowsParam: tool.ReservedParam, }, 
									rec, tool.CREATE, "CreateOrUpdate")
			
			if err != nil || len(res) == 0 { continue }
			for _, col := range rec["columns"].([]interface{}) {
				c := col.(map[string]interface{})
				c[entities.RootID(entities.DBSchema.Name)] = res[0][tool.SpecialIDParam]
				d.SuperCall(tool.Params{ tool.RootTableParam: entities.DBSchemaField.Name,
										 		   tool.RootRowsParam: tool.ReservedParam, }, 
									  c, tool.CREATE, "CreateOrUpdate")
			}
	}
	// Generate an root superadmin (ready to use...)
	found, err := d.SuperCall(tool.Params{ 
		tool.RootTableParam: entities.DBUser.Name,
		tool.RootRowsParam: tool.ReservedParam, }, 
		tool.Record{ }, tool.SELECT, "Get", "name='root'")
	if err != nil || len(found) == 0 {
		d.SuperCall(tool.Params{ 
			tool.RootTableParam: entities.DBUser.Name,
			tool.RootRowsParam: tool.ReservedParam, }, 
			tool.Record{
				"name" : "root",
				"email" : "admin@super.com",
				"super_admin" : true, // oh well think about "backin to the future"
				"password" : "$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c",
			}, tool.CREATE, "CreateOrUpdate")
	}
	addRootDatas(entities.DBRootViews, entities.DBView.Name)
	addRootDatas(entities.DBRootWorkflows, entities.DBWorkflow.Name)
	database.Conn.Close()
}
func addRootDatas(flattenedSubArray []map[string]interface{}, name string) {
	d := Domain(true, "superadmin", false)
	for _, root := range flattenedSubArray {
		if _, ok := root["link"]; ok {
			params := tool.Params{ tool.RootTableParam: entities.DBSchema.Name, tool.RootRowsParam: tool.ReservedParam, tool.RootRawView: "enable" }
			schemas, err := d.SuperCall(params, tool.Record{}, tool.SELECT, "Get", "name='" + fmt.Sprintf("%v", root["link"]) + "'")
			if err != nil || len(schemas) == 0 { continue }
			root[entities.RootID(entities.DBSchema.Name)] = schemas[0][tool.SpecialIDParam]
			delete(root, "link")
			params = tool.Params{ tool.RootTableParam: name, tool.RootRowsParam: tool.ReservedParam, tool.RootRawView: "enable" }
			d.SuperCall(params, root, tool.CREATE, "CreateOrUpdate")
		}
	}
}
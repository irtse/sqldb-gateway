package domain

import (
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)

func Load() {
	database := conn.Open()
	defer database.Conn.Close()
	tables := [][]entities.TableEntity{ entities.DBRESTRICTED, entities.ROOTTABLES }
	for _, t := range tables {
		for _, table := range t {
			rec := tool.Record{}
			data, _:= json.Marshal(table)
			json.Unmarshal(data, &rec)
			service := infrastructure.Table(database, true, "", table.Name, tool.Params{}, rec, tool.CREATE)
			service.NoLog = true
			service.CreateOrUpdate()
		}
	}
	d := Domain(true, "superadmin", false)
	for _, t := range tables {
		for _, table := range t {
			rec := tool.Record{}
			data, _:= json.Marshal(table)
			json.Unmarshal(data, &rec)
			res, err := d.SuperCall(tool.Params{ 
					tool.RootTableParam: entities.DBSchema.Name,
					tool.RootRowsParam: tool.ReservedParam,
					entities.NAMEATTR : table.Name }, tool.Record{}, tool.SELECT, "Get")
			if err != nil || len(res) == 0 { 
				res, err = d.SuperCall(tool.Params{ tool.RootTableParam: entities.DBSchema.Name,
										tool.RootRowsParam: tool.ReservedParam, }, 
							rec, tool.CREATE, "CreateOrUpdate")
				if err != nil || len(res) == 0 { continue }
				for _, col := range rec["columns"].([]interface{}) {
					c := col.(map[string]interface{})
					c[entities.RootID(entities.DBSchema.Name)] = res[0][tool.SpecialIDParam]
					d.SuperCall(tool.Params{ tool.RootTableParam: entities.DBSchemaField.Name,
											 tool.RootRowsParam: tool.ReservedParam, }, c, tool.CREATE, "CreateOrUpdate")
				}
			}
		}
	}
	// Generate an root superadmin (ready to use...)
	found, err := d.SuperCall(tool.Params{ 
		tool.RootTableParam: entities.DBUser.Name,
		tool.RootRowsParam: tool.ReservedParam,
	"login" : "root" }, 
	tool.Record{ }, tool.SELECT, "Get")
	if err == nil && found != nil {
		d.SuperCall(tool.Params{ 
			tool.RootTableParam: entities.DBUser.Name,
			tool.RootRowsParam: tool.ReservedParam, }, 
			tool.Record{
				"login" : "root",
				"super_admin" : true, // oh well think about "backin to the future"
				"password" : "$argon2id$v=19$m=65536,t=3,p=4$JooiEtVXatRxSz16N9uo2g$Y2dAHdLAK06013FhDHQ/xhd+UL2yInwDAvRS1+KKD3c",
			}, tool.CREATE, "CreateOrUpdate")
	}
	database.Conn.Close()
}
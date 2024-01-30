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
	database.Conn.Close()
}
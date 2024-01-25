package service

import (
	"strings"
	"slices"
	tool "sqldb-ws/lib"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// union PG + MYSQL
func ToFilter(tableName string, params tool.Params, db *conn.Db) *conn.Db {
	db = ClearFilter(db)
    for key, element := range params {
		col := &TableColumnInfo{ }
		col.db = db
		col.Name = tableName
		typ, ok := col.Verify(key)
		if ok { 
			if strings.Contains(element, ",") { 
				els := ""
				for _, el := range strings.Split(element, ",") { els += conn.FormatForSQL(typ, el) + "," }
				db.SQLRestriction += key + " IN (" + conn.RemoveLastChar(els) + ") AND " 
			} else { db.SQLRestriction += key + "=" + conn.FormatForSQL(typ, element) + " AND " }
		}
	}
	if len(db.SQLRestriction) > 4 { db.SQLRestriction = db.SQLRestriction[0:len(db.SQLRestriction) - 4]}
	db.SQLView = ""
	alreadySet := []string{}
	for key, element := range params { // preload restriction
		if key == tool.RootRowsParam || key == tool.RootTableParam || slices.Contains(alreadySet, key) { continue }
		if key == tool.RootShallow && element == "enable" {
			/*table := &TableInfo{ }
			table.db = db
			schemas, err := table.schema(tableName)
			if err != nil || len(schemas) == 0 { continue }
			for key, v := range schemas[0].AssColumns {
				if !v.Null { 
					if len(db.SQLView) == 0 || !strings.Contains(db.SQLView, key) { 
						db.SQLView += key + "," 
					}
				}
			}*/
			continue 
		} 
		if key == tool.RootColumnsParam { 
			for _, el := range strings.Split(element, ",") {
				if len(db.SQLView) == 0 || !strings.Contains(db.SQLView, key) {
					db.SQLView += el + ","
				}
			}
			continue 
		}
		db.SQLView = conn.RemoveLastChar(db.SQLView)
		dir, ok := params[tool.RootDirParam]
		if key == tool.RootOrderParam && ok {
			direction := strings.Split(dir, ",")
			for i, el := range strings.Split(element, ",") {
                if len(direction) > i { db.SQLOrder += el + " " + direction[i] + " "
				} else { db.SQLOrder += el + " ASC" }
			}
			continue
		}
		if key == tool.RootSQLFilterParam {
			if len(db.SQLRestriction) == 0 { db.SQLRestriction = params[tool.RootSQLFilterParam]
			} else { db.SQLRestriction += " AND " + params[tool.RootSQLFilterParam] }
			continue
		}
		alreadySet = append(alreadySet, key)
	}
	return db
}

func ClearFilter(db *conn.Db) *conn.Db {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
	return db
}

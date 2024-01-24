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
				db.SQLRestriction += key + " IN (" + conn.RemoveLastChar(els) + ") " 
			} else { db.SQLRestriction += key + "=" + conn.FormatForSQL(typ, element) + " " }
		}
	}
	db.SQLView = ""
	alreadySet := []string{}
	for key, element := range params { // preload restriction
		if key == tool.RootRowsParam || key == tool.RootTableParam || slices.Contains(alreadySet, key) { continue }
		if key == tool.RootColumnsParam { db.SQLView += element; continue }
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
			db.SQLRestriction += params[tool.RootSQLFilterParam]
			continue
		}
		alreadySet = append(alreadySet, key)
	}
	    // TODO IN 
	return db
}

func ClearFilter(db *conn.Db) *conn.Db {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
	return db
}

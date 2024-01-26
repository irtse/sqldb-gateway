package service

import (
	"strings"
	"slices"
	tool "sqldb-ws/lib"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// union PG + MYSQL
func ToFilter(tableName string, params tool.Params, db *conn.Db, permsCols string) *conn.Db {
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
		if key == tool.RootColumnsParam { 
			for _, el := range strings.Split(element, ",") {
				if (len(db.SQLView) == 0 || !strings.Contains(db.SQLView, el)) && !strings.Contains(permsCols, el) {
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
		if key == tool.RootSQLFilterParam  && params[tool.RootSQLFilterParam] != "" {
			if len(db.SQLRestriction) == 0 { db.SQLRestriction = params[tool.RootSQLFilterParam]
			} else { db.SQLRestriction += " AND " + params[tool.RootSQLFilterParam] }
			continue
		}
		alreadySet = append(alreadySet, key)
	}
	if len(permsCols) > 0 {
		for _, el := range strings.Split(permsCols, ",") {
			if len(db.SQLView) == 0 || !strings.Contains(db.SQLView, el) {
				db.SQLView += el + ","
			}
		}
	}
	return db
}

func ClearFilter(db *conn.Db) *conn.Db {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
	return db
}

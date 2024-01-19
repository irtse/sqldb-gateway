package connector

import (
	"strings"
	"slices"
	tool "sqldb-ws/lib"
)
// union PG + MYSQL
func ToFilter(tableName string, params tool.Params, db *Db) *Db{
	ClearFilter(db)
	count := 0
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
			db.SQLRestriction += params[tool.RootSQLFilterParam] + " "
			count += 1
			continue
		}
		alreadySet = append(alreadySet, key)
	}
	    // TODO IN 
	return db
}

func ClearFilter(db *Db) *Db {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
	return db
}

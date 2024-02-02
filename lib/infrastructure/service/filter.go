package service

import (
	"fmt"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// union PG + MYSQL
func ToFilter(tableName string, params tool.Params, db *conn.Db, permsCols string, restriction... string) *conn.Db {
	db = ClearFilter(db)
	fields := prepareRowField(tableName, db) 	// retrive all fields from schema...
	db = ClearFilter(db)
    for key, element := range params {
		field, ok := fields[key]
		if ok { 
			if strings.Contains(element, ",") { 
				els := ""
				for _, el := range strings.Split(element, ",") { 
					els += conn.FormatForSQL(field.Type, SQLInjectionProtector(el)) + "," 
				}
				db.SQLRestriction += key + " IN (" + conn.RemoveLastChar(els) + ") AND " 
			} else { 
				sql := conn.FormatForSQL(field.Type, element)
				if len(sql) > 1 && sql[:2] == "'%" && sql[:len(sql) - 2] == "%'" {
					db.SQLRestriction += key + " LIKE " + sql + " AND "
				} else { db.SQLRestriction += key + "=" + sql + " AND " }
				db.SQLRestriction += key + "=" + sql + " AND " 
			}
		}
	}
	if len(db.SQLRestriction) > 4 { db.SQLRestriction = db.SQLRestriction[0:len(db.SQLRestriction) - 4]}
	for _, restr := range restriction {
		if len(db.SQLRestriction) == 0 { db.SQLRestriction = restr
		} else { db.SQLRestriction += " AND " + restr }
		continue
	}

	db.SQLView = ""
	if element, valid := params[tool.RootColumnsParam]; valid { 
		for _, el := range strings.Split(element, ",") {
			if (len(db.SQLView) == 0 || !strings.Contains(db.SQLView, el)) && !strings.Contains(permsCols, el) {
				db.SQLView += SQLInjectionProtector(el) + ","
			}
		}
	}
	if element, valid := params[tool.RootOrderParam]; valid {
		dir, ok := params[tool.RootDirParam]
		if !ok { dir = "" }
		direction := strings.Split(dir, ",")
		for i, el := range strings.Split(element, ",") {
			if len(direction) > i { 
				upper := strings.Replace(strings.ToUpper(dir), " ", "", -1)
				if upper == "ASC" || upper == "DESC" { db.SQLOrder += SQLInjectionProtector(el + " " + upper + ",") }
			} 
			db.SQLOrder += SQLInjectionProtector(el + " ASC,") 
		}
		db.SQLOrder = conn.RemoveLastChar(db.SQLOrder)
	}
	if len(permsCols) > 0 {
		for _, el := range strings.Split(permsCols, ",") {
			if len(db.SQLView) == 0 || !strings.Contains(db.SQLView, el) {
				db.SQLView += el + ","
			}
		}
	}
	db.SQLView = conn.RemoveLastChar(db.SQLView)
	return db
}

func ClearFilter(db *conn.Db) *conn.Db {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
	return db
}

func SQLInjectionProtector(injection string) (string) {
	quoteCounter := strings.Count(injection, "'")
	quoteCounter2 := strings.Count(injection, "\"")
	if (quoteCounter % 2) != 0 || (quoteCounter2 % 2) != 0 {
		log.Error().Msg("injection alert: strange activity of quoting founded")
		return ""
	}
	notAllowedChar := []string{ "«", "#", "union", ";", ")", "%27", "%22", "%23", "%3B", "%29" }
	for _, char := range notAllowedChar {
		if strings.Contains(strings.ToLower(injection), char) {
			log.Error().Msg( "injection alert: not allowed " + char + " filter" )
			return ""
		}
	}
	return injection
}

func prepareRowField(tableName string, db *conn.Db) map[string]entities.SchemaColumnEntity {
	db.SQLRestriction = entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
	db.SQLRestriction += entities.DBSchema.Name + " WHERE name=" + conn.Quote(tableName) + ")"
	res, err := db.SelectResults(entities.DBSchemaField.Name)
	if err != nil && len(res) == 0 { return map[string]entities.SchemaColumnEntity {} }
	fields := map[string]entities.SchemaColumnEntity{}
	for _, rec := range res {
		var scheme entities.SchemaColumnEntity
		b, _ := json.Marshal(rec)
		json.Unmarshal(b, &scheme)
		fields[scheme.Name]=scheme
	}
	return fields
}

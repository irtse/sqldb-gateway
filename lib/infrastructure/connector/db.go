package connector

import (
	"os"
	"fmt"
	"slices"
	"errors"
	"strings"
	"reflect"
	"strconv"
	"encoding/json"
	"database/sql"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	"github.com/rs/zerolog"
	_ "github.com/go-sql-driver/mysql"
)

var log zerolog.Logger
/*
	Generic Connector to DB 
*/
const PostgresDriver = "postgres" 
const MySQLDriver = "mysql"
var drivers = []string{PostgresDriver, MySQLDriver} // define all drivers available per adapter

// verify driver is available
func checkDriver(d string) bool {
	for _, driver := range drivers {
		if driver == d { return true }
	}
	return false
}

type Db struct {
	Driver         string
	Url            string
	SQLView        string 
	SQLOrder       string 
	SQLDir		   string
	SQLRestriction string 	  	  
	LogQueries     bool
	Restricted     bool
	Conn           *sql.DB
}
// Open the database
func Open() *Db {
	var database Db
	var err error
	database.Driver = os.Getenv("driverdb")
	if checkDriver(database.Driver) == true { 
		database.Url = "host=" + os.Getenv("dbhost") + " port=" + os.Getenv("dbport")
		database.Url += " user=" + os.Getenv("dbuser") + " password=" + os.Getenv("dbpwd")
		database.Url += " dbname=" + os.Getenv("dbname") + " sslmode=" + os.Getenv("dbssl")
		database.Conn, err = sql.Open(database.Driver, database.Url)
		if err != nil { log.Error().Msg(err.Error()) }
		return &database
	} else { log.Error().Msg("Not valid DB driver !") }
	return &database
}

func (db *Db) GetSQLRestriction() string {
	return db.SQLRestriction
}
func (db *Db) GetSQLOrder() string {
	return db.SQLOrder
}
func (db *Db) GetSQLView() string {
	return db.SQLView
}

func (db *Db) Prepare(query string) (*sql.Stmt, error) {
	if db.LogQueries { log.Info().Msg(query) }
	// fmt.Printf("QUERY : %s\n", query)
	stmt, err := db.Conn.Prepare(query)
	if err != nil { return nil, err }
	return stmt, nil
}

func (db *Db) QueryRow(query string) (int64, error) {
    id := 0
	if db.LogQueries { log.Info().Msg(query) }
	err := db.Conn.QueryRow(query + " RETURNING id").Scan(&id)
	// fmt.Printf("QUERY : %s %v \n", query, err)
	if err != nil { return int64(id), err }
	return int64(id), err
}

func (db *Db) Query(query string) (error) {
	// if db.LogQueries { log.Info().Msg(query) }
	//if strings.Contains(query, "UPDATE") { fmt.Printf("QUERY : %s\n", query) }
	rows, err := db.Conn.Query(query)
	if err != nil { return err }
	err = rows.Close()
	return err
}

func (db *Db) QueryAssociativeArray(query string) (tool.Results, error) {
    if strings.Contains(query, "<nil>") { return nil, errors.New("not found")}
	// if strings.Contains(query, "dbtask") { fmt.Printf("QUERY : %s\n", query) }
	rows, err := db.Conn.Query(query)
	if err != nil { 
		// fmt.Printf("QUERY : %s\n", query)
		return nil, err 
	}
	defer rows.Close()
	// get rows
	results := tool.Results{}
	cols, err := rows.Columns()
	if err != nil { return nil, err }
	// make types map
	columnTypes, err := rows.ColumnTypes()
	if err != nil { return nil, err }
	columnType := make(map[string]string)
	for _, colType := range columnTypes {
		columnType[colType.Name()] = strings.ToUpper(colType.DatabaseTypeName())
	}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns { columnPointers[i] = &columns[i] }

		// Scan the result into the column pointers...
		err = rows.Scan(columnPointers...)
		if err != nil { return nil, err }

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(tool.Record)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if columnType[colName] != "" {
				if db.Driver == MySQLDriver {
					if (*val) == nil { m[colName] = nil
					} else { m[colName] = ValueByType(columnType[colName], *val, query) }
				}
				if db.Driver == PostgresDriver { 
					m[colName] = *val 
					if m[colName] != nil { 
						if (strings.Contains("DOUBLE", columnType[colName]) || strings.Contains("MONEY", columnType[colName]) || strings.Contains("FLOAT", columnType[colName]) || strings.Contains("NUMERIC", columnType[colName]) || strings.Contains("DECIMAL", columnType[colName])) {
							if strings.Contains("MONEY", columnType[colName]) {
								m[colName], _ = strconv.ParseFloat(string(m[colName].([]uint8))[1:], 64)
							} else { m[colName], _ = strconv.ParseFloat(string(m[colName].([]uint8)), 64) }
						}
						if strings.Contains("TIMESTAMP", columnType[colName]) || strings.Contains("DATE", columnType[colName]) {
							m[colName] = fmt.Sprintf("%v", m[colName])[:10]
						}
					}
				}
			} else {
				m[colName] = *val
				if m[colName] != nil { 
					m[colName] = fmt.Sprintf("%v", string(m[colName].([]uint8)))
				}
			}
		}
		results = append(results, m)
	}
	return results, nil
}

func (db *Db) SelectResults(name string) (tool.Results, error) {
	return db.QueryAssociativeArray(db.BuildSelect(name))
}

func (db *Db) BuildSelect(name string, view... string) string {
	var query string
	if db.SQLView == "" { 
		if len(view) > 0 {
			viewStr := ""
			for _, v := range view { viewStr += v + "," }
			query = "SELECT " + viewStr[:len(viewStr) - 1] + " FROM " + name
		} else { query = "SELECT * FROM " + name }
	} else { query = "SELECT " + db.SQLView + " FROM " + name }
	if db.SQLRestriction != "" { query += " WHERE " + db.SQLRestriction }
	if db.SQLOrder != "" { 
		query += " ORDER BY " + db.SQLOrder 
	}
	return query
}

func ValueByType(typing string, defaulting interface{}, query string) interface{} {
	switch typing {
	case "INT", "BIGINT":
		val, err := strconv.ParseInt(fmt.Sprintf("%s", defaulting), 10, 64)
		if err != nil { return err }
		return val
	case "UNSIGNED BIGINT", "UNSIGNED INT":
		val, err := strconv.ParseUint(fmt.Sprintf("%s", defaulting), 10, 64)
		if err != nil { return err }
		return val
	case "FLOAT":
		val, err := strconv.ParseFloat(fmt.Sprintf("%s", defaulting), 64)
		if err != nil { return err }
		return val
	case "TINYINT":
		val, err := strconv.ParseInt(fmt.Sprintf("%s", defaulting), 10, 64)
		if err != nil { return err }
		if val == 1 { return true }
		return false
	case "VARCHAR", "TEXT", "TIMESTAMP", "VARBINARY":
		return fmt.Sprintf("%s", defaulting)
	default:
		if reflect.ValueOf(defaulting).IsNil() == false {
			return fmt.Sprintf("%v", defaulting)
		}
		return nil
	}
	return nil
}

// HELPING TOOLS FOR DB WORKS
var SpecialTypes = []string{"char", "text", "date", "time", "interval", "var", "blob", "set", "enum", "year", "USER-DEFINED"}

func Quote(s string) string { return "'" + s + "'" }

func RemoveLastChar(s string) string {
	r := []rune(s)
	if len(r) > 0 {return string(r[:len(r)-1])}
	return string(r)
}
// transition for mysql types
func FormatForSQL(datatype string, value interface{}) string {
	if value == nil { return "NULL" }
	strval := fmt.Sprintf("%v", value)
	if len(strval) == 0 { return "NULL" }
	for _, typ := range SpecialTypes {
		if strings.Contains(datatype, typ) { 
			if value == "CURRENT_TIMESTAMP" { return fmt.Sprint(value) 
			} else { return Quote(fmt.Sprint(value))}
		}
	}
	return fmt.Sprint(strval)
}
func (db *Db) ToFilter(tableName string, params map[string]string, restriction... string) {
	db.ClearFilter()
	fields := db.prepareRowField(tableName)
	db.ClearFilter()
	already := []string{}
    for key, element := range params {
		field, ok := fields[key]
		if (ok || key == "id") && !slices.Contains(already, key){ 
			already = append(already, key)
			// TODO if AND or OR define truncature
			if strings.Contains(element, ",") { 
				els := ""
				for _, el := range strings.Split(element, ",") { 
					els += FormatForSQL(field.Type, SQLInjectionProtector(el)) + "," 
				}
				if len(db.SQLRestriction) > 0 { 
					db.SQLRestriction +=  " AND (" 
					db.SQLRestriction += key + " IN (" + RemoveLastChar(els) + ")"
					db.SQLRestriction +=  ")"
				} else { db.SQLRestriction += key + " IN (" + RemoveLastChar(els) + ")" }
			} else { 
				sql := FormatForSQL(field.Type, element)
				if len(sql) > 1 && sql[:2] == Quote("%" + sql[:len(sql) - 2] + "%") {
					if len(db.SQLRestriction) > 0 { 
						db.SQLRestriction +=  " AND (" 
						db.SQLRestriction += key + " LIKE " + sql
						db.SQLRestriction +=  ")"
					} else { db.SQLRestriction += key + " LIKE " + sql }
				} else { 
					if len(db.SQLRestriction) > 0 { 
						db.SQLRestriction +=  " AND (" 
						db.SQLRestriction += key + "=" + sql
						db.SQLRestriction +=  ")"
					} else { db.SQLRestriction += key + "=" + sql }
				}
			}
		}
	}
	for _, restr := range restriction {
		if len(restr) > 0 && !strings.Contains(db.SQLRestriction, restr) {
			if len(db.SQLRestriction) == 0 { db.SQLRestriction = restr
			} else { db.SQLRestriction += " AND (" + restr + ")" }
		}
		continue
	}
	if orderBy, ok := params["order_by"]; ok {
		direction := []string{}
		if dir, ok2 := params["dir"]; !ok2 { 
			direction = strings.Split(fmt.Sprintf("%v", dir), ",")
		}
		for i, el := range strings.Split(fmt.Sprintf("%v", orderBy), ",") {
			if len(direction) > i { 
				upper := strings.Replace(strings.ToUpper(direction[i]), " ", "", -1)
				if upper == "ASC" || upper == "DESC" { db.SQLOrder += SQLInjectionProtector(el + " " + upper + ",") }
			} 
			db.SQLOrder += SQLInjectionProtector(el + " ASC,") 
		}
		db.SQLOrder = RemoveLastChar(db.SQLOrder)
	}
}

func (db *Db) ClearFilter() {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
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

func (db *Db) prepareRowField(tableName string) map[string]entities.SchemaColumnEntity {
	db.SQLRestriction = entities.RootID(entities.DBSchema.Name) + " IN (SELECT id FROM "
	db.SQLRestriction += entities.DBSchema.Name + " WHERE name=" + Quote(tableName) + ")"
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
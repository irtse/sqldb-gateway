package connector

import (
	"os"
	"fmt"
	"sync"
	"strings"
	"strconv"
	"database/sql"
	_ "github.com/lib/pq"
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
	SQLLimit	   string
	SQLRestriction string 	  	  
	LogQueries     bool
	Conn           *sql.DB
}
var mutex = sync.RWMutex{}
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
		if err != nil { 
			fmt.Println("Error opening database: ", err)
			log.Error().Msg(err.Error()) }
		return &database
	} else { log.Error().Msg("Not valid DB driver !") }
	return &database
}

func (db *Db) Close() {
	db.Conn.Close()
	db.Conn = nil
}

func (db *Db) GetSQLRestriction() string { return db.SQLRestriction }
func (db *Db) GetSQLOrder() string { return db.SQLOrder }
func (db *Db) GetSQLView() string { return db.SQLView }

func (db *Db) Prepare(query string) (*sql.Stmt, error) {
	if db.LogQueries { log.Info().Msg(query) }
	stmt, err := db.Conn.Prepare(query)
	if err != nil { return nil, err }
	return stmt, nil
}

func (db *Db) QueryRow(query string) (int64, error) {
    id := 0
	if db.LogQueries { log.Info().Msg(query) }
	mutex.Lock()
	err := db.Conn.QueryRow(query).Scan(&id)
	mutex.Unlock()
	if err != nil { return int64(id), err }
	return int64(id), err
}

func (db *Db) Query(query string) (error) {
	mutex.Lock()
	rows, err := db.Conn.Query(query)
	mutex.Unlock()
	if err != nil { return err }
	err = rows.Close()
	return err
}

func (db *Db) QueryAssociativeArray(query string) ([]map[string]interface{}, error) {
	if db.Conn == nil || strings.Contains(query, "<nil>") { return []map[string]interface{}{}, nil }
	rows, err := db.Conn.Query(query)
	if err != nil { return nil, err }
	defer rows.Close()
	// get rows
	results := []map[string]interface{}{}
	cols, err := rows.Columns()
	if err != nil { return nil, err }
	// make types map
	columnTypes, err := rows.ColumnTypes()
	if err != nil { return nil, err }
	columnType := make(map[string]string)
	for _, colType := range columnTypes { columnType[colType.Name()] = strings.ToUpper(colType.DatabaseTypeName()) }
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
		m := make(map[string]interface{})
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if columnType[colName] == "" { 
				m[colName] = *val
				if m[colName] != nil { m[colName] = fmt.Sprintf("%v", string(m[colName].([]uint8))) }
				continue 
			}
			if (*val) == nil { m[colName] = nil; continue }
			switch columnType[colName] {
				case "MONEY": m[colName], _ = strconv.ParseFloat(string(m[colName].([]uint8))[1:], 64); break
				case "DOUBLE", "FLOAT", "NUMERIC", "DECIMAL": m[colName], _ = strconv.ParseFloat(string(m[colName].([]uint8)), 64); break
				case "TIMESTAMP", "DATE": 
					if len(fmt.Sprintf("%v", m[colName])) > 10 { m[colName] = fmt.Sprintf("%v", m[colName])[:10]
					} else { m[colName] = fmt.Sprintf("%v", m[colName]) }
					
					break
				default: m[colName] = *val; break
			}
		}
		results = append(results, m)
	}
	return results, nil
}

func (db *Db) SelectResults(name string) ([]map[string]interface{}, error) { return db.QueryAssociativeArray(db.BuildSelect(name)) }

func (db *Db) BuildCount(name string) string {
	query := "SELECT COUNT(*) FROM " + name
	if db.SQLRestriction != "" { query += " WHERE " + db.SQLRestriction }
	return query
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
	if db.SQLOrder != "" { query += " ORDER BY " + db.SQLOrder }
	if db.SQLLimit != "" { query += " " + db.SQLLimit }
	return query
}

func (db *Db) ClearFilter() {
    db.SQLOrder=""
	db.SQLRestriction=""
	db.SQLView=""
}
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
			} else { return strings.Replace(Quote(strings.Replace(fmt.Sprint(value), "'", "''", -1)), "%25", "%", -1) }
		}
	}
	if strings.Contains(strval, "%") { return strings.Replace(Quote(strings.Replace(fmt.Sprint(value), "'", "''", -1)), "%25", "%", -1) }
	return strval
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
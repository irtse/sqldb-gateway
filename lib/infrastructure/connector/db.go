package connector

import (
	"os"
	"fmt"
	"strings"
	"reflect"
	"strconv"
	"database/sql"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog"
	_ "github.com/go-sql-driver/mysql"
)

var log zerolog.Logger

const PostgresDriver = "postgres" 
const MySQLDriver = "mysql"
var drivers = []string{PostgresDriver, MySQLDriver}

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
	// fmt.Printf("QUERY : %s\n", query)
	err := db.Conn.QueryRow(query + " RETURNING id").Scan(&id)
	if err != nil { return int64(id), err
	}
	return int64(id), err
}

func (db *Db) Query(query string) (error) {
	if db.LogQueries { log.Info().Msg(query) }
	// fmt.Printf("QUERY : %s\n", query)
	rows, err := db.Conn.Query(query)
	if err != nil { return err }
	err = rows.Close()
	return err
}

func (db *Db) QueryAssociativeArray(query string) (tool.Results, error) {
	// fmt.Printf("QUERY : %s\n", query)
	rows, err := db.Conn.Query(query)

	if err != nil { 
		fmt.Printf("QUERY : %s\n", query)
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
		columnType[colType.Name()] = colType.DatabaseTypeName()
	}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		// Scan the result into the column pointers...
		err = rows.Scan(columnPointers...)
		if err != nil { return nil, err }

		// Create our map, and retrieve the value for each column from the pointers slice,
		// storing it in the map with the name of the column as the key.
		m := make(tool.Record)
		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			if db.Driver == MySQLDriver {
				if (*val) == nil { m[colName] = nil
				} else { m[colName] = ValueByType(columnType[colName], *val, query) }
			}
			if db.Driver == PostgresDriver { m[colName] = *val }
		}
		results = append(results, m)
	}
	if len(results) == 0 {
		// fmt.Printf("QUERY : %s\n", query)
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
			for _, v := range view {
				viewStr += v + ","
			}
			query = "SELECT " + viewStr[:len(viewStr) - 1] + " FROM " + name
		} else { query = "SELECT * FROM " + name }
	} else { query = "SELECT " + db.SQLView + " FROM " + name }
	if db.SQLRestriction != "" { query += " WHERE " + db.SQLRestriction }
	if db.SQLOrder != "" { query += " ORDER BY " + db.SQLOrder }
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
			// fmt.Printf("Unknow type : %s (%s)\n", typing, query)
			return fmt.Sprintf("%v", defaulting)
		}
		return nil
	}
	return nil
}

var SpecialTypes = []string{"char", "text", "date", "time", "interval", "var", "blob", "set", "enum", "year"}

func Quote(s string) string { return "'" + s + "'" }

func RemoveLastChar(s string) string {
	r := []rune(s)
	if len(r) > 0 {return string(r[:len(r)-1])}
	return string(r)
}

func FormatForSQL(datatype string, value interface{}) string {
	if value == nil { return "NULL" }
	strval := fmt.Sprintf("%v", value)
	if len(strval) == 0 { return "NULL" }
	for _, typ := range SpecialTypes {
		if strings.Contains(datatype, typ) { 
			if value == "CURRENT_TIMESTAMP" { return fmt.Sprint(value) 
			} else { return "'" + fmt.Sprint(value) + "'" }
		}
	}
	return fmt.Sprint(strval)
}
package connector

import (
	"database/sql"
	"fmt"
	"os"
	"slices"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

/*
Generic Connector to DB
*/
const PostgresDriver = "postgres"
const MySQLDriver = "mysql"

var (
	log     zerolog.Logger
	drivers = []string{
		PostgresDriver,
		MySQLDriver,
	}
)

type Database struct {
	Driver         string
	Url            string
	SQLView        string
	SQLOrder       string
	SQLDir         string
	SQLLimit       string
	SQLRestriction string
	LogQueries     bool
	Conn           *sql.DB
}

func (d *Database) GetDriver() string {
	return d.Driver
}

func (d *Database) GetConn() *sql.DB {
	return d.Conn
}

func (d *Database) GetSQLView() string {
	return d.SQLView
}

func (d *Database) GetSQLOrder() string {
	return d.SQLOrder
}

func (d *Database) GetSQLDir() string {
	return d.SQLDir
}

func (d *Database) GetSQLLimit() string {
	return d.SQLLimit
}

func (d *Database) GetSQLRestriction() string {
	return d.SQLRestriction
}

func (d *Database) SetSQLView(s string) {
	d.SQLView = s
}

func (d *Database) SetSQLOrder(s string) {
	d.SQLOrder = s
}

func (d *Database) SetSQLLimit(s string) {
	d.SQLLimit = s
}

func (d *Database) SetSQLRestriction(s string) {
	d.SQLRestriction = s
}

func Open(beforeDB *Database) *Database {
	if beforeDB != nil {
		beforeDB.Close()
	}
	db := &Database{Driver: os.Getenv("DBDRIVER")}
	if !slices.Contains(drivers, db.Driver) {
		log.Error().Msg("Invalid DB driver!")
		return nil
	}

	db.Url = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DBHOST"),
		os.Getenv("DBPORT"),
		os.Getenv("DBUSER"),
		os.Getenv("DBPWD"),
		os.Getenv("DBNAME"),
		os.Getenv("DBSSL"),
	)

	var err error
	db.Conn, err = sql.Open(db.Driver, db.Url)
	if err != nil {
		log.Error().Msgf("Error opening database: %v", err)
		return nil
	}
	return db
}

func (db *Database) Close() {
	if db.Conn != nil {
		db.Conn.Close()
		db.Conn = nil
	}
}

func (db *Database) ClearQueryFilter() *Database {
	db.SQLOrder = ""
	db.SQLRestriction = ""
	db.SQLView = ""
	return db
}

type DB interface {
	GetConn() *sql.DB
	GetSQLView() string
	GetSQLOrder() string
	GetSQLDir() string
	GetSQLLimit() string
	GetSQLRestriction() string
	SetSQLView(s string)
	SetSQLOrder(s string)
	SetSQLLimit(s string)
	SetSQLRestriction(s string)
	Close()
	ClearQueryFilter() *Database
	DeleteQueryWithRestriction(name string, restrictions map[string]interface{}, isOr bool) error
	SelectQueryWithRestriction(name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error)
	SimpleMathQuery(algo string, name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error)
	MathQuery(algo string, name string, naming ...string) ([]map[string]interface{}, error)
	SchemaQuery(name string) ([]map[string]interface{}, error)
	ListTableQuery() ([]map[string]interface{}, error)
	CreateTableQuery(name string) error
	UpdateQuery(name string, record map[string]interface{}, restriction map[string]interface{}, isOr bool) error
	DeleteQuery(name string, colName string) error
	BuildDeleteQueryWithRestriction(name string, restrictions map[string]interface{}, isOr bool) string
	BuildSimpleMathQueryWithRestriction(algo string, name string, restrictions interface{}, isOr bool, restr ...string) string
	BuildSelectQueryWithRestriction(name string, restrictions interface{}, isOr bool, view ...string) string
	BuildMathQuery(algo string, name string, naming ...string) string
	BuildDeleteQuery(tableName string, colName string) string
	BuildDropTableQueries(name string) []string
	BuildSchemaQuery(name string) string
	BuildListTableQuery() string
	BuildCreateTableQuery(name string) string
	BuildCreateQueries(tableName string, values string, cols string, typ string) []string
	ApplyQueryFilters(restr string, order string, limit string, views string, additionnalRestriction ...string)
	BuildUpdateQuery(col string, value interface{}, set string, cols []string, colValues []string, verify func(string) (string, bool)) (string, []string, []string)
	BuildUpdateQueryWithRestriction(tableName string, record map[string]interface{}, restrictions map[string]interface{}, isOr bool) (string, error)
	BuildUpdateRowQuery(tableName string, record map[string]interface{}, verify func(string) (string, bool)) (string, error)
	BuildUpdateColumnQueries(tableName string, record map[string]interface{}, verify func(string) (string, bool)) ([]string, error)
	Prepare(query string) (*sql.Stmt, error)
	RowResultToMap(rows *sql.Rows, columnNames []string, columnType map[string]string) (map[string]interface{}, error)
	GetDriver() string
	Query(query string) error
	QueryRow(query string) (int64, error)
}

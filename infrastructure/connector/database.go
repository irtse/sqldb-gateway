package connector

import (
	"database/sql"
	"fmt"
	"os"
	"slices"
	"sync"

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
	mutex   = sync.RWMutex{}
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

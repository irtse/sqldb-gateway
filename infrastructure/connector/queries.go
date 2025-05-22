package connector

import (
	"database/sql"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
)

func (db *Database) DeleteQueryWithRestriction(name string, restrictions map[string]interface{}, isOr bool) error {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	if !strings.Contains(db.BuildDeleteQueryWithRestriction(name, restrictions, isOr), "id=") {
		return errors.New("can't delete with a related or id specified")
	}
	return db.Query(db.BuildDeleteQueryWithRestriction(name, restrictions, isOr))
}

func (db *Database) SelectQueryWithRestriction(name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	if strings.Contains(db.BuildSelectQueryWithRestriction(name, restrictions, isOr), "LIKE") {
		fmt.Println("LIKE DETECTED", db.BuildSelectQueryWithRestriction(name, restrictions, isOr))
	}
	res, err := db.QueryAssociativeArray(db.BuildSelectQueryWithRestriction(name, restrictions, isOr))
	if err != nil {
		debug.PrintStack()
		fmt.Println(db.BuildSelectQueryWithRestriction(name, restrictions, isOr), err)
	}
	return res, err
}

func (db *Database) SimpleMathQuery(algo string, name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	return db.QueryAssociativeArray(db.BuildSimpleMathQueryWithRestriction(algo, name, restrictions, isOr))
}

func (db *Database) MathQuery(algo string, name string, naming ...string) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	return db.QueryAssociativeArray(db.BuildMathQuery(algo, name, naming...))
}

func (db *Database) SchemaQuery(name string) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	return db.QueryAssociativeArray(db.BuildSchemaQuery(name))
}

func (db *Database) ListTableQuery() ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	return db.QueryAssociativeArray(db.BuildListTableQuery())
}

func (db *Database) CreateQuery(name string, record map[string]interface{}, verify func(string) (string, bool)) (int64, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	var columns, values []string = []string{}, []string{}

	for key, element := range record {
		_, columns, values = db.BuildUpdateQuery(key, element, "", columns, values, verify)
	}
	for _, query := range db.BuildCreateQueries(name, strings.Join(values, ","), strings.Join(columns, ","), "") {

		if db.GetDriver() == PostgresDriver {
			i, err := db.QueryRow(query)
			if err != nil && strings.Contains(err.Error(), "unique") {
				splitted := strings.Split(err.Error(), "\"")
				if len(splitted) > 0 {
					constraint := splitted[len(splitted)-1]
					field := strings.ReplaceAll(strings.ReplaceAll(constraint, name+"_", ""), "_unique", "")
					return i, errors.New("found a <" + field + "> already existing, it should be unique !")
				}
			}
			return i, err
		} else if db.GetDriver() == MySQLDriver {
			if stmt, err := db.Prepare(query); err != nil {
				return 0, err
			} else if res, err := stmt.Exec(); err != nil {
				return 0, err
			} else if id, err := res.LastInsertId(); err != nil {
				return id, err
			}
		}
	}
	return 0, errors.New("no queries")
}

func (db *Database) CreateTableQuery(name string) error {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	return db.Query(db.BuildCreateTableQuery(name))
}

func (db *Database) UpdateQuery(name string, record map[string]interface{}, restriction map[string]interface{}, isOr bool) error {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	q, err := db.BuildUpdateQueryWithRestriction(name, record, restriction, isOr)
	if err != nil {

		return err
	}
	err = db.Query(q)
	if err != nil && strings.Contains(err.Error(), "unique") {
		splitted := strings.Split(err.Error(), "\"")
		fmt.Println(err.Error(), splitted)
		if len(splitted) > 0 {
			constraint := splitted[len(splitted)-1]
			field := strings.ReplaceAll(strings.ReplaceAll(constraint, name+"_", ""), "_unique", "")
			return errors.New("found a <" + field + "> already existing, it should be unique !")
		}
	}
	return err
}

func (db *Database) DeleteQuery(name string, colName string) error {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	if !strings.Contains(db.BuildDeleteQuery(name, colName), "id=") {
		return errors.New("can't delete with a related or id specified")
	}
	return db.Query(db.BuildDeleteQuery(name, colName))
}

/*
* Prepare a query for execution.
 */
func (db *Database) Prepare(query string) (*sql.Stmt, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	if db.Conn == nil {
		return nil, fmt.Errorf("no connection to database")
	}
	return db.Conn.Prepare(query)
}

/*
* QueryRow executes a query that is expected to return at most one row.
 */
func (db *Database) QueryRow(query string) (int64, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	id := int64(0)
	err := db.Conn.QueryRow(query).Scan(&id)
	return id, err
}

/*
* Query executes a query that returns multiple rows, typically a SELECT.
 */
func (db *Database) Query(query string) error {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	rows, err := db.Conn.Query(query)
	if err != nil {
		return err
	}
	return rows.Close()
}

/*
* QueryAssociativeArray executes a query that returns multiple rows and returns the result as an array of associative arrays.
 */
func (db *Database) QueryAssociativeArray(query string) ([]map[string]interface{}, error) {
	if db == nil || db.Conn == nil {
		db = Open(db)
		defer db.Close()
	}
	rows, err := db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	columnTypes, _ := rows.ColumnTypes()
	columnType := map[string]string{}
	for _, col := range columnTypes {
		columnType[col.Name()] = strings.ToUpper(col.DatabaseTypeName())
	}
	var results []map[string]interface{}
	for rows.Next() {
		if res, err := db.RowResultToMap(rows, cols, columnType); err == nil {
			results = append(results, res)
		} else {
			return nil, err
		}
	}
	return results, nil
}

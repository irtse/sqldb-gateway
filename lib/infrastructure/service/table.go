package service

import (
	"fmt"
	"errors"
	"strings"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	conn "sqldb-ws/lib/infrastructure/connector"
)

var listTablesCmd = map[string]string{
	conn.PostgresDriver: "SELECT table_name :: varchar as name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;", 
	conn.MySQLDriver: "SELECT TABLE_NAME as name FROM information_schema.TABLES WHERE TABLE_TYPE LIKE 'BASE_TABLE';" }

// Table is a table structure description
type TableInfo struct {
	ID 		   int64                                    `json:"id"`
	AssColumns map[string]TableColumnEntity    			`json:"columns"`
	Cols 	   []string                                 `json:"-"`
	InfraService
}

func (t *TableInfo) TableRow(specializedService InfraSpecializedServiceItf) *TableRowInfo {
	row := &TableRowInfo{} 
	row.db = t.db
	row.NoLog = t.NoLog
	row.Fill(t.Name, t.SuperAdmin, t.User, t.Record)
	row.Table = Table(t.db, t.SuperAdmin, t.User, t.Name, map[string]interface{}{})
	row.EmptyCol = &TableColumnInfo{ } 
	row.Name = t.Name
	row.SpecializedService = specializedService
	row.EmptyCol.db = t.db
    return row
}

func (t *TableInfo) TableColumn(specializedService InfraSpecializedServiceItf, views string) *TableColumnInfo {
	col := &TableColumnInfo{ Views: views } 
	col.db = t.db
	col.NoLog = t.NoLog
	col.SpecializedService = specializedService
	col.Fill(t.Name, t.SuperAdmin, t.User, t.Record)
	col.Row = Table(t.db, t.SuperAdmin, t.User, t.Name, map[string]interface{}{},).TableRow(nil)
    return col
}
func querySchemaCmd(name string, tablename string) string {
	if name == conn.MySQLDriver { return "SELECT COLUMN_NAME as name, column_default as default_value, IS_NULLABLE as null, CONCAT(DATA_TYPE, COALESCE(CONCAT('(' , CHARACTER_MAXIMUM_LENGTH, ')'), '')) as type FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = " + conn.Quote(tablename) + ";" } 
	if name == conn.PostgresDriver { return "SELECT column_name :: varchar as name, column_default as default_value, IS_NULLABLE as null, REPLACE(REPLACE(data_type,'character varying','varchar'),'character','char') || COALESCE('(' || character_maximum_length || ')', '') as type, col_description('public." + tablename + "'::regclass, ordinal_position) as comment  from INFORMATION_SCHEMA.COLUMNS where table_name =" + conn.Quote(tablename) + ";" }
    return ""
}

func (t *TableInfo) Template(restriction... string) (interface{}, error) {
	res, err := t.schema(t.Name)
	if err != nil { return nil, err  }
	data := struct {
		Tbl []TableInfo
		Lnk []Link
	}{ res, buildLinks(res), }
	return data, nil
}

func (t *TableInfo) EmptyRecord() (map[string]interface{}, error) {
	res, err := t.schema(t.Name)
	if err != nil || len(res) == 0 || len(res[0].AssColumns) == 0 { 
		return nil, errors.New("any schema available") 
	}
	record := map[string]interface{}{}
	for k, _ := range res[0].AssColumns { 
		if k != "id" { record[k]=nil }
	}
	return record, nil
}

func (t *TableInfo) Math(algo string, restriction... string) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented...")
}

func (t *TableInfo) Get(restriction... string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	schema, err := t.schema(t.Name)
	if err != nil { return t.DBError(nil, err) }
	res := []map[string]interface{}{}
	for _, s := range schema {
		rec := map[string]interface{}{}
		rec["name"] = s.Name
		rec["columns"] = s.Cols
		res = append(res, rec)
	}
	t.Results = res
	return t.Results, nil
}

func (t *TableInfo) schema(name string) ([]TableInfo, error) {
	schema:=[]TableInfo{}
	tables, err := t.db.QueryAssociativeArray(listTablesCmd[t.db.Driver])
	if err != nil { return nil, err }
	for _, row := range tables {
		if element, ok := row["name"]; ok && !(name != "all" && name != fmt.Sprintf("%v", element)) {
			mapped, cols, err := RetrieveTable(fmt.Sprintf("%v", element), t.db.Driver, t.db)
			if err != nil { continue }
			table := TableInfo{ AssColumns : mapped, Cols : cols,}
			table.Name = fmt.Sprintf("%v", element)
			schema = append(schema, table)
		}
	}
	return schema, nil
}

func RetrieveTable(name string, driver string, db *conn.Db) (map[string]TableColumnEntity, []string, error) {
	columns := []string{}
	mapped := map[string]TableColumnEntity{}
	cols, err := db.QueryAssociativeArray(querySchemaCmd(driver, name))
	if err != nil { return mapped, columns, err }
	for _, row := range cols {
		var tableCol TableColumnEntity
		b, _ := json.Marshal(row)
		json.Unmarshal(b, &tableCol)
		if null, ok := row["null"]; ok { tableCol.Null = null == "YES" }
		if tableCol.Default != nil && strings.Contains(tableCol.Default.(string), "NULL") { tableCol.Default = nil }
		mapped[tableCol.Name] = tableCol 
		columns = append(columns, tableCol.Name)
	}
	return mapped, columns, nil
}

func (t *TableInfo) Verify(name string) (string, bool) {
	t.db.ClearFilter()
    schema, err :=t.schema(name)
   	if len(schema) == 0 || err != nil { return name, false }
   	return name, true	
}
func (t *TableInfo) Create() ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	name := fmt.Sprintf("%v", t.Record["name"])
	if name == "" || name == "<nil>" { return nil, errors.New("Missing one of the needed value type & name") }
	query := "CREATE TABLE " + name + " ( id SERIAL PRIMARY KEY, active BOOLEAN DEFAULT TRUE,"
	query = query[:len(query)-1] + " )"
	err := t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	for _, rowtype := range t.Record["fields"].([]interface{}) {
		if fmt.Sprintf("%v", name) != "id" {
			tc := t.TableColumn(nil, "")
			tc.Name=name
			tc.Record = rowtype.(map[string]interface{})

			tc.Create()
		}
	}
	t.Name=name
	_, err = t.Get()
	return t.Results, err
}

func (t *TableInfo) Update() ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	name := fmt.Sprintf("%v", t.Record["name"])
	if name == "" || name == "<nil>" { return nil, errors.New("Missing one of the needed value type & name") }
	for _, rowtype := range t.Record["fields"].([]interface{}) {
		if fmt.Sprintf("%v", name) != "id" {
			tc := t.TableColumn(nil, "")
			tc.Name=name
			col, err := json.Marshal(rowtype)
			if err != nil { continue }
			json.Unmarshal(col, &tc.Record)
			tc.Update()
		}
	}
	t.Name=name
	_, err := t.Get()
	return t.Results, err
}

func (t *TableInfo) Delete(restriction... string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	var err error
	if err = t.db.Query("DROP TABLE " + t.Name); err != nil { return t.DBError(nil, err) }
	if err = t.db.Query("DROP SEQUENCE IF EXISTS sq_" + t.Name); err != nil { return t.DBError(nil, err) }
	t.Results = append(t.Results, map[string]interface{}{ "name" : t.Name })
	return t.Results, err
}
// Generate templates from a scheme
type Link struct {
	Source      string
	Destination string
}

func buildLinks(schema []TableInfo) []Link {
	var links []Link
	for _, ti := range schema {
		for column, _ := range ti.AssColumns {
			if strings.HasSuffix(column, "_id") {
				tokens := strings.Split(column, "_")
				linkedtable := tokens[len(tokens)-2]
				var link Link
				link.Source = ti.Name
				link.Destination = linkedtable
				links = append(links, link)
			}
		}
	}
	return links
}

package service

import (
	"os"
	"fmt"
	"errors"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog/log"
	_ "github.com/go-sql-driver/mysql"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

var listTablesCmd = map[string]string{
	conn.PostgresDriver: "SELECT table_name :: varchar as name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;", 
	conn.MySQLDriver: "SELECT TABLE_NAME as name FROM information_schema.TABLES WHERE TABLE_TYPE LIKE 'BASE_TABLE';" }

// Table is a table structure description
type TableInfo struct {
	AssColumns map[string]string    `json:"columns"`
	Columns    []TableColumnInfo    `json:"-"`
	Rows       []TableRowInfo       `json:"-"`
	InfraService
}
func (t *TableInfo) TableRow(specializedService tool.SpecializedService, adminView bool) *TableRowInfo {
	row := &TableRowInfo{} 
	row.db = t.db
	row.PermService = t.PermService
	row.AdminView = adminView && t.SuperAdmin
	row.Fill(t.Name, t.SuperAdmin, t.User, t.Params, t.Record, t.Method)
	row.Table = Table(t.db, t.SuperAdmin, t.User, t.Name, tool.Params{}, tool.Record{}, t.Method)
	row.EmptyCol = &TableColumnInfo{ } 
	row.EmptyCol.db = t.db
	row.EmptyCol.Name = t.Name
	row.SpecializedService = specializedService
    row.Verified = true
    return row
}

func (t *TableInfo) TableColumn() *TableColumnInfo {
	col := &TableColumnInfo{ } 
	col.db = t.db
	col.PermService = t.PermService
	col.Fill(t.Name, t.SuperAdmin, t.User, t.Params, t.Record, t.Method)
	col.Row = Table(t.db, t.SuperAdmin, t.User, t.Name, tool.Params{}, tool.Record{}, t.Method,).TableRow(nil, true)
    return col
}

func (t *TableInfo) querySchemaCmd(name string, tablename string) string {
	if name == conn.MySQLDriver { return "SELECT COLUMN_NAME as name, IS_NULLABLE as null, CONCAT(DATA_TYPE, COALESCE(CONCAT('(' , CHARACTER_MAXIMUM_LENGTH, ')'), '')) as type FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '" + tablename + "';" } 
	if name == conn.PostgresDriver { return "SELECT column_name :: varchar as name, IS_NULLABLE as null, REPLACE(REPLACE(data_type,'character varying','varchar'),'character','char') || COALESCE('(' || character_maximum_length || ')', '') as type, col_description('public." + tablename + "'::regclass, ordinal_position) as comment  from INFORMATION_SCHEMA.COLUMNS where table_name ='" + tablename + "';" }
    return ""
}

func (t *TableInfo) Template() (interface{}, error) {
	res, err := t.schema(t.Name)
	if err != nil { return nil, err  }
	data := struct {
		Tbl []TableInfo
		Lnk []Link
	}{ res, buildLinks(res), }
	return data, nil
}

func (t *TableInfo) EmptyRecord() (tool.Record, error) {
	res, err := t.schema(t.Name)
	if err != nil || len(res) == 0 || len(res[0].Columns) == 0 { 
		return nil, errors.New("any schema available") 
	}
	record := tool.Record{}
	for k, _ := range res[0].AssColumns { 
		if k != tool.SpecialIDParam { record[k]=nil }
	}
	return record, nil
}
// GetAssociativeArray : Provide table data as an associative arra
func (t *TableInfo) Get() (tool.Results, error) {
	schema, err := t.schema(t.Name)
	if err != nil { return DBError(nil, err) }
	res := tool.Results{}
	for _, s := range schema {
		rec := tool.Record{}
		rec[entities.NAMEATTR] = s.Name
		rec["columns"] = s.AssColumns
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
		for _, element := range row {
			if name == tool.ReservedParam || fmt.Sprintf("%v", element) == name {
				table, err := EmptyTable(t.db, fmt.Sprintf("%v", element)).get()
				if err != nil { log.Error().Msg(err.Error()); continue }
				schema = append(schema, *table)
			}
		}
	}
	return schema, nil
}

func (t *TableInfo) get() (*TableInfo, error) {
	cols, err := t.db.QueryAssociativeArray(t.querySchemaCmd(t.db.Driver, t.Name))
	if err != nil { return nil, err }
	t.AssColumns = make(map[string]string)
	for _, row := range cols {
		var name, null, rowtype, comment string
		for key, element := range row {
			if key == entities.NAMEATTR { name = fmt.Sprintf("%s", element) }
			if key == "null" { 
				null = fmt.Sprintf("%s", element) 
				if null == "NO" { null="required"
				} else { null="nullable" }
			}
			if key == entities.TYPEATTR { rowtype = fmt.Sprintf("%s", element) }
			if key == "comment" { comment = fmt.Sprintf("%s", element) }
		}
		t.AssColumns[name] = rowtype
		if comment != "<nil>" && strings.TrimSpace(comment) != "" {
			t.AssColumns[name] = t.AssColumns[name] + "|" + comment + "|" + null
		}
	}
	return t, nil
}

func (t *TableInfo) Verify(name string) (string, bool) {
    schema, err :=t.schema(name)
   	if len(schema) == 0 || err !=nil { return name, false }
   	return name, true	
}
func (t *TableInfo) Create() (tool.Results, error) {
	v := Validator[entities.TableEntity]()
	v.data = entities.TableEntity{}
	te, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New(
		"Not a proper struct to create a table - expect <TableEntity> Scheme " + err.Error()) }
	query := "CREATE TABLE " + te.Name + " ( id SERIAL PRIMARY KEY,"
	query = query[:len(query)-1] + " )"
	_, err = t.db.Query(query)
	if err != nil { return DBError(nil, err) }
	for name, rowtype := range te.Columns {
		if fmt.Sprintf("%v", name) != tool.SpecialIDParam {
			tc := t.TableColumn()
			tc.Name=te.Name
			col, err := json.Marshal(rowtype)
			if err != nil { continue }
			json.Unmarshal(col, &tc.Record)
			tc.CreateOrUpdate()
		}
	}
	t.Name=te.Name
	_, err = t.Get()
	if t.PermService != nil && entities.IsRootDB(te.Name) {
		t.PermService.SpecializedFill(t.Params, 
			tool.Record{ "name" : te.Name, 
						 "results" : t.Results, 
						 "info" : "" }, 
			t.Method)
		t.PermService.CreateOrUpdate()
	}
	return t.Results, err
}

func (t *TableInfo) Update() (tool.Results, error) {
	return nil, errors.New("not implemented for integrity reason")
	/*if strings.Contains(t.Name, "db") { log.Error().Msg("can't rename protected root db.") }
	v := Validator[entities.TableUpdateEntity]()
	v.data = entities.TableUpdateEntity{}
	tcue, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New(
		"Not a proper struct to update a table - expect <entities.TableUpdateEntity> Scheme " + err.Error()) }
	query := "ALTER TABLE IF EXISTS " + t.Name + " RENAME TO " + tcue.Name + ";"
	rows, err := t.db.Query(query)
	if err != nil { return DBError(nil, err) }
	defer rows.Close()
    t.Results = append(t.Results, map[string]interface{}{ "name" : tcue.Name, "old" : t.Name })
	err = t.PermService.Manage(Info{ Name : t.Name, Results : t.Results }, "", tool.UPDATE)
	return t.Results, err*/
}

func (t *TableInfo) CreateOrUpdate() (tool.Results, error) { 
	if _, ok := t.Verify(t.Name); !ok && t.Name != tool.ReservedParam { return t.Create() }
	return t.Update()
}

func (t *TableInfo) Delete() (tool.Results, error) {
	if entities.IsRootDB(t.Name) || t.Name == tool.ReservedParam { log.Error().Msg("can't delete protected root db.") }
	if _, err := t.db.Query("DROP TABLE " + t.Name); err != nil { return DBError(nil, err) }
	if _, err := t.db.Query("DROP SEQUENCE IF EXISTS sq_" + t.Name); err != nil { return DBError(nil, err) }
	t.Results = append(t.Results, tool.Record{ entities.NAMEATTR : t.Name })
	t.PermService.SpecializedFill(t.Params, 
		tool.Record{ "name" : t.Name, 
					 "results" : t.Results, 
					 "info" : "" }, 
		t.Method)
	_, err := t.PermService.Delete()
	return t.Results, err
}

func (t *TableInfo) Add() (tool.Results, error) { return nil, errors.New("not implemented...") }

func (t *TableInfo) Remove() (tool.Results, error) { return nil, errors.New("not implemented...") }

func (t *TableInfo) Import(filename string) (tool.Results, error) {
	var jsonSource []TableInfo
	byteValue, _ := os.ReadFile(filename)
	err := json.Unmarshal([]byte(byteValue), &jsonSource) 
	if err != nil { return DBError(nil, err) }
	for _, ti := range jsonSource {
		ti.db = t.db
		if t.Method == tool.DELETE { _, err = ti.Delete() 
		} else { _, err = ti.Create() }
		if err != nil { log.Error().Msg(err.Error()) }
	}
	return t.Results, nil
}
// Generate templates from a scheme
type Link struct {
	Source      string
	Destination string
}

func buildLinks(schema []TableInfo) []Link {
	var links []Link
	for _, ti := range schema {
		fmt.Println(ti.Name)
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

func (t *TableInfo) Link() (tool.Results, error) {
	//will generate a many to many table
	if _, ok := t.Params[tool.RootToTableParam]; !ok || t.Name == tool.ReservedParam { return nil, errors.New("no destination table") }
	otherName := t.Params[tool.RootToTableParam]
	ok := true; ok2 := true
	if strings.Contains(otherName[:2], "db") { 
		_, ok = t.Verify(t.Name + "_" + otherName[2:])
	} else { _, ok = t.Verify(t.Name + "_" + otherName) }
	if strings.Contains(t.Name[:2], "db") { _, ok2 = t.Verify(otherName + "_" + t.Name[2:])
	} else { _, ok2 = t.Verify(otherName + "_" + t.Name) }
	if !ok && !ok2 {
		v := Validator[entities.ShallowTableEntity]()
		v.data = entities.ShallowTableEntity{}
		te, err := v.ValidateStruct(t.Record)
		if err != nil { return nil, errors.New("Not a proper struct to create a table - expect <TableEntity> Scheme " + err.Error()) }
	    name := ""
		rawName := t.Name
		if strings.Contains(otherName[:2], "db") { name = rawName + "_" + otherName[2:]
	    } else { name = rawName + "_" + otherName } 
		if te.Name != "" { name = te.Name }
		record := tool.Record{ "name" : name, 
	                      "columns" : []map[string]interface{}{
							map[string]interface{}{ "name" : rawName + "_id", "type" : "integer", "not_null" : true, "foreign_table": otherName, },
							map[string]interface{}{ "name" : otherName + "_id", "type" : "integer", "not_null" : true, "foreign_table": rawName, },
						  },
						}
		/*if te.Columns != nil {
			for _, col := range te.Columns {
				var mapped map[string]interface{}
				inrec, _ := json.Marshal(col)
    			json.Unmarshal(inrec, &mapped)
				record["columns"]= append(record["columns"].([]map[string]interface{}), mapped) 
			}
		}*/
		t.Record=record
		return t.Create()
	}
	return nil, errors.New("link table already exists") 
}

func (t *TableInfo) UnLink() (tool.Results, error) {
	//will delete a many to many table
	if _, ok := t.Params[tool.RootToTableParam]; !ok || t.Name == tool.ReservedParam { return nil, errors.New("no destination table") }
	rootName := t.Name 
	otherName := t.Params[tool.RootToTableParam]
	if strings.Contains(otherName[:2], "db") { otherName=otherName[2:] }
	v := Validator[entities.ShallowTableEntity]()
	v.data = entities.ShallowTableEntity{}
	te, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to delete a table - expect <TableEntity> Scheme " + err.Error()) }
	name := ""
	if strings.Contains(otherName[:2], "db") {  name = rootName + "_" + otherName[2:]
	} else { name = rootName + "_" + otherName }
	if te.Name != "" { name = te.Name }
	if _, ok := t.Verify(name); !ok {
		if strings.Contains(rootName[:2], "db") {
			if _, ok2 := t.Verify(otherName + "_" + rootName[2:]); ok2 { name = otherName + "_" + rootName[2:] }
		} else {
			if _, ok2 := t.Verify(otherName + "_" + rootName); ok2 { name = otherName + "_" + rootName }
		}
	}
	t.Name=name
	return t.Delete()
}


/*func (db *Db) ListSequences() (Rows, error) {
	return db.QueryAssociativeArray("SELECT sequence_name :: varchar FROM information_schema.sequences WHERE sequence_schema = 'public' ORDER BY sequence_name;")
}*/
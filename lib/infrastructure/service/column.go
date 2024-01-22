package service

import (
    "os"
	"errors"
	"strings"
	"encoding/json"
	"github.com/lib/pq"
	tool "sqldb-ws/lib"
	"github.com/rs/zerolog/log"
	_ "github.com/go-sql-driver/mysql"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// Table is a table structure description
type TableColumnInfo struct { 
	perms      	*PermissionInfo
	Row 		*TableRowInfo
	InfraService 
}

func (t *TableColumnInfo) Template() (interface{}, error) { return t.Get() }

func (t *TableColumnInfo) Get() (tool.Results, error) {
	t.db = ToFilter(t.Name, t.Params, t.db)
	d, err := t.db.SelectResults(t.Name)
	t.Results = d
	if err != nil { return DBError(nil, err) }
	return t.Results, nil
}

func (t *TableColumnInfo) get(name string) (tool.Results, error) {
	empty := EmptyTable(t.db, t.Name)
	if empty == nil { return nil, errors.New("no table available...") }
	scheme, err := empty.Get()
	if err != nil { return nil, err }
	res := tool.Results{}
	rec := tool.Record{}
	if len(scheme) > 0 { 
		for k, v := range scheme[0]["columns"].(map[string]string) {
			if k == name { rec[k] = v }
		}
		res = append(res, rec)
	}
	return res, nil
}

func (t *TableColumnInfo) Verify(name string) (string, bool) {
	empty := EmptyTable(t.db, t.Name)
	if empty == nil { return "", false }
	scheme, err := empty.Get()
	if err != nil { return "", false }
	typ := ""
	if len(scheme) > 0 { 
		typ = strings.Split(scheme[0]["columns"].(map[string]string)[name], "|")[0] 
	}
	return typ, typ != "" 
}
func (t *TableColumnInfo) Create() (tool.Results, error) {
	v := Validator[entities.TableColumnEntity]()
	v.data = entities.TableColumnEntity{}
	tcce, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to create a column - expect <entities.TableColumnEntity> Scheme " + err.Error()) }
	query := "ALTER TABLE " + t.Name + " ADD " + tcce.Name + " " + tcce.Type
	if t.db.Driver == conn.MySQLDriver {
		if strings.TrimSpace(tcce.Comment) != "" { query += " COMMENT " + pq.QuoteLiteral(tcce.Comment) }
	}
	rows, err := t.db.Query(query)
	if err != nil { return DBError(nil, err) }
	defer rows.Close()
	err = t.update(tcce)
	if err != nil { return DBError(nil, err) }
	if len(t.Name) > 1 {
		t.PermService.SpecializedFill(t.Params, 
			                          tool.Record{ "name" : t.Name, 
									               "results" : tool.Results{t.Record}, 
												   "info" : tcce.Name }, 
									  t.Method)
		t.PermService.CreateOrUpdate()
	}
	res, err := t.get(tcce.Name)
	if err != nil { return nil, err }
	return res, nil
}

func (t *TableColumnInfo) Update() (tool.Results, error) {
	v := Validator[entities.TableColumnEntity]()
	v.data = entities.TableColumnEntity{}
	tcue, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to update a column - expect <entities.TableColumnEntity> Scheme " + err.Error()) }
	if strings.TrimSpace(tcue.NewName) != "" {
		if strings.Contains(t.Name, "db") { return nil, errors.New("can't rename protected root db columns.") }
		query := "ALTER TABLE " + t.Name + " RENAME COLUMN " + tcue.Name + " TO " + tcue.NewName + ";"
		rows, err := t.db.Query(query)
		if err != nil { return DBError(nil, err) }
		defer rows.Close()
	}
	err = t.update(tcue)
	if err != nil { return DBError(nil, err) }
	if len(t.Name) > 1 {
		t.PermService.SpecializedFill(t.Params, 
									  tool.Record{ "name" : t.Name, 
												   "results" : tool.Results{t.Record}, 
												   "info" : tcue.Name }, 
									  t.Method)
		t.PermService.CreateOrUpdate()
	}
	res, err := t.get(tcue.Name)
	if err != nil { return nil, err }
	return res, err
}

func (t *TableColumnInfo) update(tcce *entities.TableColumnEntity) (error) {
	if strings.TrimSpace(tcce.Constraint) != "" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + "  ADD CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + " " + strings.ToUpper(tcce.Constraint) + "(" + tcce.Name + ");"
		_, err := t.db.Query(query)
		if err != nil { return err }
	}
	if strings.TrimSpace(tcce.Constraint) != "" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + " " + strings.ToUpper(tcce.Constraint) + "(" + tcce.Name + ");"
		_, err := t.db.Query(query)
		if err != nil { return err }
	}
	if strings.TrimSpace(tcce.ForeignTable) != "" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT fk_" + tcce.Name + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT  fk_" + tcce.Name +  " FOREIGN KEY(" + tcce.Name  + ") REFERENCES " + tcce.ForeignTable + "(id);"
        _, err := t.db.Query(query)
		if err != nil { return err }
	}
	if tcce.Default != "" {
		query := "ALTER TABLE " + t.Name + " ALTER " + tcce.Name  + " SET DEFAULT " + conn.FormatForSQL(tcce.Type, tcce.Default) + ";"
        _, err := t.db.Query(query)
		if err != nil { return err } // then iterate on field to update value if null
		params := tool.Params{ tool.RootSQLFilterParam : tcce.Name + " IS NULL" }
		record := tool.Record{ tcce.Name : tcce.Default }
		t.Row.SpecializedFill(params, record, tool.UPDATE)
		t.Row.CreateOrUpdate()
	}
	if tcce.NotNull {
		query := "ALTER TABLE " + t.Name + "  ALTER COLUMN " + tcce.Name + " SET NOT NULL;"
        _, err := t.db.Query(query)
		if err != nil { return err }
	}
	if t.db.Driver == conn.PostgresDriver { // PG COMMENT
		if strings.TrimSpace(tcce.Comment) != "" {
			query := "COMMENT ON COLUMN " + t.Name + "." + tcce.Name + " IS '" + tcce.Comment + "'"
			_, err := t.db.Query(query)
			if err != nil { return err }
		}
	}
	return nil
}

func (t *TableColumnInfo) CreateOrUpdate() (tool.Results, error) {
	if col, ok:= t.Record[entities.NAMEATTR]; ok {
		if _, ok := t.Verify(col.(string)); ok { return t.Update() 
		} else { return t.Create() }
	}
	return nil, errors.New("nothing to do...")
}

func (t *TableColumnInfo) Delete() (tool.Results, error) {
	if strings.Contains(t.Name, "db") { log.Error().Msg("can't delete protected root db columns.") }
	for _, col := range strings.Split(t.Params[tool.RootColumnsParam], ",") {
		query := "ALTER TABLE " + t.Name + " DROP " + col
		rows, err := t.db.Query(query)
		if err != nil { return DBError(nil, err) }
		defer rows.Close()
		t.Results = append(t.Results, tool.Record{ entities.NAMEATTR : col })
		t.PermService.SpecializedFill(t.Params, 
			tool.Record{ "name" : t.Name, 
						 "results" : t.Results, 
						 "info" : col }, 
			t.Method)
		t.PermService.Delete()
	}
	return t.Results, nil
}

func (t *TableColumnInfo) Add() (tool.Results, error) { 
	return nil, errors.New("not implemented...")
}

func (t *TableColumnInfo) Remove() (tool.Results, error) { 
	return nil, errors.New("not implemented...")
}

func (t *TableColumnInfo) Import(filename string) (tool.Results, error) {
	var jsonSource []TableColumnInfo
	byteValue, _ := os.ReadFile(filename)
	err := json.Unmarshal([]byte(byteValue), &jsonSource)
	if err != nil { return DBError(nil, err) }
	for _, col := range jsonSource {
		col.db = t.db
		if t.Method == tool.DELETE { _, err = col.Delete() 
		} else { _, err = col.CreateOrUpdate() }
		if err != nil { log.Error().Msg(err.Error()) }
	}
	return t.Results, nil
}

func (t *TableColumnInfo) Link() (tool.Results, error) {
	var err error
	if _, ok := t.Params[tool.RootToTableParam]; !ok { return nil, errors.New("no destination table") }
	otherName := t.Params[tool.RootToTableParam]
	cols := strings.Split(t.Params[tool.RootColumnsParam], ",")
	res := tool.Results{}
	for _, col := range cols {
		rename := otherName + "_id"
		t.Record = tool.Record{ "name" : col, "new_name": rename, "type" : "integer", "foreign_table": otherName, "not_null" : true }
		res, err = t.CreateOrUpdate()
		if err == nil { t.Results = append(t.Results, res...) }
	}
	return t.Results, nil
}
func (t *TableColumnInfo) UnLink() (tool.Results, error) {
	if _, ok := t.Params[tool.RootToTableParam]; !ok { return nil, errors.New("no destination table") }
	t.Params[tool.RootColumnsParam] = t.Params[tool.RootToTableParam] + "_id"
	return t.Delete()
}
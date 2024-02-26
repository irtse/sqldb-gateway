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
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)
// Table is a table structure description
type TableColumnInfo struct { 
	Row 		*TableRowInfo
	InfraService 
}

func (t *TableColumnInfo) Template(restriction... string) (interface{}, error) { return t.Get(restriction...) }

func (t *TableColumnInfo) Get(restriction... string) (tool.Results, error) {
	t.db.ToFilter(t.Name, t.Params, restriction...)
	d, err := t.db.SelectResults(t.Name)
	t.Results = d
	if err != nil { return t.DBError(nil, err) }
	return t.Results, nil
}

func (t *TableColumnInfo) get(name string) (tool.Results, error) {
	t.db.ClearFilter()
	empty := EmptyTable(t.db, t.Name)
	if empty == nil { return nil, errors.New("no table available...") }
	scheme, err := empty.Get()
	if err != nil { return nil, err }
	res := tool.Results{}
	rec := tool.Record{}
	if len(scheme) > 0 { 
		b, err := json.Marshal(scheme[0])
		if err != nil { return res, err  }
		err = json.Unmarshal(b, &rec)
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
		var info TableInfo
		b, err := json.Marshal(scheme[0])
		if err != nil { return typ, typ != ""  }
		err = json.Unmarshal(b, &info)
		if err != nil { return typ, typ != "" }
		col := info.AssColumns[name]
		if col.Null {
			typ = col.Type + ":nullable"
		} else { typ = col.Type + ":required" }
	}
	return typ, typ != "" 
}
func (t *TableColumnInfo) Create() (tool.Results, error) {
	t.db.ClearFilter()
	v := Validator[entities.TableColumnEntity]()
	v.data = entities.TableColumnEntity{}
	tcce, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to create a column - expect <entities.TableColumnEntity> Scheme " + err.Error()) }
	found := false
	for _, verifiedType := range tool.DATATYPE {
		if strings.Contains(strings.ToUpper(tcce.Type), verifiedType) {
			found = true; break;
		}
	}
	if ! found { return nil, errors.New("not allowed type") }
	if strings.Contains(strings.ToLower(tcce.Type), "enum") && t.db.Driver == conn.PostgresDriver {
		query := "CREATE TYPE " + t.Name + "_" + tcce.Name  + " AS " + tcce.Type
		t.db.Query(query)
	}
	query := ""
	if strings.Contains(strings.ToLower(tcce.Type), "enum") && t.db.Driver == conn.PostgresDriver {
		query = "ALTER TABLE " + t.Name + " ADD " + tcce.Name + " " + t.Name + "_" + tcce.Name + "  NULL"
	} else { query = "ALTER TABLE " + t.Name + " ADD " + tcce.Name + " " + tcce.Type + "  NULL" }
	
	if t.db.Driver == conn.MySQLDriver {
		if strings.TrimSpace(tcce.Comment) != "" { query += " COMMENT " + pq.QuoteLiteral(tcce.Comment) }
	}
	err = t.db.Query(query)
	if err != nil { return t.Update() }
	err = t.update(tcce)
	if err != nil { return t.DBError(nil, err) }
	res, err := t.get(tcce.Name)
	if err != nil { return nil, err }
	return res, nil
}

func (t *TableColumnInfo) Update() (tool.Results, error) {
	t.db.ClearFilter()
	v := Validator[entities.TableColumnEntity]()
	v.data = entities.TableColumnEntity{}
	tcue, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to update a column - expect <entities.TableColumnEntity> Scheme " + err.Error()) }
	err = t.update(tcue)
	if err != nil { return t.DBError(nil, err) }
	found := false
	for _, verifiedType := range tool.DATATYPE {
		if strings.Contains(strings.ToUpper(tcue.Type), verifiedType) {
			found = true; break;
		}
	}
	if ! found { return nil, errors.New("not allowed type") }
	if strings.TrimSpace(tcue.NewName) != "" {
		if strings.Contains(t.Name, "db") { return nil, errors.New("can't rename protected root db columns.") }
		query := "ALTER TABLE " + t.Name + " RENAME COLUMN " + tcue.Name + " TO " + tcue.NewName + ";"
		err := t.db.Query(query)
		if err != nil { return t.DBError(nil, err) }
	}
	res, err := t.get(tcue.Name)
	if err != nil { return nil, err }
	return res, err
}

func (t *TableColumnInfo) update(tcce *entities.TableColumnEntity) (error) {
	if strings.TrimSpace(tcce.Constraint) != "" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT " + t.Name + "_" + tcce.Name + "_" + tcce.Constraint + " " + strings.ToUpper(tcce.Constraint) + "(" + tcce.Name + ");"
		t.db.Query(query)
	}
	if strings.TrimSpace(tcce.ForeignTable) != "" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT fk_" + tcce.Name + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT  fk_" + tcce.Name +  " FOREIGN KEY(" + tcce.Name  + ") REFERENCES " + tcce.ForeignTable + "(id);"
		t.db.Query(query)
	}
	if tcce.Default != "" && conn.FormatForSQL(tcce.Type, tcce.Default) != "NULL" && !strings.Contains(strings.ToLower(tcce.Type), "bool") {
		query := "ALTER TABLE " + t.Name + " ALTER " + tcce.Name  + " SET DEFAULT " + conn.FormatForSQL(tcce.Type, tcce.Default) + ";"
        err := t.db.Query(query)
		if err != nil { return err } // then iterate on field to update value if null
		/*record := tool.Record{ tcce.Name : tcce.Default }
		t.Row.SpecializedFill(tool.Params{}, record, tool.UPDATE)
		t.Row.CreateOrUpdate(tcce.Name + " IS NULL")*/
	}
	if !tcce.Null {
		query := "ALTER TABLE " + t.Name + " ALTER COLUMN " + tcce.Name + " SET NOT NULL;"
        t.db.Query(query)
	}
	if tcce.Null {
		query := "ALTER TABLE " + t.Name + " ALTER COLUMN " + tcce.Name + " DROP NOT NULL;"
        t.db.Query(query)
	}
	if t.db.Driver == conn.PostgresDriver { // PG COMMENT
		if strings.TrimSpace(tcce.Comment) != "" {
			query := "COMMENT ON COLUMN " + t.Name + "." + tcce.Name + " IS " + conn.Quote(tcce.Comment) + ""
			t.db.Query(query)
		}
	}
	return nil
}

func (t *TableColumnInfo) CreateOrUpdate(restriction... string) (tool.Results, error) {
	return t.Create()
}

func (t *TableColumnInfo) Delete(restriction... string) (tool.Results, error) {
	t.db.ClearFilter()
	if strings.Contains(t.Name, "db") { log.Error().Msg("can't delete protected root db columns.") }
	for _, col := range strings.Split(t.Params[tool.RootColumnsParam], ",") {
		query := "ALTER TABLE " + t.Name + " DROP " + col
		err := t.db.Query(query)
		if err != nil { return t.DBError(nil, err) }
		t.Results = append(t.Results, tool.Record{ entities.NAMEATTR : col })	
	}
	return t.Results, nil
}

func (t *TableColumnInfo) Add() (tool.Results, error) { 
	return nil, errors.New("not implemented...")
}

func (t *TableColumnInfo) Remove() (tool.Results, error) { 
	return nil, errors.New("not implemented...")
}

func (t *TableColumnInfo) Import(filename string, restriction... string) (tool.Results, error) {
	t.db.ClearFilter()
	var jsonSource []TableColumnInfo
	byteValue, _ := os.ReadFile(filename)
	err := json.Unmarshal([]byte(byteValue), &jsonSource)
	if err != nil { return t.DBError(nil, err) }
	for _, col := range jsonSource {
		col.db = t.db
		if t.Method == tool.DELETE { col.Delete() 
		} else { col.CreateOrUpdate(restriction...) }
	}
	return t.Results, nil
}
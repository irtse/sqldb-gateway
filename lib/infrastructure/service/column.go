package service

import (
	"fmt"
	"errors"
	"strings"
	"encoding/json"
	"github.com/rs/zerolog/log"
	conn "sqldb-ws/lib/infrastructure/connector"
)
type TableColumnEntity struct { // definition a db table columns
	Name string         `json:"name" validate:"required"`
	Label string        `json:"label"`
	Type string         `json:"type"`
	Index int64         `json:"-"`
	Default interface{} `json:"default_value"`
	Level 	string 		`json:"read_level"`
	ForeignTable string `json:"link"`
	Readonly bool		`json:"readonly"`
	Constraint string   `json:"constraints"`
	Null bool           `json:"nullable"`
	Comment string      `json:"comment"`
	NewName string      `json:"-"`
}
// Table is a table structure description
type TableColumnInfo struct { 
	Row 		*TableRowInfo
	Views 	    string
	InfraService 
}

func (t *TableColumnInfo) Template(restriction... string) (interface{}, error) { return t.Get(restriction...) }

func (t *TableColumnInfo) Count(restriction... string) ([]map[string]interface{}, error) {
	t.db.SQLView = t.Views
	if t.SpecializedService != nil {
		restr, _, order, limit := t.SpecializedService.ConfigureFilter(t.Name)
		if restr != "" { t.db.SQLRestriction = restr }
		if len(restriction) > 0 { 
			for _, r := range restriction {
				if r == "" { continue }
				if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + r + ")"
				} else { t.db.SQLRestriction = r }
			}
		}
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	var err error; var count int64
	if t.db.Driver == conn.PostgresDriver { 
		count, err = t.db.QueryRow(t.db.BuildCount(t.Name))
		if err != nil { return nil, err }
	}
	if t.db.Driver == conn.MySQLDriver {
		stmt, err := t.db.Prepare(t.db.BuildCount(t.Name))
		if err != nil { return t.DBError(nil, err) }
		res, err := stmt.Exec()
		if err != nil { return nil, err }
		count, err = res.LastInsertId()
		if err != nil { return t.DBError(nil, err) }
	}
	if err != nil { return t.DBError(nil, err) }
	t.Results = append(t.Results, map[string]interface{}{ "count" : count, })
	return t.Results, nil
}

func (t *TableColumnInfo) Get(restriction... string) ([]map[string]interface{}, error) {
	t.db.SQLView = t.Views
	if t.SpecializedService != nil {
		restr, _, order, limit := t.SpecializedService.ConfigureFilter(t.Name)
		if restr != "" { t.db.SQLRestriction = restr }
		if len(restriction) > 0 { 
			for _, r := range restriction {
				if r == "" { continue }
				if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + r + ")"
				} else { t.db.SQLRestriction = r }
			}
		}
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	d, err := t.db.SelectResults(t.Name)
	t.Results = d
	if err != nil { return t.DBError(nil, err) }
	return t.Results, nil
}

func (t *TableColumnInfo) get(name string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	empty := EmptyTable(t.db, t.Name)
	if empty == nil { return nil, errors.New("no table available...") }
	scheme, err := empty.Get()
	if err != nil { return nil, err }
	res := []map[string]interface{}{}
	rec := map[string]interface{}{}
	if len(scheme) > 0 { 
		b, err := json.Marshal(scheme[0])
		if err != nil { return res, err  }
		err = json.Unmarshal(b, &rec)
		res = append(res, rec)
	}
	return res, nil
}

func (t *TableColumnInfo) Verify(name string) (string, bool) {
	mapped, _, err := RetrieveTable(t.Name, t.db.Driver, t.db)
	if err != nil { return "", false }
	typ := ""
	col := mapped[name]
	if col.Null { typ = col.Type + ":nullable" } else { typ = col.Type + ":required" }
	return typ, typ != "" 
}
func enumName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), ",", "_"), "'", ""), "(", "__"), ")", ""), " ", "")
}

func reverseEnumName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), "__", "('"), "_", "','") + "')"
}

func (t *TableColumnInfo) Create() ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	typ := fmt.Sprintf("%v", t.Record["type"])
	name := fmt.Sprintf("%v", t.Record["name"])
	if typ == "" || typ == "<nil>" || name == "" || name == "<nil>" { return nil, errors.New("Missing one of the needed value type & name") }
	if strings.Contains(strings.ToLower(typ), "enum") && t.db.Driver == conn.PostgresDriver {
		query := ""
		if strings.Contains(strings.ToLower(typ), "')"){ query = "CREATE TYPE " + enumName(typ) + " AS " + typ
		} else { query = "CREATE TYPE " + enumName(typ) + " AS " + reverseEnumName(typ) }
		t.db.Query(query)
	}
	query := ""
	if strings.Contains(strings.ToLower(typ), "enum") && t.db.Driver == conn.PostgresDriver {
		query = "ALTER TABLE " + t.Name + " ADD " + name + " " + enumName(typ) + " NULL"
	} else { query = "ALTER TABLE " + t.Name + " ADD " + name + " " + typ + " NULL" }

	err := t.db.Query(query)
	err = t.update(t.Record)
	if err != nil { return t.DBError(nil, err) }
	res, err := t.get(name)
	if err != nil { return nil, err }
	return res, nil
}

func (t *TableColumnInfo) Update() ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	typ := fmt.Sprintf("%v", t.Record["type"])
	name := fmt.Sprintf("%v", t.Record["name"])
	if typ == "" || typ == "<nil>" || name == "" || name == "<nil>" { return nil, errors.New("Missing one of the needed value type & name") }
	err := t.update(t.Record)
	if err != nil { return t.DBError(nil, err) }
	if strings.TrimSpace(name) != "" && !strings.Contains(t.Name, "db") {
		col := strings.Split(t.Views, ",")[0]
		query := "ALTER TABLE " + t.Name + " RENAME COLUMN " + col + " TO " + name + ";" // CHANGE colu
		err := t.db.Query(query)
		if err != nil { return t.DBError(nil, err) }
	}
	res, err := t.get(name)
	if err != nil { return nil, err }
	return res, err
}

func (t *TableColumnInfo) update(record map[string]interface{}) (error) {
	typ := fmt.Sprintf("%v", t.Record["type"]); name := fmt.Sprintf("%v", t.Record["name"])
	constraint := fmt.Sprintf("%v", t.Record["constraints"]); 
	fk := fmt.Sprintf("%v", t.Record["foreign_table"]); def := fmt.Sprintf("%v", t.Record["default_value"]);
	if strings.TrimSpace(constraint) != "" && strings.TrimSpace(constraint) != "<nil>" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT " + t.Name + "_" + name + "_" + constraint + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT " + t.Name + "_" + name + "_" + constraint + " " + strings.ToUpper(constraint) + "(" + name + ");"
		t.db.Query(query)
	}
	if strings.TrimSpace(fk) != "" && strings.TrimSpace(fk) != "<nil>" {
		query := "ALTER TABLE " + t.Name + " DROP CONSTRAINT fk_" + name + ";"
		t.db.Query(query)
		query = "ALTER TABLE " + t.Name + " ADD CONSTRAINT  fk_" + name +  " FOREIGN KEY(" + name + ") REFERENCES " + fk + "(id) ON DELETE CASCADE;"
		t.db.Query(query)
	}
	if def != "" && strings.TrimSpace(def) != "<nil>" && conn.FormatForSQL(typ, def) != "NULL"  {
		query := "ALTER TABLE " + t.Name + " ALTER " + name  + " SET DEFAULT " + conn.FormatForSQL(typ, def) + ";"
        err := t.db.Query(query)
		if err != nil { return err } // then iterate on field to update value if null
	}
	return nil
}

func (t *TableColumnInfo) Delete(restriction... string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	if strings.Contains(t.Name, "db") { log.Error().Msg("can't delete protected root db columns.") }
	for _, col := range strings.Split(t.Views, ",") {
		query := "ALTER TABLE " + t.Name + " DROP " + col
		err := t.db.Query(query)
		if err != nil { return t.DBError(nil, err) }
		t.Results = append(t.Results, map[string]interface{}{ "name" : col })	
	}
	return t.Results, nil
}
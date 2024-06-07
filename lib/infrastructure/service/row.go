package service

import (
	"fmt"
	"errors"
	"strings"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TableRowInfo struct {
	Table				*TableInfo
	EmptyCol            *TableColumnInfo
	InfraService
}

func (t *TableRowInfo) Template(restriction... string) (interface{}, error) { return t.Get(restriction...) }

func (t *TableRowInfo) Verify(name string) (string, bool) {
	res, err := t.Get("id=" + name)
	return name, err == nil && len(res) > 0
}

func (t *TableRowInfo) Count(restriction... string) ([]map[string]interface{}, error) {
	err := t.setupFilter(false, false, restriction...)
	if err != nil { return nil, err }
	var count int64
	if t.db.Driver == conn.PostgresDriver { 
		count, err = t.db.QueryRow(t.db.BuildCount(t.Table.Name))
		if err != nil { return nil, err }
	}
	if t.db.Driver == conn.MySQLDriver {
		stmt, err := t.db.Prepare(t.db.BuildCount(t.Table.Name))
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

func (t *TableRowInfo) Get(restriction... string) ([]map[string]interface{}, error) {
	var err error
	if err = t.setupFilter(false, false, restriction...); err != nil { return nil, err }
	if t.Results, err = t.db.SelectResults(t.Table.Name); err != nil { return t.DBError(nil, err) }
	return t.Results, nil
}

func (t *TableRowInfo) Create() ([]map[string]interface{}, error) {
	if len(t.Record) == 0 { return nil, errors.New("no data to insert") }
	t.db.ClearFilter()
	var id int64; var err error
	result := []map[string]interface{}{}
	columns := ""; values := ""
	if t.SpecializedService != nil {
		r, ok, forceChange := t.SpecializedService.VerifyRowAutomation(t.Record, t.Name)
		if !ok { return nil, errors.New("verification failed.") }
		if forceChange { t.Record = r }
	}
	for key, element := range t.Record {
		if ((strings.Contains(key, "_id") && (fmt.Sprintf("%v", element) == "0" || fmt.Sprintf("%v", element) == "")) || key == "id") { continue }
 		t.EmptyCol.Name = t.Name
		typ, _ := t.EmptyCol.Verify(key) 
		realType := strings.Split(typ, ":")[0]
		if realType != "" { 
			if conn.FormatForSQL(realType, element) == "" { continue }
			columns += key + ","
			values += conn.FormatForSQL(realType, element) + "," 
		}
	}
	query := "INSERT INTO " + t.Table.Name + "(" + conn.RemoveLastChar(columns) + ") VALUES (" + conn.RemoveLastChar(values) + ")"
	if t.db.Driver == conn.PostgresDriver { 
		id, err = t.db.QueryRow(query + " RETURNING ID")
		if err != nil { return nil, err }
		t.db.ClearFilter()
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
	}
	if t.db.Driver == conn.MySQLDriver {
		stmt, err := t.db.Prepare(query)
		if err != nil { return t.DBError(nil, err) }
		res, err := stmt.Exec()
		if err != nil { return nil, err }
		id, err = res.LastInsertId()
		if err != nil { return t.DBError(nil, err) }
		t.db.ClearFilter()
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
	}
	result, err = t.db.SelectResults(t.Table.Name)
	t.Results = result
	if t.SpecializedService != nil && len(t.Results) > 0 {
		t.SpecializedService.WriteRowAutomation(t.Results[0], t.Table.Name)
	}
	return t.Results, err
}

func (t *TableRowInfo) Update(restriction... string) ([]map[string]interface{}, error) {
	var err error
	if id, ok := t.Record["id"]; (!ok || id == "" || fmt.Sprintf("%v", id) == "0") { return nil, errors.New("verification failed on id.") }
	if err = t.setupFilter(false, true, restriction...); err != nil { return nil, err }
	stack := ""
	restr := t.db.SQLRestriction
	for key, element := range t.Record {
		if (strings.Contains(key, "_id") || key == "id") && fmt.Sprintf("%v", element) == "0" { continue }
		if key == "id" && fmt.Sprintf("%v", element) != "0" && fmt.Sprintf("%v", element) != "" { restr = "id=" + fmt.Sprintf("%v", element) + " "; continue } 
		t.EmptyCol.Name = t.Name
		typ, ok := t.EmptyCol.Verify(key)
		realType := strings.Split(typ, ":")[0]
		//isNull := len(strings.Split(typ, ":")) > 1 && strings.Split(typ, ":")[1] == "nullable"
		if ok && conn.FormatForSQL(realType, element) != "NULL" && conn.FormatForSQL(realType, element) != "" { 
			stack += key + "=" + conn.FormatForSQL(realType, element) + "," 
		}
	}
	stack = conn.RemoveLastChar(stack)
	query := ("UPDATE " + t.Table.Name + " SET " + stack) // REMEMBER id is a restriction !
	if restr != "" { query += " WHERE " + restr
	} else if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction } 
	if stack != "" { 
		if err := t.db.Query(query); err != nil { return t.DBError(nil, err) }
	}
	t.db.ClearFilter()
	if restr != "" { t.db.SQLRestriction = restr }
	if t.Results, err = t.db.SelectResults(t.Table.Name); err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil { t.SpecializedService.UpdateRowAutomation(t.Results, t.Record) }
	return t.Results, nil
}
func (t *TableRowInfo) Delete(restriction... string) ([]map[string]interface{}, error) {
	var err error
	if err = t.setupFilter(true, true, restriction...); err != nil { return t.DBError(nil, err) }
	t.Results, err = t.db.SelectResults(t.Table.Name)
	if t.db.SQLRestriction == "" { return t.DBError(nil, errors.New("no restriction can't delete all")) }
	if err = t.db.Query("DELETE FROM " + t.Table.Name + " WHERE " + t.db.SQLRestriction); err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil { t.SpecializedService.DeleteRowAutomation(t.Results, t.Table.Name) }
	return t.Results, nil
}

func (t *TableRowInfo) setupFilter(reverse bool, verify bool, restriction... string) error {
	if t.SpecializedService != nil && verify {
		r, ok, forceChange := t.SpecializedService.VerifyRowAutomation(t.Record, t.Name)
		if !ok { return errors.New("verification failed.") }
		if forceChange { t.Record = r }
	}
	t.db.ClearFilter()
	if len(restriction) > 0 && reverse { 
		for _, r := range restriction { 
			if strings.TrimSpace(r) == "" { continue }
			if len(r) > 5 && r[:5] == " AND " { r = r[5:] }
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + r
			} else { t.db.SQLRestriction = r }
		}
	}
	if t.SpecializedService != nil {
		restr, view, order, limit := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if restr != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + restr
			} else { t.db.SQLRestriction = restr }
		}
		if view != "" { t.db.SQLView = view }
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	if len(restriction) > 0 && !reverse { 
		for _, r := range restriction { 
			if strings.TrimSpace(r) == "" { continue }
			if len(r) > 5 && r[:5] == " AND " { r = r[5:] }
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + r 
			} else { t.db.SQLRestriction = r }
		}
	}
	if len(t.db.SQLRestriction) > 5 && t.db.SQLRestriction[:5] == " AND " { t.db.SQLRestriction = t.db.SQLRestriction[5:] }
	return nil
}
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
	Verified  	        bool
	InfraService
}

func (t *TableRowInfo) Template(restriction... string) (interface{}, error) { return t.Get(restriction...) }

func (t *TableRowInfo) Verify(name string) (string, bool) {
	res, err := t.Get("id=" + name)
	return name, err == nil && len(res) > 0
}

func (t *TableRowInfo) Count(restriction... string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	if t.SpecializedService != nil {
		restriction, _, order, limit := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	var err error; var count int64
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
	t.db.ClearFilter()
	if t.SpecializedService != nil {
		restriction, view, order, limit:= t.SpecializedService.ConfigureFilter(t.Table.Name)
		if view != "" { t.db.SQLView = view }
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	d, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = d
	return t.Results, nil
}

func (t *TableRowInfo) Create() ([]map[string]interface{}, error) {
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
		if (strings.Contains(key, "_id") || key == "id") && fmt.Sprintf("%v", element) == "0" { continue }
 		t.EmptyCol.Name = t.Name
		typ, _ := t.EmptyCol.Verify(key) 
		realType := strings.Split(typ, ":")[0]
		if realType != "" { 
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
	t.db.ClearFilter()
	if t.SpecializedService != nil {
		r, ok, forceChange := t.SpecializedService.VerifyRowAutomation(t.Record, t.Name)
		if !ok { return nil, errors.New("verification failed.") }
		if forceChange { t.Record = r }
	}
	if t.SpecializedService != nil {
		restriction, view, order, limit := t.SpecializedService.ConfigureFilter(t.Table.Name) 
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction  + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view } 
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	stack := ""
	restr := t.db.SQLRestriction
	if id, ok := t.Record["id"]; (!ok || fmt.Sprintf("%v", id) == "0") && !strings.Contains(t.db.SQLRestriction, "id=") { return t.Create() }
	for key, element := range t.Record {
		if (strings.Contains(key, "_id") || key == "id") && fmt.Sprintf("%v", element) == "0" { continue }
		if key == "id" { 
			if fmt.Sprintf("%v", element) != "0" { restr = "id=" + fmt.Sprintf("%v", element) + " " }
			continue 
		}
		if t.Verified {
			typ, ok := t.EmptyCol.Verify(key)
			realType := strings.Split(typ, ":")[0]
			isNull := len(strings.Split(typ, ":")) > 1 && strings.Split(typ, ":")[1] == "nullable"
			if ok { 
				if (!isNull && conn.FormatForSQL(realType, element) == "NULL") { continue }
				stack += key + "=" + conn.FormatForSQL(realType, element) + "," 
			}
		} else { stack += " " + key + "=" + fmt.Sprintf("%v", element) + "," }
	}
	stack = conn.RemoveLastChar(stack)
	query := ("UPDATE " + t.Table.Name + " SET " + stack) // REMEMBER id is a restriction !
	if restr != "" { query += " WHERE " + restr
	} else if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction } 
	if stack != "" {
		if t.db.Query(query) != nil { return t.DBError(nil, errors.New("nothing to update")) }
	}
	t.db.ClearFilter()
	if restr != "" { t.db.SQLRestriction = restr }
	result, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = result
	if t.SpecializedService != nil { t.SpecializedService.UpdateRowAutomation(t.Results, t.Record) }
	return t.Results, nil
}

func (t *TableRowInfo) CreateOrUpdate(restriction... string) ([]map[string]interface{}, error) { return t.Update(restriction...) }

func (t *TableRowInfo) Delete(restriction... string) ([]map[string]interface{}, error) {
	t.db.ClearFilter()
	if t.SpecializedService != nil {
		_, ok, _ := t.SpecializedService.VerifyRowAutomation(t.Record, t.Name)
		if !ok { return nil, errors.New("verification failed.") }
		restriction, view, order, limit := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction  + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view }
		if order != "" { t.db.SQLOrder = order }
		if limit != "" { t.db.SQLLimit = limit }
	}
	res, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = res
	query := ("DELETE FROM " + t.Table.Name)
	if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction }
	err = t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil { t.SpecializedService.DeleteRowAutomation(t.Results, t.Table.Name) }
	return t.Results, nil
}
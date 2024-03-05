package service

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"encoding/json"
	tool "sqldb-ws/lib"
	_ "github.com/go-sql-driver/mysql"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TableRowInfo struct {
	SpecializedService  tool.SpecializedService `json:"-"`
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

func (t *TableRowInfo) Count(restriction... string) (tool.Results, error) {
	t.db.ToFilter(t.Name, t.Params, restriction...)
	if t.SpecializedService != nil {
		restriction, _ := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
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
	t.Results = append(t.Results, tool.Record{ "count" : count, })
	return t.Results, nil
}

func (t *TableRowInfo) Get(restriction... string) (tool.Results, error) {
	t.db.ToFilter(t.Name, t.Params, restriction...)
	if t.SpecializedService != nil {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if view != "" { t.db.SQLView = view }
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
	}
	d, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = d
	return t.Results, nil
}

func (t *TableRowInfo) Create() (tool.Results, error) {
	var id int64
	var err error
	var result tool.Results
	columns := ""
	values := ""
	if t.SpecializedService != nil {
		r, ok, forceChange := t.SpecializedService.VerifyRowAutomation(t.Record, true)
		if !ok { return nil, errors.New("verification failed.") }
		if forceChange { t.Record = r }
	}
	if len(t.Record) > 0 {
		v := Validator[map[string]interface{}]()
		rec, err := v.ValidateSchema(t.Record, t.Table, false)
		t.Record = rec
		if err != nil { return nil, errors.New("Not a proper struct to create a row " + err.Error()) }
	} else if !entities.IsRootDB(t.Table.Name) {
		emptyRec, err := t.Table.EmptyRecord()
		if err != nil { return nil, errors.New("Empty record got a problem : " + err.Error()) }
		t.Record = emptyRec
	} else { return nil, errors.New("Empty is not a proper struct to create a row ") }
	for key, element := range t.Record {
		columns += key + ","
		typ := ""
		if t.Verified { 
			typ, _ = t.EmptyCol.Verify(key) 
			realType := strings.Split(typ, ":")[0]
			values += conn.FormatForSQL(realType, element) + ","
		} else { values += fmt.Sprintf("%v", element) + "," }
	}
	query := "INSERT INTO " + t.Table.Name + "(" + conn.RemoveLastChar(columns) + ") VALUES (" + conn.RemoveLastChar(values) + ")"
	if t.db.Driver == conn.PostgresDriver { 
		id, err = t.db.QueryRow(query + " RETURNING ID")
		if err != nil { return nil, err }
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
	}
	if t.db.Driver == conn.MySQLDriver {
		stmt, err := t.db.Prepare(query)
		if err != nil { return t.DBError(nil, err) }
		res, err := stmt.Exec()
		if err != nil { return nil, err }
		id, err = res.LastInsertId()
		if err != nil { return t.DBError(nil, err) }
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
	}
	result, err = t.db.SelectResults(t.Table.Name)
	t.Results = result
	if t.SpecializedService != nil && len(t.Results) > 0 {
		t.SpecializedService.WriteRowAutomation(t.Results[0], t.Table.Name)
	}
	return t.Results, err
}

func (t *TableRowInfo) Update(restriction... string) (tool.Results, error) {
	if t.SpecializedService != nil {
		r, ok, forceChange := t.SpecializedService.VerifyRowAutomation(t.Record, false)
		if !ok { return nil, errors.New("verification failed.") }
		if forceChange { t.Record = r }
	}
	v := Validator[map[string]interface{}]()
	rec, err := v.ValidateSchema(t.Record, t.Table, true)
	t.Record = rec
	if err != nil { return nil, errors.New("Not a proper struct to update a row") }
	t.db.ToFilter(t.Name, t.Params, restriction...)
	if t.SpecializedService != nil {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name) 
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction  + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view } 
	}
	stack := ""
	restr := ""
	for key, element := range t.Record {
		if key != tool.SpecialIDParam { 
			if t.Verified {
				typ, ok := t.EmptyCol.Verify(key)
				realType := strings.Split(typ, ":")[0]
				isNull := strings.Split(typ, ":")[1] == "nullable"
				if ok { 
					if (!isNull && conn.FormatForSQL(realType, element) == "NULL") || (!isNull && strings.Contains(strings.ToLower(realType), "bool")) { continue }
					stack = stack + key + "=" + conn.FormatForSQL(realType, element) + "," 
				}
			} else { stack = stack + " " + key + "=" + fmt.Sprintf("%v", element) + "," }
		} else { restr = "id=" + fmt.Sprintf("%v", element) + " " }
	}
	stack = conn.RemoveLastChar(stack)
	query := ("UPDATE " + t.Table.Name + " SET " + stack) // REMEMBER id is a restriction !
	if restr != "" { query += " WHERE " + restr 
    } else { return t.Create() }
	err = t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	t.db.SQLRestriction = restr
	result, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = result
	if t.SpecializedService != nil {
		t.SpecializedService.UpdateRowAutomation(t.Results, t.Record)
	}
	return t.Results, nil
}

func (t *TableRowInfo) CreateOrUpdate(restriction... string) (tool.Results, error) {
	_, ok := t.Params[tool.SpecialIDParam]
	if ok == false && t.Method != tool.UPDATE { return t.Create() 
	} else { return t.Update(restriction...) }
}

func (t *TableRowInfo) Delete(restriction... string) (tool.Results, error) {
	t.db.ToFilter(t.Name, t.Params, restriction...)
	if t.SpecializedService != nil {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction = t.db.SQLRestriction + " AND (" + restriction  + ")"
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view }
	}
	res, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = res
	query := ("DELETE FROM " + t.Table.Name)
	if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction }
	fmt.Printf("QUERY %v \n", query)
	err = t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil {
		t.SpecializedService.DeleteRowAutomation(t.Results, t.Table.Name)
	}
	return t.Results, nil
}

func (t *TableRowInfo) Import(filename string, restriction... string) (tool.Results, error)  {
	var jsonSource []TableRowInfo
	byteValue, _ := os.ReadFile(filename)
	err := json.Unmarshal([]byte(byteValue), &jsonSource)
	if err != nil { return t.DBError(nil, err) }
	for _, row := range jsonSource {
		row.db = t.db
		if t.Method == tool.DELETE { row.Delete(restriction...) 
		} else { row.Create() }
	}
	return t.Results, nil
}
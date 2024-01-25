package service

import (
	"os"
	"fmt"
	"strings"
	"errors"
	"encoding/json"
	tool "sqldb-ws/lib"
	_ "github.com/go-sql-driver/mysql"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TableRowInfo struct {
	SpecializedService  tool.SpecializedService `json:"-"`
	Table				*TableInfo
	EmptyCol            *TableColumnInfo
	Verified  	        bool
	AdminView           bool
	InfraService
}

func (t *TableRowInfo) Template() (interface{}, error) { return t.Get() }

func (t *TableRowInfo) Verify(name string) (string, bool) {
	t.db.SQLRestriction = "id=" + name
	res, err := t.Get()
	return name, err == nil && len(res) > 0
}

func (t *TableRowInfo) Get() (tool.Results, error) {
	t.db = ToFilter(t.Table.Name, t.Params, t.db)
	if t.SpecializedService != nil && ! t.AdminView {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name, t.Params)
		if view != "" { t.db.SQLView = view }
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + restriction 
		    } else { t.db.SQLRestriction = restriction }
		}
	}
	d, err := t.db.SelectResults(t.Table.Name)
	t.Results = d
	if err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil && t.PostTreatment {
		t.Results = t.SpecializedService.PostTreatment(t.Results)
	}
	return t.Results, nil
}

func (t *TableRowInfo) Create() (tool.Results, error) {
	var id int64
	var err error
	var result tool.Results
	columns := ""
	values := ""
	if t.SpecializedService != nil {
		if _, ok := t.SpecializedService.VerifyRowAutomation(t.Record, true); !ok { return nil, errors.New("verification failed.") }
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
			values += conn.FormatForSQL(typ, element) + ","
		} else {
			values += fmt.Sprintf("%v", element) + ","
		}
	}
	query := "INSERT INTO " + t.Table.Name + "(" + conn.RemoveLastChar(columns) + ") VALUES (" + conn.RemoveLastChar(values) + ")"
	if t.db.Driver == conn.PostgresDriver { 
		id, err = t.db.QueryRow(query)
		if err != nil { return t.DBError(nil, err) }
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
		if err != nil { return t.DBError(nil, err) }
	}
	if t.db.Driver == conn.MySQLDriver {
		stmt, err := t.db.Prepare(query)
		if err != nil { return t.DBError(nil, err) }
		res, err := stmt.Exec()
		if err != nil { return t.DBError(nil, err) }
		id, err = res.LastInsertId()
		t.db.SQLRestriction = fmt.Sprintf("id=%d", id)
		if err != nil { return t.DBError(nil, err) }
		if err != nil { return t.DBError(nil, err) }
	}

	if t.SpecializedService != nil && !t.AdminView {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name, t.Params)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + restriction 
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view }
	}
	result, err = t.db.SelectResults(t.Table.Name)
	t.Results = result
	if t.SpecializedService != nil {
		t.SpecializedService.WriteRowAutomation(t.Record)
		if t.PostTreatment { t.Results = t.SpecializedService.PostTreatment(t.Results) }
	}
	return t.Results, nil
}

func (t *TableRowInfo) Update() (tool.Results, error) {
	v := Validator[map[string]interface{}]()
	if t.SpecializedService != nil {
		r, _ := t.SpecializedService.VerifyRowAutomation(t.Record, false) 
		t.Record = r
	}
	rec, err := v.ValidateSchema(t.Record, t.Table, true)
	t.Record = rec
	if err != nil { return nil, errors.New("Not a proper struct to update a row") }
	t.db = ToFilter(t.Table.Name, t.Params, t.db)
	stack := ""
	filter := ""
	for key, element := range t.Record {
		if key != tool.SpecialIDParam { 
			if t.PermService != nil && len(t.PermService.WarningUpdateField) > 0 {
				found := false
				for _, w := range t.PermService.WarningUpdateField { 
					if w == key { found = true; break }
				}
				if found {
					t.Params[key]="NULL"
					resp, _ := t.db.SelectResults(t.Table.Name)
					if len(resp) == 0 { continue }
				}					
			} 
			if t.Verified {
				typ, ok := t.EmptyCol.Verify(key)
				if ok { 
					stack = stack + key + "=" + conn.FormatForSQL(typ, element) + "," 
					filter += key + "=" + conn.FormatForSQL(typ, element) + " and " 
				}
			} else { 
				stack = stack + " " + key + "=" + fmt.Sprintf("%v", element) + "," 
				filter += key + "=" + fmt.Sprintf("%v", element) + " and " 
			}
		} else if !strings.Contains(t.db.SQLRestriction, "id=") { t.db.SQLRestriction += "id=" + fmt.Sprintf("%v", int64(element.(float64))) + " " }
	}
	stack = conn.RemoveLastChar(stack)
	query := ("UPDATE " + t.Table.Name + " SET " + stack) // REMEMBER id is a restriction !
	if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction }
	err = t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	if len(t.db.SQLRestriction) > 0 { 
		if (len(filter) > 0) {
			t.db.SQLRestriction += "and " + filter[:len(filter) - 4]
		}
    } else { if (len(filter) > 0) { t.db.SQLRestriction = filter[:len(filter) - 4] }  }
	if t.SpecializedService != nil && !t.AdminView {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name, t.Params)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + restriction 
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view } 
	}
	res, err := t.db.SelectResults(t.Table.Name)
	t.Results = res
	if err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil {
		t.SpecializedService.UpdateRowAutomation(res, t.Record) 
		if t.PostTreatment { t.Results = t.SpecializedService.PostTreatment(t.Results) }
	}
	return t.Results, nil
}

func (t *TableRowInfo) CreateOrUpdate() (tool.Results, error) {
	_, ok := t.Params[tool.SpecialIDParam]
	if ok == false && t.Method != tool.UPDATE { return t.Create() 
	} else { return t.Update() }
}

func (t *TableRowInfo) Delete() (tool.Results, error) {
	t.db = ToFilter(t.Table.Name, t.Params, t.db)
	if t.SpecializedService != nil && !t.AdminView {
		restriction, view := t.SpecializedService.ConfigureFilter(t.Table.Name, t.Params)
		if restriction != "" { 
			if len(t.db.SQLRestriction) > 0 { t.db.SQLRestriction += " AND " + restriction 
		    } else { t.db.SQLRestriction = restriction }
		}
		if view != "" { t.db.SQLView = view }
	}
	res, err := t.db.SelectResults(t.Table.Name)
	if err != nil { return t.DBError(nil, err) }
	t.Results = res
	query := ("DELETE FROM " + t.Table.Name)
	if t.db.SQLRestriction != "" { query += " WHERE " + t.db.SQLRestriction }
	err = t.db.Query(query)
	if err != nil { return t.DBError(nil, err) }
	if t.SpecializedService != nil {
		t.SpecializedService.DeleteRowAutomation(t.Results)
		if t.PostTreatment { t.Results = t.SpecializedService.PostTreatment(t.Results) }
	}
	return t.Results, nil
}

func (t *TableRowInfo) Import(filename string) (tool.Results, error)  {
	var jsonSource []TableRowInfo
	byteValue, _ := os.ReadFile(filename)
	err := json.Unmarshal([]byte(byteValue), &jsonSource)
	if err != nil { return t.DBError(nil, err) }
	for _, row := range jsonSource {
		row.db = t.db
		if t.Method == tool.DELETE { row.Delete() 
		} else { row.Create() }
	}
	return t.Results, nil
}

func (t *TableRowInfo) Link() (tool.Results, error) {
	if _, ok := t.Params[tool.RootToTableParam]; !ok { return nil, errors.New("no destination table") }
	otherName := t.Params[tool.RootToTableParam]
	v := Validator[entities.LinkEntity]()
	v.data = entities.LinkEntity{}
	te, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to create a table - expect <LinkEntity> Scheme " + err.Error()) }
	if _, ok := t.EmptyCol.Verify(otherName + "_id"); ok && te.Anchor == "" {
		// should verify record from_id to_id
		res, err := t.link(te, otherName, false)
		if err != nil { t.Results = append(t.Results, res...) }
	} else {
		// here FIND LINK TABLE
		schemas, err := t.Table.schema(tool.ReservedParam)
		if err != nil { return nil, errors.New("problem on schema")}
		for _, scheme := range schemas {
			_, findRoot := scheme.AssColumns[t.Name + "_id"] 
			_, findOther := scheme.AssColumns[otherName  + "_id"] 
			if findRoot && findOther && strings.Contains(scheme.Name, te.Anchor) {
				t.EmptyCol.Name = scheme.Name
				t.Table.Name = t.EmptyCol.Name
				res, err := t.link(te, otherName, false)
				if err == nil { t.Results = append(t.Results, res...) }
			}
		}
	}
	return t.Results, nil
}
func (t *TableRowInfo) link(te *entities.LinkEntity, otherName string, nullable bool) (tool.Results, error)  {
	t.Record = tool.Record{ otherName + "_id" : te.To, t.Name + "_id" : te.From }
	if te.Columns != nil && !nullable {
		for col, val := range te.Columns {
			if _, ok := t.EmptyCol.Verify(col); ok && val != "" { t.Record[col]=val }
		}
	}
	if len(t.Record) == 0 { return nil, errors.New("no data to set or create")}
	if !nullable { return t.CreateOrUpdate() 
	} else { return t.Delete()  }
}

func (t *TableRowInfo) UnLink() (tool.Results, error) {
	if _, ok := t.Params[tool.RootToTableParam]; !ok { return nil, errors.New("no destination table") }
	otherName := t.Params[tool.RootToTableParam]
	v := Validator[entities.LinkEntity]()
	v.data = entities.LinkEntity{}
	te, err := v.ValidateStruct(t.Record)
	if err != nil { return nil, errors.New("Not a proper struct to create a table - expect <LinkEntity> Scheme " + err.Error()) }
	if _, ok := t.EmptyCol.Verify(otherName + "_id"); ok {
		res, err := t.link(te, otherName, true)
		if err != nil { t.Results = append(t.Results, res...) }
	} else { 
		schema, err := t.Table.schema(tool.ReservedParam)
		if err != nil { return nil, errors.New("problem on schema")}
		for _, scheme := range schema {
			_, findRoot := scheme.AssColumns[t.Name + "_id"] 
			_, findOther := scheme.AssColumns[otherName + "_id"] 
			if findRoot && findOther && strings.Contains(scheme.Name, te.Anchor) {
				t.EmptyCol.Name = scheme.Name
				t.Table.Name = t.EmptyCol.Name
				res, err := t.link(te, otherName, true)
				if err != nil { t.Results = append(t.Results, res...) }
			}
		}
	}
	return t.Results, nil
}
package service

import (
	"errors"
	"fmt"
	conn "sqldb-ws/infrastructure/connector"
	"strings"
)

type TableRowService struct {
	Table    *TableService
	EmptyCol *TableColumnService
	InfraService
}

func (t *TableRowService) Template(restriction ...string) (interface{}, error) {
	return t.Get(restriction...)
}

func (t *TableRowService) Verify(name string) (string, bool) {
	res, err := t.Get("id=" + name)
	return name, err == nil && len(res) > 0
}

func (t *TableRowService) Math(algo string, restriction ...string) ([]map[string]interface{}, error) {
	if err := t.setupFilter(false, restriction...); err != nil {
		return nil, err
	}
	res, err := t.DB.MathQuery(algo, t.Table.Name)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	t.Results = append(t.Results, map[string]interface{}{"result": res[0]["result"]})
	return t.Results, nil
}

func (t *TableRowService) Get(restriction ...string) ([]map[string]interface{}, error) {
	var err error
	if err = t.setupFilter(false, restriction...); err != nil {
		return nil, err
	}
	if t.Results, err = t.DB.SelectQueryWithRestriction(
		t.Table.Name, map[string]interface{}{}, false); err != nil {
		return t.DBError(nil, err)
	}
	return t.Results, nil
}

func (t *TableRowService) Create() ([]map[string]interface{}, error) {
	if len(t.Record) == 0 {
		return nil, errors.New("no data to insert")
	}
	t.DB.ClearQueryFilter()
	var id int64
	var err error
	if r, err, forceChange := t.SpecializedService.VerifyDataIntegrity(t.Record, t.Name); err != nil {
		return nil, err
	} else if forceChange {
		t.Record = r
	}
	var columns, values string
	t.EmptyCol.Name = t.Name
	for key, element := range t.Record {
		_, columns, values = t.DB.BuildUpdateQuery(key, element, "", columns, values, t.EmptyCol.Verify)
	}
	for _, query := range t.DB.BuildCreateQueries(t.Table.Name, values, columns, "") {
		if t.DB.Driver == conn.PostgresDriver {
			id, err = t.DB.QueryRow(query)
			if err != nil {
				return t.DBError(nil, err)
			}
		} else if t.DB.Driver == conn.MySQLDriver {
			if stmt, err := t.DB.Prepare(query); err != nil {
				return t.DBError(nil, err)
			} else if res, err := stmt.Exec(); err != nil {
				return t.DBError(nil, err)
			} else if id, err = res.LastInsertId(); err != nil {
				return t.DBError(nil, err)
			}
		}
	}
	t.DB.ClearQueryFilter().ApplyQueryFilters(fmt.Sprintf("id=%d", id), "", "", "")

	if t.Results, err = t.DB.SelectQueryWithRestriction(
		t.Table.Name, map[string]interface{}{}, false); len(t.Results) > 0 {
		t.SpecializedService.SpecializedCreateRow(t.Results[0], t.Table.Name)
	}
	return t.Results, err
}

func (t *TableRowService) Update(restriction ...string) ([]map[string]interface{}, error) {
	var err error
	if err = t.setupFilter(true, restriction...); err != nil {
		return nil, err
	}
	t.EmptyCol.Name = t.Name
	if query, err := t.DB.BuildUpdateRowQuery(t.Table.Name, t.Record, t.EmptyCol.Verify); err == nil {
		if err := t.DB.Query(query); err != nil {
			return t.DBError(nil, err)
		}
		t.DB.ClearQueryFilter().ApplyQueryFilters("", "", "", "")
		if restr := strings.Split(query, "WHERE"); len(restr) > 1 {
			t.DB.ApplyQueryFilters(restr[len(restr)-1], "", "", "")
		}
	} else {
		return t.DBError(nil, err)
	}
	if t.Results, err = t.DB.SelectQueryWithRestriction(
		t.Table.Name, map[string]interface{}{}, false); err != nil {
		return t.DBError(nil, err)
	}
	t.SpecializedService.SpecializedUpdateRow(t.Results, t.Record)
	return t.Results, nil
}
func (t *TableRowService) Delete(restriction ...string) ([]map[string]interface{}, error) {
	var err error
	if t.Results, err = t.Get(restriction...); err == nil {
		if t.DB.SQLRestriction == "" {
			return t.DBError(nil, errors.New("no restriction can't delete all"))
		} else if err = t.DB.DeleteQuery(t.Table.Name, ""); err != nil {
			return t.DBError(nil, err)
		}
		t.SpecializedService.SpecializedDeleteRow(t.Results, t.Table.Name)
	}
	return t.Results, err
}

func (t *TableRowService) setupFilter(verify bool, restriction ...string) error {
	if verify {
		if r, err, forceChange := t.SpecializedService.VerifyDataIntegrity(t.Record, t.Name); err != nil {
			return err
		} else if forceChange {
			t.Record = r
		}
	}
	restr, order, limit, view := t.SpecializedService.GenerateQueryFilter(t.Table.Name, restriction...)
	t.DB.ClearQueryFilter().ApplyQueryFilters(restr, order, limit, view)
	if len(t.DB.SQLRestriction) > 5 && t.DB.SQLRestriction[len(t.DB.SQLRestriction)-5:] == " AND " {
		t.DB.SQLRestriction = t.DB.SQLRestriction[:len(t.DB.SQLRestriction)-5]
	}
	return nil
}

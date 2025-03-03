package service

import (
	"errors"
	"fmt"
	"strings"
)

// Table is a table structure description
type TableColumnService struct {
	Views string
	InfraService
}

func (t *TableColumnService) Template(restriction ...string) (interface{}, error) {
	return t.Get(restriction...)
}

func (t *TableColumnService) Math(algo string, restriction ...string) ([]map[string]interface{}, error) {
	restr, order, limit, _ := t.SpecializedService.GenerateQueryFilter(t.Name, strings.Join(restriction, " AND "))
	t.DB.ApplyQueryFilters(restr, order, limit, t.Views)
	res, err := t.DB.MathQuery(algo, t.Name)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	t.Results = append(t.Results, map[string]interface{}{"result": res[0]["result"]})
	return t.Results, nil
}

func (t *TableColumnService) Get(restriction ...string) ([]map[string]interface{}, error) {
	var err error
	restr, order, limit, _ := t.SpecializedService.GenerateQueryFilter(t.Name, restriction...)
	t.DB.ApplyQueryFilters(restr, order, limit, t.Views)
	if t.Results, err = t.DB.SelectQueryWithRestriction(t.Name, map[string]interface{}{}, false); err != nil {
		return t.DBError(nil, err)
	}
	return t.Results, nil
}

func (t *TableColumnService) Verify(name string) (string, bool) {
	var typ string
	if cols, _, err := RetrieveTable(t.Name, t.DB); err == nil {
		if cols[name].Null {
			typ = cols[name].Type + ":nullable"
		} else {
			typ = cols[name].Type + ":required"
		}
	}
	return typ, typ != ""
}

func (t *TableColumnService) Create() ([]map[string]interface{}, error) {
	queries := t.DB.ClearQueryFilter().BuildCreateQueries(t.Name, "",
		fmt.Sprintf("%v", t.Record["name"]),
		fmt.Sprintf("%v", t.Record["type"]))
	for i, query := range queries {
		if err := t.DB.Query(query); err != nil && i != 0 {
			return t.DBError(nil, err)
		} else if err = t.update(); err != nil {
			return t.DBError(nil, err)
		}
	}
	if len(queries) > 0 {
		t.Views = fmt.Sprintf("%v", t.Record["name"])
		return t.Get()
	}
	return nil, errors.New("no query to execute")
}

func (t *TableColumnService) Update(restr ...string) ([]map[string]interface{}, error) {
	t.DB.ClearQueryFilter()
	typ := fmt.Sprintf("%v", t.Record["type"])
	name := fmt.Sprintf("%v", t.Record["name"])
	if typ == "" || typ == "<nil>" || name == "" || name == "<nil>" {
		return nil, errors.New("missing one of the needed value type & name")
	}
	if err := t.update(); err != nil {
		return t.DBError(nil, err)
	}
	if strings.TrimSpace(name) != "" && !strings.Contains(t.Name, "db") {
		col := strings.Split(t.Views, ",")[0]
		query := "ALTER TABLE " + t.Name + " RENAME COLUMN " + col + " TO " + name + ";" // TODO
		err := t.DB.Query(query)
		if err != nil {
			return t.DBError(nil, err)
		}
	}
	return t.Get()
}

func (t *TableColumnService) Delete(restriction ...string) ([]map[string]interface{}, error) {
	if strings.Contains(t.Name, "db") { // protect root db columns
		return nil, errors.New("can't delete protected root db columns")
	}
	for _, col := range strings.Split(t.Views, ",") {
		if err := t.DB.ClearQueryFilter().DeleteQuery(t.Name, col); err != nil {
			return t.DBError(nil, err)
		}
		t.Results = append(t.Results, map[string]interface{}{"name": col})
	}
	return t.Results, nil
}

func (t *TableColumnService) update() error {
	if queries, err := t.DB.BuildUpdateColumnQueries(t.Name, t.Record, nil); err != nil {
		return err
	} else {
		for _, query := range queries {
			fmt.Println(t.DB.Query(query))
		}
	}
	return nil
}

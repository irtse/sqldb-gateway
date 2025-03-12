package service

import (
	"encoding/json"
	"errors"
	"fmt"
	conn "sqldb-ws/infrastructure/connector"
	"sqldb-ws/infrastructure/models"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Table is a table structure description
type TableService struct {
	InfraService
}

// Generate an Empty TableService Service
func NewTableService(database *conn.Database, admin bool, user string, name string) *TableService {
	table := &TableService{}
	table.Name = name
	table.SuperAdmin = admin
	table.User = user
	table.DB = database
	if table.SpecializedService == nil {
		table.SpecializedService = &InfraSpecializedService{}
	}
	return table
}

func (t *TableService) NewTableRowService(specializedService InfraSpecializedServiceItf) *TableRowService {
	row := &TableRowService{
		Table: NewTableService(t.DB, t.SuperAdmin, t.User, t.Name),
		EmptyCol: &TableColumnService{InfraService: InfraService{
			DB:                 t.DB,
			SpecializedService: &InfraSpecializedService{},
		}},
		InfraService: InfraService{
			DB:                 t.DB,
			NoLog:              t.NoLog,
			SpecializedService: specializedService,
		},
	}
	row.Fill(t.Name, t.SuperAdmin, t.User)
	if row.SpecializedService == nil {
		row.SpecializedService = &InfraSpecializedService{}
	}
	return row
}

func (t *TableService) NewTableColumnService(specializedService InfraSpecializedServiceItf, views string) *TableColumnService {
	col := &TableColumnService{
		Views: views,
		InfraService: InfraService{
			DB: t.DB, NoLog: t.NoLog, SpecializedService: specializedService},
	}
	col.Fill(t.Name, t.SuperAdmin, t.User)
	if col.SpecializedService == nil {
		col.SpecializedService = &InfraSpecializedService{}
	}
	return col
}

func (t *TableService) Template(restriction ...string) (interface{}, error) {
	if res, err := t.getTableSchema(t.Name); err != nil {
		return nil, err
	} else {
		return struct { // TODO FORMALIZED
			Tbl []models.TableEntity
			Lnk []models.Link
		}{res, models.BuildLinks(res)}, nil
	}
}

// Math is a method to perform math operations on a table
func (t *TableService) Math(algo string, restriction ...string) ([]map[string]interface{}, error) {
	return nil, errors.New("not implemented")
}

func (t *TableService) Get(restriction ...string) ([]map[string]interface{}, error) {
	schema, err := t.getTableSchema(t.Name)
	if err != nil {
		return t.DBError(nil, err)
	}
	t.Results = []map[string]interface{}{} // clear any previous results
	for _, s := range schema {
		t.Results = append(t.Results, map[string]interface{}{
			"name":    s.Name,
			"columns": s.Cols,
		})
	}
	return t.Results, nil
}

func (t *TableService) Delete(restriction ...string) ([]map[string]interface{}, error) {
	for _, dropQuery := range t.DB.ClearQueryFilter().BuildDropTableQueries(t.Name) { // drop all for the table
		if err := t.DB.Query(dropQuery); err != nil {
			return t.DBError(nil, err)
		}
	}
	t.Results = append(t.Results, map[string]interface{}{"name": t.Name}) // return the name of the table as a result
	return t.Results, nil
}

func (t *TableService) Verify(name string) (string, bool) {
	schema, err := t.getTableSchema(name)
	return name, len(schema) == 0 || err != nil
}

func (t *TableService) Create(record map[string]interface{}) ([]map[string]interface{}, error) {
	return t.Write(record, false)
}

func (t *TableService) Update(record map[string]interface{}, restr ...string) ([]map[string]interface{}, error) {
	return t.Write(record, true)
}

func (t *TableService) Write(record map[string]interface{}, isUpdate bool) ([]map[string]interface{}, error) {
	t.DB.ClearQueryFilter()
	t.Name = fmt.Sprintf("%v", record["name"])
	if t.Name == "" || t.Name == "<nil>" {
		return nil, errors.New("missing one of the needed value type & name")
	}
	if !isUpdate {
		t.DB.CreateTableQuery(t.Name)
	}
	if fields, ok := record["fields"]; ok {
		for _, rowtype := range fields.([]interface{}) {
			tc := t.NewTableColumnService(t.SpecializedService, "")
			if isUpdate {
				tc.Update(rowtype.(map[string]interface{}))
			} else {
				tc.Create(rowtype.(map[string]interface{}))
			}
		}
	}
	return t.Get()
}

func (t *TableService) getTableSchema(name string) ([]models.TableEntity, error) {
	schema := []models.TableEntity{}
	tables, err := t.DB.ClearQueryFilter().ListTableQuery()
	if err != nil {
		return nil, err
	}
	for _, row := range tables {
		if element, ok := row["name"]; ok && !(name != "all" && name != fmt.Sprintf("%v", element)) {
			if mapped, cols, err := RetrieveTable(fmt.Sprintf("%v", element), t.DB); err == nil {
				schema = append(schema, models.TableEntity{
					Name:       fmt.Sprintf("%v", element),
					AssColumns: mapped,
					Cols:       cols})
			}
		}
	}
	return schema, nil
}

func RetrieveTable(name string, db *conn.Database) (map[string]models.TableColumnEntity, []string, error) {
	cols, err := db.SchemaQuery(name)
	if err != nil {
		return nil, nil, err
	}
	mapped := make(map[string]models.TableColumnEntity)
	var columns []string
	for _, row := range cols {
		var tableCol models.TableColumnEntity
		b, _ := json.Marshal(row)
		json.Unmarshal(b, &tableCol)
		tableCol.Null = row["null"] == "YES"
		if defaultVal, ok := tableCol.Default.(string); ok && strings.Contains(defaultVal, "NULL") {
			tableCol.Default = nil
		}
		mapped[tableCol.Name] = tableCol
		columns = append(columns, tableCol.Name)
	}
	return mapped, columns, nil
}

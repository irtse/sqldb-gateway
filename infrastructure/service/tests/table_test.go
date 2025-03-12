package service_test

import (
	"sqldb-ws/infrastructure/service"
	"testing"
)

func TestNewTableRowService(t *testing.T) {
	ts := &service.TableService{}
	trow := ts.NewTableRowService(nil)
	if trow == nil {
		t.Errorf("Expected TableRowService instance, got nil")
	}
}

func TestNewTableColumnService(t *testing.T) {
	ts := &service.TableService{}
	col := ts.NewTableColumnService(nil, "view")
	if col == nil {
		t.Errorf("Expected TableColumnService instance, got nil")
	}
}

func TestTemplate(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Template("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestMath(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Math("sum", "some_restriction")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}

func TestGet(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Get("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDelete(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Delete("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestVerify(t *testing.T) {
	ts := &service.TableService{}
	typeDesc, ok := ts.Verify("column_name")
	if !ok {
		t.Errorf("Expected column verification to pass, but it failed")
	}
	if typeDesc == "" {
		t.Errorf("Expected type description, but got empty string")
	}
}

func TestCreate(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Create(map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestUpdate(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Update(map[string]interface{}{}, "some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDeleteWithNonExistentTable(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Delete("non_existent_table")
	if err == nil {
		t.Errorf("Expected error for non-existent table deletion, got nil")
	}
}

func TestCreateWithoutValues(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Create(map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error for missing values, got nil")
	}
}

func TestUpdateWithoutValues(t *testing.T) {
	ts := &service.TableService{}
	_, err := ts.Update(map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error for missing update values, got nil")
	}
}

func TestVerifyNonExistentTable(t *testing.T) {
	ts := &service.TableService{}
	typeDesc, ok := ts.Verify("non_existent_table")
	if ok {
		t.Errorf("Expected verification to fail, but it passed")
	}
	if typeDesc != "" {
		t.Errorf("Expected empty type description, but got: %v", typeDesc)
	}
}

func TestGetWithNoResults(t *testing.T) {
	ts := &service.TableService{}
	res, err := ts.Get("non_existent_condition")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected no results, but got: %v", res)
	}
}

func TestUpdateWithValidValues(t *testing.T) {
	tcs := &service.TableService{}
	_, err := tcs.Update(map[string]interface{}{"type": "varchar", "name": "updated_column"})
	if err != nil {
		t.Errorf("Unexpected error while updating: %v", err)
	}
}

func TestDeleteNonExistentTable(t *testing.T) {
	tcs := &service.TableService{}
	tcs.Name = "non_existent_table"
	_, err := tcs.Delete()
	if err == nil {
		t.Errorf("Expected error for deleting non-existent table, but got nil")
	}
}

func TestRetrieveTableValid(t *testing.T) {
	// Simulate retrieval with valid table name
	_, _, err := service.RetrieveTable("valid_table", nil)
	if err != nil {
		t.Errorf("Unexpected error retrieving valid table: %v", err)
	}
}

func TestRetrieveTableInvalid(t *testing.T) {
	// Simulate retrieval with invalid table name
	_, _, err := service.RetrieveTable("invalid_table", nil)
	if err == nil {
		t.Errorf("Expected error for retrieving invalid table, but got nil")
	}
}

func TestWriteCreate(t *testing.T) {
	tcs := &service.TableService{}
	_, err := tcs.Create(map[string]interface{}{"name": "test_table", "fields": []interface{}{}})
	if err != nil {
		t.Errorf("Unexpected error while creating table: %v", err)
	}
}

func TestWriteUpdate(t *testing.T) {
	tcs := &service.TableService{}
	_, err := tcs.Update(map[string]interface{}{"name": "test_table", "fields": []interface{}{}})
	if err != nil {
		t.Errorf("Unexpected error while updating table: %v", err)
	}
}

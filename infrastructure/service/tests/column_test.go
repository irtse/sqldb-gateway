package service_test

import (
	"errors"
	connector_test "sqldb-ws/infrastructure/connector/tests"
	"sqldb-ws/infrastructure/service"
	"testing"
)

func TestColumnTemplate(t *testing.T) {
	tcs := &service.TableColumnService{
		InfraService: service.InfraService{
			DB:                 &connector_test.MockDB{},
			SpecializedService: &service.InfraSpecializedService{},
		},
	}
	_, err := tcs.Template("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnMath(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Math("sum", "some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnGet(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Get("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnVerify(t *testing.T) {
	tcs := &service.TableColumnService{}
	typeDesc, ok := tcs.Verify("column_name")
	if !ok {
		t.Errorf("Expected column verification to pass, but it failed")
	}
	if typeDesc == "" {
		t.Errorf("Expected type description, but got empty string")
	}
}

func TestColumnCreate(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Create(map[string]interface{}{})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnUpdate(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Update(map[string]interface{}{}, "some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnDelete(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Delete("some_restriction")
	if err != nil && !errors.Is(err, errors.New("can't delete protected root db columns")) {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestColumnUpdateWithMissingValues(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Update(map[string]interface{}{"type": "", "name": ""})
	if err == nil {
		t.Errorf("Expected error for missing type & name, but got nil")
	}
}

func TestColumnDeleteProtectedDBColumn(t *testing.T) {
	tcs := &service.TableColumnService{}
	tcs.Name = "db_protected_table"
	_, err := tcs.Delete()
	if err == nil || err.Error() != "can't delete protected root db columns" {
		t.Errorf("Expected error for deleting protected columns, got: %v", err)
	}
}

func TestColumnCreateWithoutQueries(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Create(map[string]interface{}{})
	if err == nil || err.Error() != "no query to execute" {
		t.Errorf("Expected error for no queries, got: %v", err)
	}
}

func TestColumnMathWithInvalidAlgo(t *testing.T) {
	tcs := &service.TableColumnService{}
	_, err := tcs.Math("invalid_algo", "some_restriction")
	if err == nil {
		t.Errorf("Expected error for invalid algorithm, but got nil")
	}
}

func TestColumnGetWithNoResults(t *testing.T) {
	tcs := &service.TableColumnService{}
	res, err := tcs.Get("non_existent_condition")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected no results, but got: %v", res)
	}
}

func TestColumnDeleteNonExistentColumn(t *testing.T) {
	tcs := &service.TableColumnService{}
	tcs.Views = "non_existent_column"
	_, err := tcs.Delete()
	if err == nil {
		t.Errorf("Expected error for deleting non-existent column, but got nil")
	}
}

func TestColumnVerifyNonExistentColumn(t *testing.T) {
	tcs := &service.TableColumnService{}
	typeDesc, ok := tcs.Verify("non_existent_column")
	if ok {
		t.Errorf("Expected verification to fail, but it passed")
	}
	if typeDesc != "" {
		t.Errorf("Expected empty type description, but got: %v", typeDesc)
	}
}

package service_test

import (
	"sqldb-ws/infrastructure/service"
	"testing"
)

func TestRowTemplate(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Template("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowVerify(t *testing.T) {
	tcs := &service.TableRowService{}
	name, ok := tcs.Verify("row_id")
	if !ok {
		t.Errorf("Expected row verification to pass, but it failed")
	}
	if name != "row_id" {
		t.Errorf("Expected name to be 'row_id', but got: %v", name)
	}
}

func TestRowMath(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Math("sum", "some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowGet(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Get("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowCreate(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Create(map[string]interface{}{"name": "test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowCreateWithNoData(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Create(map[string]interface{}{})
	if err == nil || err.Error() != "no data to insert" {
		t.Errorf("Expected error for no data, got: %v", err)
	}
}

func TestRowUpdate(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Update(map[string]interface{}{"name": "updated_test"}, "some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowDelete(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Delete("some_restriction")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestRowDeleteWithNoRestriction(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Delete()
	if err == nil || err.Error() != "no restriction can't delete all" {
		t.Errorf("Expected error for missing restriction, got: %v", err)
	}
}

func TestRowUpdateWithNoRecord(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Update(map[string]interface{}{})
	if err == nil {
		t.Errorf("Expected error for empty record, but got nil")
	}
}

func TestRowMathWithInvalidAlgo(t *testing.T) {
	tcs := &service.TableRowService{}
	_, err := tcs.Math("invalid_algo", "some_restriction")
	if err == nil {
		t.Errorf("Expected error for invalid algorithm, but got nil")
	}
}

func TestRowGetWithNoResults(t *testing.T) {
	tcs := &service.TableRowService{}
	res, err := tcs.Get("non_existent_condition")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("Expected no results, but got: %v", res)
	}
}

func TestRowVerifyNonExistentRow(t *testing.T) {
	tcs := &service.TableRowService{}
	name, ok := tcs.Verify("non_existent_id")
	if ok {
		t.Errorf("Expected verification to fail, but it passed")
	}
	if name != "non_existent_id" {
		t.Errorf("Expected name to be 'non_existent_id', but got: %v", name)
	}
}

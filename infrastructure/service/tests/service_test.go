package service_test

import (
	"errors"
	"os"
	"sqldb-ws/infrastructure/service"
	"testing"
)

func TestFill(t *testing.T) {
	infra := &service.InfraService{}
	infra.Fill("test_table", true, "admin_user", map[string]interface{}{"key": "value"})

	if infra.Name != "test_table" {
		t.Errorf("Expected name to be 'test_table', got %v", infra.Name)
	}
	if infra.User != "admin_user" {
		t.Errorf("Expected user to be 'admin_user', got %v", infra.User)
	}
	if infra.SuperAdmin != true {
		t.Errorf("Expected SuperAdmin to be true, got %v", infra.SuperAdmin)
	}
}

func TestGenerateFromTemplateInvalidFile(t *testing.T) {
	infra := &service.InfraService{Name: "output.txt"}
	err := infra.GenerateFromTemplate("non_existent_template.txt")
	if err == nil {
		t.Errorf("Expected an error for non-existent template file, but got nil")
	}
}

func TestDBErrorLoggingEnabled(t *testing.T) {
	os.Setenv("log", "enable")
	infra := &service.InfraService{NoLog: false}

	res, err := infra.DBError(nil, errors.New("test error"))
	if err == nil || err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err)
	}
	if res != nil {
		t.Errorf("Expected result to be nil, got %v", res)
	}
}

func TestDBErrorLoggingDisabled(t *testing.T) {
	os.Setenv("log", "disable")
	infra := &service.InfraService{NoLog: true}

	res, err := infra.DBError(nil, errors.New("test error"))
	if err == nil || err.Error() != "test error" {
		t.Errorf("Expected 'test error', got %v", err)
	}
	if res != nil {
		t.Errorf("Expected result to be nil, got %v", res)
	}
}

func TestGenerateQueryFilter(t *testing.T) {
	s := &service.InfraSpecializedService{}
	q1, q2, q3, q4 := s.GenerateQueryFilter("test_table", "condition")

	if q1 != "" || q2 != "" || q3 != "" || q4 != "" {
		t.Errorf("Expected empty query filters, got %v %v %v %v", q1, q2, q3, q4)
	}
}

func TestVerifyDataIntegrity(t *testing.T) {
	s := &service.InfraSpecializedService{}
	record := map[string]interface{}{"id": 1}
	res, err, exists := s.VerifyDataIntegrity(record, "test_table")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if exists {
		t.Errorf("Expected exists to be false, got true")
	}
	if res["id"] != 1 {
		t.Errorf("Expected record id 1, got %v", res["id"])
	}
}

func TestSpecializedDeleteRow(t *testing.T) {
	s := &service.InfraSpecializedService{}
	s.SpecializedDeleteRow([]map[string]interface{}{}, "test_table")
	// No expected output, just ensuring no panic occurs
}

func TestSpecializedUpdateRow(t *testing.T) {
	s := &service.InfraSpecializedService{}
	s.SpecializedUpdateRow([]map[string]interface{}{}, map[string]interface{}{})
	// No expected output, just ensuring no panic occurs
}

func TestSpecializedCreateRow(t *testing.T) {
	s := &service.InfraSpecializedService{}
	s.SpecializedCreateRow(map[string]interface{}{}, "test_table")
	// No expected output, just ensuring no panic occurs
}

package schema_service_tests

import (
	service "sqldb-ws/domain/specialized_service/schema_service"
	"sqldb-ws/domain/utils"
	"testing"
)

func TestViewVerifyDataIntegrity(t *testing.T) {
	service := service.ViewService{}
	record := map[string]interface{}{"id": 1}
	tablename := "test_table"
	_, err, valid := service.VerifyDataIntegrity(record, tablename)
	if err != nil || !valid {
		t.Errorf("Expected valid record, got error: %v", err)
	}
}

func TestViewGenerateQueryFilter(t *testing.T) {
	service := service.ViewService{}
	restr, _, _, _ := service.GenerateQueryFilter("test_table")
	if restr == "" {
		t.Errorf("Expected non-empty restriction string")
	}
}

func TestViewTransformToGenericView(t *testing.T) {
	service := service.ViewService{}
	results := utils.Results{{"index": 1}, {"index": 2}}
	res := service.TransformToGenericView(results, "test_table")
	if len(res) != 2 {
		t.Errorf("Expected 2 results, got %d", len(res))
	}
}

// TODO COMPLEXIFY TESTS

package schema_test

import (
	"sqldb-ws/domain/schema"
	"sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"testing"
)

func TestGetTablename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "users"},
	}

	for _, test := range tests {
		result := schema.GetTablename(test.input)
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestGetSchemaByFieldID_NotFound(t *testing.T) {
	_, err := schema.GetSchemaByFieldID(999)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestGetFieldByID_NotFound(t *testing.T) {
	_, err := schema.GetFieldByID(999)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestSetSchema(t *testing.T) {
	_, err := schema.SetSchema(map[string]interface{}{"name": "test_schema"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestHasSchema_NotFound(t *testing.T) {
	if schema.HasSchema("unknown_table") {
		t.Error("Expected false but got true")
	}
}

func TestHasField_NotFound(t *testing.T) {
	if schema.HasField("unknown_table", "unknown_field") {
		t.Error("Expected false but got true")
	}
}

func TestGetSchema_NotFound(t *testing.T) {
	_, err := schema.GetSchema("unknown_table")
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestGetSchemaByID_NotFound(t *testing.T) {
	_, err := schema.GetSchemaByID(999)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestValidateBySchema_NoSchema(t *testing.T) {
	_, err := schema.ValidateBySchema(utils.Record{}, "unknown_table", utils.CREATE, nil)
	if err == nil {
		t.Error("Expected error but got nil")
	}
}

func TestValidateBySchema_Valid(t *testing.T) {
	// Mock schema with fields
	schemaModel := models.SchemaModel{Name: "test_table", Fields: []models.FieldModel{}}
	models.SchemaRegistry["test_table"] = schemaModel

	_, err := schema.ValidateBySchema(utils.Record{}, "test_table", utils.CREATE, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestValidateBySchema_MissingField(t *testing.T) {
	models.SchemaRegistry["test_table"] = models.SchemaModel{
		Name: "test_table",
		Fields: []models.FieldModel{
			{Name: "required_field", Required: true},
		},
	}
	_, err := schema.ValidateBySchema(utils.Record{}, "test_table", utils.CREATE, nil)
	if err == nil {
		t.Error("Expected missing field error but got nil")
	}
}

func TestValidateBySchema_IgnoreSelectDelete(t *testing.T) {
	data := utils.Record{"field": "value"}
	result, err := schema.ValidateBySchema(data, "test_table", utils.SELECT, nil)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if result["field"] != "value" {
		t.Errorf("Expected record to be unchanged, got %+v", result)
	}
}

func TestValidateBySchema_UpdateValid(t *testing.T) {
	models.SchemaRegistry["test_table"] = models.SchemaModel{
		Name: "test_table",
		Fields: []models.FieldModel{
			{Name: "existing_field"},
		},
	}
	data := utils.Record{"existing_field": "value", "extra_field": "should be ignored"}
	result, _ := schema.ValidateBySchema(data, "test_table", utils.UPDATE, nil)
	if _, exists := result["extra_field"]; exists {
		t.Error("Unexpected field in update")
	}
}

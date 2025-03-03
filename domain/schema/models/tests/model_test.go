package models_test

import (
	"sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sync"
	"testing"
)

var CacheMutex sync.Mutex

func TestDeserialize(t *testing.T) {
	rec := utils.Record{"id": int64(1), "name": "TestSchema", "label": "Test Label", "category": "TestCategory"}
	schema := models.SchemaModel{}.Deserialize(rec)

	if schema.ID != 1 || schema.Name != "TestSchema" || schema.Label != "Test Label" || schema.Category != "TestCategory" {
		t.Errorf("Unexpected deserialization result: %+v", schema)
	}
}

func TestGetName(t *testing.T) {
	schema := models.SchemaModel{Name: "TestSchema"}
	if schema.GetName() != "TestSchema" {
		t.Errorf("Expected 'TestSchema', got %s", schema.GetName())
	}
}

func TestHasField(t *testing.T) {
	schema := models.SchemaModel{Fields: []models.FieldModel{{Name: "Field1"}, {Name: "Field2"}}}

	if !schema.HasField("Field1") {
		t.Errorf("Expected HasField to return true for 'Field1'")
	}
	if schema.HasField("Field3") {
		t.Errorf("Expected HasField to return false for 'Field3'")
	}
}

func TestGetField(t *testing.T) {
	schema := models.SchemaModel{Fields: []models.FieldModel{{Name: "Field1", ID: 1}}}

	field, err := schema.GetField("Field1")
	if err != nil || field.Name != "Field1" {
		t.Errorf("Expected to retrieve 'Field1', got error: %v", err)
	}

	_, err = schema.GetField("FieldX")
	if err == nil {
		t.Errorf("Expected error for non-existent field")
	}
}

func TestGetFieldByID(t *testing.T) {
	schema := models.SchemaModel{Fields: []models.FieldModel{{Name: "Field1", ID: 1}}}

	field, err := schema.GetFieldByID(1)
	if err != nil || field.ID != 1 {
		t.Errorf("Expected to retrieve field with ID 1, got error: %v", err)
	}

	_, err = schema.GetFieldByID(99)
	if err == nil {
		t.Errorf("Expected error for non-existent field ID")
	}
}

func TestViewModelToRecord(t *testing.T) {
	view := models.ViewModel{Name: "TestView"}
	record := view.ToRecord()
	if record["name"] != "TestView" {
		t.Errorf("Expected 'TestView', got %v", record["name"])
	}
}

func TestViewItemModelDefaultValues(t *testing.T) {
	item := models.ViewItemModel{}
	if item.IsEmpty != false || item.Sort != 0 {
		t.Errorf("Unexpected default values in ViewItemModel: %+v", item)
	}
}

func TestWorkflowModelDefaults(t *testing.T) {
	workflow := models.WorkflowModel{}
	if workflow.IsDismiss != false || workflow.IsClose != false {
		t.Errorf("Unexpected default values in WorkflowModel: %+v", workflow)
	}
}

func TestFieldModelConstraints(t *testing.T) {
	field := models.FieldModel{Name: "Field1", Constraint: "NOT NULL"}
	if field.Constraint != "NOT NULL" {
		t.Errorf("Expected 'NOT NULL', got %s", field.Constraint)
	}
}

func TestFilterModelDefaults(t *testing.T) {
	filter := models.FilterModel{Name: "Filter1", Type: "string", Operator: "="}
	if filter.Operator != "=" {
		t.Errorf("Expected '=', got %s", filter.Operator)
	}
}

package schema_test

import (
	"os"
	"sqldb-ws/domain/schema"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/tests"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"testing"

	"github.com/schollz/progressbar/v3"
)

// MockDomain implements DomainITF for testing purpose

func TestLoad(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	schema.Load(mockDomain)
}

func TestInitializeTables(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	schema.InitializeTables(mockDomain, bar)
}

func TestInitializeRootTables(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	schema.InitializeRootTables(mockDomain, []sm.SchemaModel{}, bar)
}

func TestCreateRootTable(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	record := utils.Record{"name": "test_table"}
	if !schema.CreateRootTable(mockDomain, record) {
		t.Error("Expected root table creation to succeed")
	}
}

func TestCreateWorkflowView(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	schemaModel := sm.SchemaModel{Name: "test_schema"}
	schema.CreateWorkflowView(mockDomain, schemaModel, bar)
}

func TestCreateRootView(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	schema.CreateRootView(mockDomain, bar)
}

func TestCreateView(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	schemaModel := sm.SchemaModel{Name: "test_schema"}
	schema.CreateView(mockDomain, schemaModel, bar)
}

func TestCreateSuperAdmin(t *testing.T) {
	mockDomain := tests.NewMockDomain()
	bar := &progressbar.ProgressBar{}
	os.Setenv("SUPERADMIN_NAME", "admin")
	os.Setenv("SUPERADMIN_EMAIL", "admin@example.com")
	os.Setenv("SUPERADMIN_PASSWORD", "password")
	schema.CreateSuperAdmin(mockDomain, bar)
}

func TestLoadCache(t *testing.T) {
	db := connector.Open(nil)
	defer db.Close()
	schema.LoadCache(utils.ReservedParam, db)
}

func TestHasSchema(t *testing.T) {
	if schema.HasSchema("test_table") {
		t.Error("Expected schema to not exist")
	}
}

func TestHasField(t *testing.T) {
	if schema.HasField("test_table", "test_field") {
		t.Error("Expected field to not exist")
	}
}

func TestGetSchema(t *testing.T) {
	_, err := schema.GetSchema("test_table")
	if err == nil {
		t.Error("Expected error for missing schema")
	}
}

func TestGetSchemaByID(t *testing.T) {
	_, err := schema.GetSchemaByID(1)
	if err == nil {
		t.Error("Expected error for missing schema by ID")
	}
}

func TestValidateBySchema(t *testing.T) {
	data := utils.Record{"field": "value"}
	_, err := schema.ValidateBySchema(data, "test_table", utils.CREATE, nil)
	if err == nil {
		t.Error("Expected validation error")
	}
}

func TestDeleteSchema(t *testing.T) {
	schema.DeleteSchema("test_table")
}

func TestDeleteSchemaField(t *testing.T) {
	schema.DeleteSchemaField("test_table", "test_field")
}

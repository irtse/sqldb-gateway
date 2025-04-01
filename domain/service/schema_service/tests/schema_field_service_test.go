package schema_service_tests

import (
	"testing"

	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/service/schema_service"
	"sqldb-ws/domain/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDomain struct {
	mock.Mock
	utils.DomainITF
}

func (m *MockDomain) GetMethod() utils.Method {
	args := m.Called()
	return args.Get(0).(utils.Method)
}

func (m *MockDomain) GetParams() utils.Params {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *MockDomain) ValidateBySchema(record utils.Record, tablename string) (utils.Record, error) {
	args := m.Called(record, tablename)
	return args.Get(0).(utils.Record), args.Error(1)
}

func TestSchemaFieldVerifyDataIntegrity(t *testing.T) {
	mockDomain := new(MockDomain)
	schemaFields := schema_service.SchemaFields{}
	schemaFields.Domain = mockDomain

	// Case 1: Deleting root schema field should return an error
	mockDomain.On("GetMethod").Return(utils.DELETE)
	mockDomain.On("GetParams").Return(map[string]string{utils.RootRowsParam: "invalid"})

	record := map[string]interface{}{}
	_, err, _ := schemaFields.VerifyDataIntegrity(record, "test_table")
	assert.Error(t, err, "Expected error when deleting root schema field with invalid ID")

	// Case 2: Valid record should pass validation
	mockDomain.On("GetMethod").Return(utils.CREATE)
	mockDomain.On("ValidateBySchema", mock.Anything, "test_table").Return(record, nil)

	record[sm.TYPEKEY] = "valid_type"
	record[sm.LABELKEY] = "valid_label"
	result, err, valid := schemaFields.VerifyDataIntegrity(record, "test_table")
	assert.Nil(t, err)
	assert.True(t, valid)
	assert.Equal(t, "valid_type", result[sm.TYPEKEY])
}

func TestSchemaFieldSpecializedCreateRow(t *testing.T) {
	mockDomain := new(MockDomain)
	schemaFields := schema_service.SchemaFields{}
	schemaFields.Domain = mockDomain

	record := map[string]interface{}{sm.NAMEKEY: "UserDBField"}

	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.CREATE).Return(nil)
	schemaFields.SpecializedCreateRow(record, "test_table")

	mockDomain.AssertCalled(t, "SuperCall", mock.Anything, mock.Anything, utils.CREATE)
}

func TestSchemaFieldSpecializedUpdateRow(t *testing.T) {
	mockDomain := new(MockDomain)
	schemaFields := schema_service.SchemaFields{}
	schemaFields.Domain = mockDomain

	results := []map[string]interface{}{
		{sm.NAMEKEY: "test1"},
	}
	mockDomain.On("UpdateSuperCall", mock.Anything, mock.Anything).Return(nil)

	schemaFields.SpecializedUpdateRow(results, map[string]interface{}{})
	mockDomain.AssertCalled(t, "UpdateSuperCall", mock.Anything, mock.Anything)
}

func TestSchemaFieldSpecializedDeleteRow(t *testing.T) {
	mockDomain := new(MockDomain)
	schemaFields := schema_service.SchemaFields{}
	schemaFields.Domain = mockDomain

	results := []map[string]interface{}{
		{sm.NAMEKEY: "test_delete"},
	}

	mockDomain.On("DeleteSuperCall", mock.Anything).Return(nil)

	schemaFields.SpecializedDeleteRow(results, "test_table")
	mockDomain.AssertCalled(t, "DeleteSuperCall", mock.Anything)
}

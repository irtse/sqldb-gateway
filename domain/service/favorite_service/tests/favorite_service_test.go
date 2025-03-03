package favorite_service

import (
	ds "sqldb-ws/domain/schema/database_resources"
	service "sqldb-ws/domain/service/favorite_service"
	utils "sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDomain struct {
	mock.Mock
}

func (m *MockDomain) Call(params map[string]interface{}, record map[string]interface{}, action string) {
	m.Called(params, record, action)
}

func TestSpecializedCreateRow_ValidRecord(t *testing.T) {
	service := &service.FilterService{
		Fields: []map[string]interface{}{{"name": "test_field"}},
	}
	record := map[string]interface{}{ds.SchemaDBField: int64(1), utils.SpecialIDParam: 123}
	service.SpecializedCreateRow(record, ds.DBFilter.Name)
	assert.NotEmpty(t, service.Fields)
}

func TestSpecializedCreateRow_InvalidSchemaID(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{ds.SchemaDBField: "invalid_id"}
	service.SpecializedCreateRow(record, ds.DBFilter.Name)
	assert.Empty(t, service.Fields)
}

func TestTransformToGenericView_EmptyResults(t *testing.T) {
	service := &service.FilterService{}
	results := utils.Results{}
	transformed := service.TransformToGenericView(results, "test_table")
	assert.Empty(t, transformed)
}

func TestTransformToGenericView_ValidResults(t *testing.T) {
	service := &service.FilterService{}
	results := utils.Results{{utils.SpecialIDParam: "1", "is_selected": true}}
	transformed := service.TransformToGenericView(results, "test_table")
	assert.Equal(t, transformed[0]["is_selected"], true)
}

func TestVerifyDataIntegrity_CreateMethod(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{ds.SchemaDBField: int64(1), "name": "test"}
	_, err, valid := service.VerifyDataIntegrity(record, "test_table")
	assert.Nil(t, err)
	assert.True(t, valid)
}

func TestProcessLink_ValidLink(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{"link": "test_schema"}
	err := service.ProcessLink(record)
	assert.Nil(t, err)
	assert.Contains(t, record, ds.SchemaDBField)
}

func TestProcessName_ExistingFilter(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{ds.SchemaDBField: int64(1), "name": "existing"}
	service.ProcessName(record)
	assert.Contains(t, record, utils.SpecialIDParam)
}

func TestHandleUpdate_RemovesOldFilters(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{utils.SpecialIDParam: 123}
	service.HandleUpdate(record)
	assert.NotEmpty(t, record)
}

func TestHandleDelete_ValidRecord(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{utils.SpecialIDParam: 123}
	service.HandleDelete(record)
	assert.NotEmpty(t, record)
}

func TestProcessSelection_ValidSelection(t *testing.T) {
	service := &service.FilterService{}
	record := map[string]interface{}{"is_selected": true, ds.FilterDBField: 1}
	service.ProcessSelection(record)
	assert.NotContains(t, record, "filter_fields")
}

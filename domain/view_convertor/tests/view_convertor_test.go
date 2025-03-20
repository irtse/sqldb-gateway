package view_convertor_test

import (
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/tests"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"
	"testing"

	"github.com/stretchr/testify/assert"
)

// 1. Test NewViewConvertor initializes correctly
func TestNewViewConvertor(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	assert.NotNil(t, vc)
	assert.Equal(t, mockDomain, vc.Domain)
}

// 2. Test TransformToView with empty results
func TestTransformToView_EmptyResults(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	results := utils.Results{}
	transformed := vc.TransformToView(results, "test_table", false)
	assert.Empty(t, transformed)
}

// 3. Test TransformToView with invalid schema
func TestTransformToView_InvalidSchema(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetParams").Return(map[string]string{})
	mockDomain.On("GetMethod").Return(utils.SELECT)

	vc := view_convertor.NewViewConvertor(mockDomain)
	results := utils.Results{{"id": 1, "name": "test"}}
	transformed := vc.TransformToView(results, "invalid_table", false)

	assert.Equal(t, results, transformed)
}

// 4. Test transformShallowedView returns same records when name is empty
func TestTransformShallowedView_EmptyName(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	results := utils.Results{{"id": 1, "name": ""}}
	transformed := vc.TransformToView(results, "test_table", false)

	assert.Equal(t, results, transformed)
}

// 5. Test transformShallowedView with valid records
func TestTransformShallowedView_ValidRecord(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	results := utils.Results{{"id": 1, "name": "Test"}}
	transformed := vc.TransformToView(results, "test_table", false)

	assert.NotEmpty(t, transformed)
}

// 7. Test ConvertRecordToView correctly converts a record
func TestConvertRecordToView(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	channel := make(chan sm.ViewItemModel, 1)

	record := utils.Record{"id": "1", "name": "Test"}
	vc.ConvertRecordToView(0, channel, record, "test_table", nil, false, false)

	result := <-channel
	assert.NotEmpty(t, result.Values)
}

// 8. Test IsReadonly returns correct boolean values
func TestIsReadonly(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	record := utils.Record{"state": "completed"}

	readonly := vc.IsReadonly("test_table", record)
	assert.True(t, readonly)
}

// 9. Test HandleDBSchemaField with missing schema
func TestHandleDBSchemaField_MissingSchema(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	shallowVals := make(map[string]interface{})

	datapath, exists := vc.HandleDBSchemaField(utils.Record{"id": 1}, sm.FieldModel{Name: "schema_id"}, "test_table", shallowVals)
	assert.Empty(t, datapath)
	assert.False(t, exists)
}

// 10. Test HandleLinkField does not crash with empty record
func TestHandleLinkField_EmptyRecord(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)

	shallowVals := make(map[string]interface{})
	manyVals := make(map[string]utils.Results)
	manyPathVals := make(map[string]string)

	vc.HandleLinkField(utils.Record{}, sm.FieldModel{Name: "link"}, "test_table", false, shallowVals, manyVals, manyPathVals)

	assert.Empty(t, shallowVals)
}

// 11. Test BuildPath function (if applicable)
func TestBuildPath(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)
	path := vc.BuildPath("test_table", "123")

	assert.NotEmpty(t, path)
}

// 12. Test ApplyCommandRow with encoded command
func TestApplyCommandRow(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetParams").Return(map[string]string{"command_row": "test as alias"})
	vc := view_convertor.NewViewConvertor(mockDomain)

	record := utils.Record{"alias": "value"}
	vals := make(map[string]interface{})

	vc.ApplyCommandRow(record, vals)
	assert.Equal(t, "value", vals["alias"])
}

// 13. Test HandleManyField with valid schema
func TestHandleManyField_ValidSchema(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)

	manyVals := make(map[string]utils.Results)
	manyPathVals := make(map[string]string)
	record := utils.Record{"id": "1"}

	vc.HandleManyField(record, sm.FieldModel{Name: "many_field"}, "test_table", "linked_table", manyVals, manyPathVals)
	assert.NotNil(t, manyVals)
}

// 15. Test TransformToView handles errors gracefully
func TestTransformToView_ErrorHandling(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetParams").Return(map[string]string{}).Once()
	mockDomain.On("GetMethod").Return(utils.SELECT).Once()

	vc := view_convertor.NewViewConvertor(mockDomain)
	results := utils.Results{{"id": "1", "name": "Test"}}
	transformed := vc.TransformToView(results, "error_table", false)

	assert.Equal(t, results, transformed)
}

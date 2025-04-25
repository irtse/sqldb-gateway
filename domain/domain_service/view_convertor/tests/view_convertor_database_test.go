package view_convertor_test

import (
	"errors"
	"testing"

	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/tests"
	"sqldb-ws/domain/utils"
	"sqldb-ws/domain/view_convertor"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for Domain Interface
type MockDomain struct {
	mock.Mock
}

func (m *MockDomain) GetDb() *MockDB {
	args := m.Called()
	return args.Get(0).(*MockDB)
}

func (m *MockDomain) GetUserID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDomain) GetUser() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDomain) IsSuperAdmin() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDomain) VerifyAuth(table, field, level string, method utils.Method) bool {
	args := m.Called(table, field, level, method)
	return args.Bool(0)
}

func (m *MockDomain) GetEmpty() bool {
	args := m.Called()
	return args.Bool(0)
}

// Mock for Database Interface
type MockDB struct {
	mock.Mock
}

func (m *MockDB) SelectQueryWithRestriction(table string, conditions map[string]interface{}, flag bool) ([]map[string]interface{}, error) {
	args := m.Called(table, conditions, flag)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) ClearQueryFilter() {
	m.Called()
}

// Test 1: GetShortcuts returns correct shortcut mapping
func TestGetShortcuts(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDB)
	mockDomain.On("GetDb").Return(mockDB)

	mockDB.On("SelectQueryWithRestriction", "db_view", map[string]interface{}{"is_shortcut": true}, false).
		Return([]map[string]interface{}{{"name": "shortcut1", "id": "123"}}, nil)

	vc := view_convertor.NewViewConvertor(mockDomain)
	shortcuts := vc.GetShortcuts("123", []string{"get"})

	assert.Equal(t, "#123", shortcuts["shortcut1"])
}

// Test 2: FetchRecord returns correct record
func TestFetchRecord(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDB)
	mockDomain.On("GetDb").Return(mockDB)

	mockDB.On("SelectQueryWithRestriction", "test_table", map[string]interface{}{"id": "123"}, false).
		Return([]map[string]interface{}{{"id": "123", "name": "Test"}}, nil)

	vc := view_convertor.NewViewConvertor(mockDomain)
	record := vc.FetchRecord("test_table", map[string]interface{}{
		utils.SpecialIDParam: "123",
	})

	assert.NotNil(t, record)
	assert.Equal(t, "Test", record[0]["name"])
}

// Test 3: FetchRecord returns nil on error
func TestFetchRecord_Error(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDB)
	mockDomain.On("GetDb").Return(mockDB)

	mockDB.On("SelectQueryWithRestriction", "test_table", map[string]interface{}{"id": "123"}, false).
		Return(nil, errors.New("DB error"))

	vc := view_convertor.NewViewConvertor(mockDomain)
	record := vc.FetchRecord("test_table", map[string]interface{}{
		utils.SpecialIDParam: "123",
	})

	assert.Nil(t, record)
}

// Test 4: NewDataAccess creates data access entries
func TestNewDataAccess(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDB)
	mockDomain.On("GetDb").Return(mockDB)
	mockDomain.On("GetUser").Return("test_user")

	mockDB.On("SelectQueryWithRestriction", "db_user", mock.Anything, true).
		Return([]map[string]interface{}{{"id": "user123"}}, nil)

	vc := view_convertor.NewViewConvertor(mockDomain)
	vc.NewDataAccess(1, []string{"dest1"}, utils.CREATE)

	mockDB.AssertExpectations(t)
}

// Test 5: GetViewFields returns empty on schema error
func TestGetViewFields_Error(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("IsSuperAdmin").Return(false)

	vc := view_convertor.NewViewConvertor(mockDomain)
	schemes, id, keysOrdered, cols, actions, isReadonly := vc.GetViewFields("invalid_table", false)

	assert.Empty(t, schemes)
	assert.Equal(t, int64(-1), id)
	assert.Empty(t, keysOrdered)
	assert.Empty(t, cols)
	assert.Empty(t, actions)
	assert.True(t, isReadonly)
}

// Test 7: processLinkedSchema updates field type
func TestProcessLinkedSchema(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	vc := view_convertor.NewViewConvertor(mockDomain)

	field := sm.ViewFieldModel{}
	scheme := sm.FieldModel{Type: "many-to-many"}

	vc.ProcessLinkedSchema(&field, scheme, "test_table")

	assert.Equal(t, "link", field.Type)
}

// Test 8: processPermissions updates actions
func TestProcessPermissions(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("VerifyAuth", "test_table", "", "", utils.CREATE).Return(true)
	mockDomain.On("GetEmpty").Return(false)

	vc := view_convertor.NewViewConvertor(mockDomain)
	field := sm.ViewFieldModel{}
	actions := []string{}

	field, actions = vc.ProcessPermissions(field, sm.FieldModel{}, "test_table", actions, sm.SchemaModel{})

	assert.Contains(t, actions, "create")
}

// Test 9: checkAndAddImportAction adds import if conditions met
func TestCheckAndAddImportAction(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDB)
	mockDomain.On("GetDb").Return(mockDB)

	mockDB.On("SelectQueryWithRestriction", "db_workflow", mock.Anything, false).
		Return([]map[string]interface{}{{"id": "workflow123"}}, nil)

	vc := view_convertor.NewViewConvertor(mockDomain)
	actions := []string{}

	actions = vc.CheckAndAddImportAction(actions, sm.SchemaModel{})

	assert.Contains(t, actions, "import")
}

// Test 10: handleRecursivePermissions updates schema
func TestHandleRecursivePermissions(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("VerifyAuth", "linked_table", "", "", utils.SELECT).Return(true)

	vc := view_convertor.NewViewConvertor(mockDomain)
	field := sm.ViewFieldModel{}
	scheme := sm.FieldModel{Type: "many", Name: "linked_table"}

	field = vc.HandleRecursivePermissions(field, scheme, utils.SELECT)

	assert.Equal(t, "link", field.Type)
	assert.Contains(t, field.Actions, "select")
}

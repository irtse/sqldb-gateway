package filter_test

import (
	"errors"
	"testing"

	"sqldb-ws/domain/domain_service/filter"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock dependencies
type MockDatabase struct {
	mock.Mock
}

func (m *MockDatabase) BuildSelectQueryWithRestriction(table string, filters map[string]interface{}, distinct bool, field ...string) string {
	args := m.Called(table, filters, distinct, field)
	return args.String(0)
}

func (m *MockDatabase) SelectQueryWithRestriction(table string, filters []interface{}, distinct bool) ([]map[string]interface{}, error) {
	args := m.Called(table, filters, distinct)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDatabase) SimpleMathQuery(operation, table string, filters []string, distinct bool) ([]map[string]interface{}, error) {
	args := m.Called(operation, table, filters, distinct)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func TestGetEntityFilterQuery(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	expectedQuery := "SELECT * FROM entity WHERE user_id = ?"
	mockDB.On("BuildSelectQueryWithRestriction", ds.DBEntityUser.Name, mock.Anything, true, "id").Return(expectedQuery)

	query := service.GetEntityFilterQuery("id")

	assert.Equal(t, expectedQuery, query)
}

func TestGetUserFilterQuery(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)
	mockDomain.On("GetUser").Return("test_user")

	expectedQuery := "SELECT * FROM users WHERE name = 'test_user' AND email = 'test_user'"
	mockDB.On("BuildSelectQueryWithRestriction", ds.DBUser.Name, mock.Anything, true, "id").Return(expectedQuery)

	query := mockDomain.GetUserID()

	assert.Equal(t, expectedQuery, query)
}

func TestCountNewDataAccess_Success(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDB.On("SelectQueryWithRestriction", "test_table", mock.Anything, false).Return([]map[string]interface{}{
		{"id": "123"}, {"id": "456"},
	}, nil)

	mockDB.On("SimpleMathQuery", "COUNT", "test_table", mock.Anything, false).Return([]map[string]interface{}{
		{"result": 2},
	}, nil)

	ids, count := service.CountNewDataAccess("test_table", []interface{}{})

	assert.Equal(t, []string{"123", "456"}, ids)
	assert.Equal(t, int64(2), count)
}

func TestCountNewDataAccess_EmptyResult(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDB.On("SelectQueryWithRestriction", "test_table", mock.Anything, false).Return([]map[string]interface{}{}, nil)
	mockDB.On("SimpleMathQuery", "COUNT", "test_table", mock.Anything, false).Return([]map[string]interface{}{
		{"result": nil},
	}, nil)

	ids, count := service.CountNewDataAccess("test_table", []interface{}{})

	assert.Empty(t, ids)
	assert.Equal(t, int64(0), count)
}

func TestGetFilterFields_Success(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDB.On("SelectQueryWithRestriction", ds.DBFilterField.Name, mock.Anything, false).Return([]map[string]interface{}{
		{"index": 1, "field": "name"},
		{"index": 2, "field": "email"},
	}, nil)

	fields := service.GetFilterFields("viewfilter123", "schema123")

	assert.Len(t, fields, 2)
	assert.Equal(t, "name", fields[0]["field"])
	assert.Equal(t, "email", fields[1]["field"])
}

func TestGetFilterFields_NoViewFilterID(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	service := filter.FilterService{Domain: mockDomain}

	fields := service.GetFilterFields("", "schema123")

	assert.Empty(t, fields)
}

func TestGetFilterFields_ErrorHandling(t *testing.T) {
	mockDB := new(MockDatabase)
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDB.On("SelectQueryWithRestriction", ds.DBFilterField.Name, mock.Anything, false).Return(nil, errors.New("DB error"))

	fields := service.GetFilterFields("viewfilter123", "schema123")

	assert.Empty(t, fields)
}

func TestGetFilterIDs_Success(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDatabase)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDomain.On("GetParams").Return(map[string]string{
		"filterID":     "123",
		"viewFilterID": "456",
	})

	mockDB.On("SelectQueryWithRestriction", ds.DBFilterField.Name, mock.Anything, false).Return([]map[string]interface{}{
		{"filter_id": "123"},
	}, nil)

	filterIDs := service.GetFilterIDs("123", "456", "schemaID")

	assert.Equal(t, "123", filterIDs["filterID"])
	assert.Equal(t, "456", filterIDs["viewFilterID"])
}

func TestGetFilterIDs_MissingParams(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDomain.On("GetParams").Return(map[string]string{})

	service := filter.FilterService{Domain: mockDomain}

	filterIDs := service.GetFilterIDs("123", "456", "schemaID")

	assert.Equal(t, "123", filterIDs["filterID"])
	assert.Equal(t, "456", filterIDs["viewFilterID"])
}

func TestGetFilterIDs_ErrorHandling(t *testing.T) {
	mockDomain := new(tests.MockDomain)
	mockDB := new(MockDatabase)
	mockDomain.On("GetDb").Return(mockDB)

	service := filter.FilterService{Domain: mockDomain}

	mockDomain.On("GetParams").Return(map[string]string{
		"filterID": "123",
	})

	mockDB.On("SelectQueryWithRestriction", ds.DBFilterField.Name, mock.Anything, false).Return(nil, errors.New("DB error"))

	filterIDs := service.GetFilterIDs("123", "", "schemaID")

	assert.Equal(t, "123", filterIDs["filterID"])
}

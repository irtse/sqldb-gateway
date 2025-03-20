package tests

import (
	"errors"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"sync"

	"github.com/stretchr/testify/mock"
)

type MockDomain struct {
	mock.Mock
	records map[string][]map[string]interface{}
	params  utils.Params
	mu      sync.Mutex
}

func NewMockDomain() *MockDomain {
	return &MockDomain{
		records: make(map[string][]map[string]interface{}),
		params:  make(utils.Params),
	}
}

func (m *MockDomain) SuperCall(params utils.Params, record utils.Record, method utils.Method, isOwn bool, args ...interface{}) (utils.Results, error) {
	return utils.Results{}, nil
}

func (m *MockDomain) CreateSuperCall(params utils.Params, record utils.Record, args ...interface{}) (utils.Results, error) {
	return utils.Results{}, nil
}

func (m *MockDomain) UpdateSuperCall(params utils.Params, rec utils.Record, args ...interface{}) (utils.Results, error) {
	return utils.Results{}, nil
}

func (m *MockDomain) DeleteSuperCall(params utils.Params, args ...interface{}) (utils.Results, error) {
	return utils.Results{}, nil
}

func (m *MockDomain) Call(params utils.Params, rec utils.Record, mthd utils.Method, args ...interface{}) (utils.Results, error) {
	return utils.Results{}, nil
}

func (m *MockDomain) GetDb() *connector.Database {
	return &connector.Database{}
}

func (m *MockDomain) GetMethod() utils.Method {
	return utils.SELECT
}

func (m *MockDomain) GetTable() string {
	return "mock_table"
}

func (m *MockDomain) GetUser() string {
	return "mock_user"
}

func (m *MockDomain) GetEmpty() bool {
	return false
}

func (m *MockDomain) GetParams() utils.Params {
	return m.params
}

func (m *MockDomain) HandleRecordAttributes(record utils.Record) {
}

func (m *MockDomain) IsOwn(checkPerm, force bool, method utils.Method) bool {
	return true
}

func (m *MockDomain) IsSuperCall() bool {
	return false
}

func (m *MockDomain) IsSuperAdmin() bool {
	return false
}

func (m *MockDomain) IsShallowed() bool {
	return false
}

func (m *MockDomain) IsLowerResult() bool {
	return false
}

func (m *MockDomain) VerifyAuth(tableName, colName, level string, method utils.Method, args ...string) bool {
	return true
}

// Simulate fetching records based on a table name
func (m *MockDomain) FetchRecords(tableName string, params utils.Params) ([]map[string]interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if records, exists := m.records[tableName]; exists {
		return records, nil
	}
	return nil, errors.New("no records found")
}

// Set mock responses for FetchRecords
func (m *MockDomain) SetFetchRecordResponse(tableName string, records []map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records[tableName] = records
}

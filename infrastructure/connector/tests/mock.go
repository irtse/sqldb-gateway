package connector_test

import (
	"database/sql"
	"sqldb-ws/infrastructure/connector"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetDriver() string {
	return connector.PostgresDriver
}

func (m *MockDB) GetConn() *sql.DB {
	args := m.Called()
	return args.Get(0).(*sql.DB)
}

func (m *MockDB) GetSQLView() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) GetSQLOrder() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) GetSQLDir() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) GetSQLLimit() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) GetSQLRestriction() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) SetSQLView(s string) {
	m.Called(s)
}

func (m *MockDB) SetSQLOrder(s string) {
	m.Called(s)
}

func (m *MockDB) SetSQLLimit(s string) {
	m.Called(s)
}

func (m *MockDB) SetSQLRestriction(s string) {
	m.Called(s)
}

func (m *MockDB) Close() {
	m.Called()
}

func (m *MockDB) ClearQueryFilter() *connector.Database {
	args := m.Called()
	return args.Get(0).(*connector.Database)
}

func (m *MockDB) DeleteQueryWithRestriction(name string, restrictions map[string]interface{}, isOr bool) error {
	args := m.Called(name, restrictions, isOr)
	return args.Error(0)
}

func (m *MockDB) SelectQueryWithRestriction(name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error) {
	args := m.Called(name, restrictions, isOr)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) SimpleMathQuery(algo string, name string, restrictions interface{}, isOr bool) ([]map[string]interface{}, error) {
	args := m.Called(algo, name, restrictions, isOr)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) MathQuery(algo string, name string, naming ...string) ([]map[string]interface{}, error) {
	args := m.Called(algo, name, naming)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) SchemaQuery(name string) ([]map[string]interface{}, error) {
	args := m.Called(name)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) ListTableQuery() ([]map[string]interface{}, error) {
	args := m.Called()
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) CreateTableQuery(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockDB) UpdateQuery(name string, record map[string]interface{}, restriction map[string]interface{}, isOr bool) error {
	args := m.Called(name, record, restriction, isOr)
	return args.Error(0)
}

func (m *MockDB) DeleteQuery(name string, colName string) error {
	args := m.Called(name, colName)
	return args.Error(0)
}

func (m *MockDB) Prepare(query string) (*sql.Stmt, error) {
	args := m.Called(query)
	return nil, args.Error(0)
}

func (m *MockDB) Query(query string) error {
	args := m.Called(query)
	return args.Error(0)
}

func (m *MockDB) QueryRow(query string) (int64, error) {
	args := m.Called(query)
	return 0, args.Error(0)
}

func (m *MockDB) QueryAssociativeArray(query string) ([]map[string]interface{}, error) {
	args := m.Called(query)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockDB) BuildDeleteQueryWithRestriction(name string, restrictions map[string]interface{}, isOr bool) string {
	args := m.Called(name, restrictions, isOr)
	return args.String(0)
}

func (m *MockDB) BuildSimpleMathQueryWithRestriction(algo, name string, restrictions interface{}, isOr bool, restr ...string) string {
	args := m.Called(algo, name, restrictions, isOr, restr)
	return args.String(0)
}

func (m *MockDB) BuildSelectQueryWithRestriction(name string, restrictions interface{}, isOr bool, view ...string) string {
	args := m.Called(name, restrictions, isOr, view)
	return args.String(0)
}

func (m *MockDB) BuildMathQuery(algo, name string, naming ...string) string {
	args := m.Called(algo, name, naming)
	return args.String(0)
}

func (m *MockDB) BuildDeleteQuery(tableName, colName string) string {
	args := m.Called(tableName, colName)
	return args.String(0)
}

func (m *MockDB) BuildDropTableQueries(name string) []string {
	args := m.Called(name)
	return args.Get(0).([]string)
}

func (m *MockDB) BuildSchemaQuery(name string) string {
	args := m.Called(name)
	return args.String(0)
}

func (m *MockDB) BuildListTableQuery() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDB) BuildCreateTableQuery(name string) string {
	args := m.Called(name)
	return args.String(0)
}

func (m *MockDB) BuildCreateQueries(tableName, values, cols, typ string) []string {
	args := m.Called(tableName, values, cols, typ)
	return args.Get(0).([]string)
}

func (m *MockDB) ApplyQueryFilters(restr, order, limit, views string, additionalRestrictions ...string) {
	m.Called(restr, order, limit, views, additionalRestrictions)
}

func (m *MockDB) BuildUpdateQuery(col string, value interface{}, set string, cols, colValues []string, verify func(string) (string, bool)) (string, []string, []string) {
	args := m.Called(col, value, set, cols, colValues, verify)
	return args.String(0), args.Get(1).([]string), args.Get(2).([]string)
}

func (m *MockDB) BuildUpdateQueryWithRestriction(tableName string, record, restrictions map[string]interface{}, isOr bool) (string, error) {
	args := m.Called(tableName, record, restrictions, isOr)
	return args.String(0), args.Error(1)
}

func (m *MockDB) BuildUpdateRowQuery(tableName string, record map[string]interface{}, verify func(string) (string, bool)) (string, error) {
	args := m.Called(tableName, record, verify)
	return args.String(0), args.Error(1)
}

func (m *MockDB) BuildUpdateColumnQueries(tableName string, record map[string]interface{}, verify func(string) (string, bool)) ([]string, error) {
	args := m.Called(tableName, record, verify)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDB) RowResultToMap(rows *sql.Rows, columnNames []string, columnType map[string]string) (map[string]interface{}, error) {
	args := m.Called(rows, columnNames, columnType)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockDB) ParseColumnValue(colType string, val *interface{}) interface{} {
	args := m.Called(colType, val)
	return args.Get(0)
}

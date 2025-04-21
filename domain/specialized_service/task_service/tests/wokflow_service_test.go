// Mock domain and services
package task_service_test

import (
	ds "sqldb-ws/domain/schema/database_resources"
	service "sqldb-ws/domain/specialized_service/task_service"
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDomain struct {
	utils.DomainITF
}

func (m *MockDomain) IsSuperAdmin() bool { return true }
func (m *MockDomain) PermsCheck(schema, action, resource string, perm utils.Method) bool {
	return true
}
func (m *MockDomain) TransformToView(res utils.Results, tableName string, flag bool) utils.Results {
	return res
}
func (m *MockDomain) DefineQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return "", "", "", ""
}
func (m *MockDomain) GetMethod() utils.Method { return utils.CREATE }
func (m *MockDomain) ValidateBySchema(record utils.Record, tableName string) (utils.Record, error) {
	return record, nil
}

func TestTransformToGenericView(t *testing.T) {
	service := &service.WorkflowService{}
	service.Domain = &MockDomain{}

	mockResults := utils.Results{
		{ds.SchemaDBField: float64(1)},
	}

	res := service.TransformToGenericView(mockResults, "test_table")
	assert.Equal(t, len(mockResults), len(res))
}

func TestGenerateQueryFilter(t *testing.T) {
	service := &service.WorkflowService{}
	service.Domain = &MockDomain{}

	q1, q2, q3, q4 := service.GenerateQueryFilter("test_table")
	assert.Equal(t, "", q1)
	assert.Equal(t, "", q2)
	assert.Equal(t, "", q3)
	assert.Equal(t, "", q4)
}

func TestVerifyDataIntegrity(t *testing.T) {
	service := &service.WorkflowService{}
	service.Domain = &MockDomain{}

	record := map[string]interface{}{"key": "value"}
	result, err, flag := service.VerifyDataIntegrity(record, "test_table")
	assert.Nil(t, err)
	assert.False(t, flag)
	assert.Equal(t, record, result)
}

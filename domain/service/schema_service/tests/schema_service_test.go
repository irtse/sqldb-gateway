package schema_service_tests

import (
	"testing"

	"sqldb-ws/domain/schema/models"
	service "sqldb-ws/domain/service/schema_service"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (m *MockDomain) GetDb() *connector.Database {
	args := m.Called()
	return args.Get(0).(*connector.Database)
}

func SchemaTestVerifyDataIntegrity(t *testing.T) {
	mockDomain := new(MockDomain)
	svc := service.SchemaService{}
	svc.Domain = mockDomain

	tests := []struct {
		name      string
		method    string
		record    map[string]interface{}
		tablename string
		expectErr bool
	}{
		{
			name:      "Delete method with root DB",
			method:    utils.DELETE.String(),
			record:    map[string]interface{}{"id": 1},
			tablename: "test_table",
			expectErr: false,
		},
		{
			name:      "Delete method with non-root DB",
			method:    utils.DELETE.String(),
			record:    map[string]interface{}{"id": 2},
			tablename: "non_root_table",
			expectErr: true,
		},
		{
			name:      "Non-delete method",
			method:    utils.UPDATE.String(),
			record:    map[string]interface{}{"id": 3},
			tablename: "update_table",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomain.On("GetMethod").Return(tt.method)
			record, err, _ := svc.VerifyDataIntegrity(tt.record, tt.tablename)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.record, record)
			}
		})
	}
}

func SchemaTestSpecializedDeleteRow(t *testing.T) {
	svc := service.SchemaService{}
	mockDomain := new(MockDomain)
	svc.Domain = mockDomain

	mockDomain.On("SetIsCustom", true).Return()
	mockDomain.On("DeleteSuperCall", mock.Anything).Return()

	results := []map[string]interface{}{{"id": 1}, {"id": 2}}
	svc.SpecializedDeleteRow(results, "test_table")

	mockDomain.AssertCalled(t, "SetIsCustom", true)
	mockDomain.AssertCalled(t, "DeleteSuperCall", mock.Anything)
}

func SchemaTestSpecializedCreateRow(t *testing.T) {
	svc := service.SchemaService{}
	mockDomain := new(MockDomain)
	svc.Domain = mockDomain

	record := map[string]interface{}{models.NAMEKEY: "test_record"}
	svc.SpecializedCreateRow(record, "test_table")

	mockDomain.AssertCalled(t, "CreateSuperCall", mock.Anything, mock.Anything)
}

func SchemaTestSpecializedUpdateRow(t *testing.T) {
	svc := service.SchemaService{}
	mockDomain := new(MockDomain)
	svc.Domain = mockDomain

	record := map[string]interface{}{models.NAMEKEY: "update_record"}
	svc.SpecializedUpdateRow(nil, record)

	mockDomain.AssertCalled(t, "GetDb")
}

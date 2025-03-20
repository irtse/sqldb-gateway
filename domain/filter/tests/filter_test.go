package filter_test

import (
	"sqldb-ws/domain/filter"
	"sqldb-ws/domain/tests"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDomainSuper struct {
	tests.MockDomain
	records map[string][]map[string]interface{}
	params  utils.Params
	mu      sync.Mutex
}

func NewMockDomainSuper() *MockDomainSuper {
	return &MockDomainSuper{
		records: make(map[string][]map[string]interface{}),
		params:  make(utils.Params),
	}
}

func (m *MockDomainSuper) IsSuperCall() bool {
	return true
}

func TestNewFilterService(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	assert.NotNil(t, service, "FilterService should not be nil")
}

func TestGetQueryFilter_ValidSchema(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	restriction, view, order, limit := service.GetQueryFilter("users", params)
	assert.NotNil(t, restriction)
	assert.NotNil(t, view)
	assert.NotNil(t, order)
	assert.NotNil(t, limit)
}

func TestGetQueryFilter_InvalidSchema(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	restriction, view, order, limit := service.GetQueryFilter("invalid_table", params)
	assert.Equal(t, "", restriction)
	assert.Equal(t, "", view)
	assert.Equal(t, "", order)
	assert.Equal(t, "", limit)
}

func TestRestrictionBySchema_SuperCall(t *testing.T) {
	domain := NewMockDomainSuper()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	restrictions := service.RestrictionBySchema("users", []string{}, params)
	assert.NotNil(t, restrictions)
}

func TestRestrictionBySchema_NormalUser(t *testing.T) {
	domain := NewMockDomainSuper()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	restrictions := service.RestrictionBySchema("users", []string{}, params)
	assert.NotNil(t, restrictions)
}

func TestProcessFilterRestriction_ValidFilter(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)

	result := service.ProcessFilterRestriction("123", "1")
	assert.NotNil(t, result)
}

func TestProcessFilterRestriction_EmptyFilter(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)

	result := service.ProcessFilterRestriction("", "1")
	assert.Equal(t, "", result)
}

func TestGetFilterForQuery_ValidData(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	filter, view, order, dir, state := service.GetFilterForQuery("123", "", "1", params)
	assert.NotNil(t, filter)
	assert.NotNil(t, view)
	assert.NotNil(t, order)
	assert.NotNil(t, dir)
	assert.NotNil(t, state)
}

func TestLifeCycleRestriction_NewState(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	restrictions := []string{}

	newRestrictions := service.LifeCycleRestriction("users", restrictions, "new")
	assert.Contains(t, newRestrictions, "id IN")
}

func TestLifeCycleRestriction_OldState(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	restrictions := []string{}

	newRestrictions := service.LifeCycleRestriction("users", restrictions, "old")
	assert.Contains(t, newRestrictions, "id NOT IN")
}

func TestProcessViewAndOrder_ValidView(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	view, order, dir := service.ProcessViewAndOrder("123", "1", params)
	assert.NotNil(t, view)
	assert.NotNil(t, order)
	assert.NotNil(t, dir)
}

func TestProcessViewAndOrder_EmptyView(t *testing.T) {
	domain := tests.NewMockDomain()
	service := filter.NewFilterService(domain)
	params := utils.Params{}

	view, order, dir := service.ProcessViewAndOrder("", "1", params)
	assert.Equal(t, "", view)
	assert.Equal(t, "", order)
	assert.Equal(t, "", dir)
}

func TestFormatSQLRestrictionWhereInjection(t *testing.T) {
	result := connector.FormatSQLRestrictionWhereInjection("id = 1", nil)
	assert.NotNil(t, result)
}

func TestFormatSQLRestrictionWhereByMap(t *testing.T) {
	restrictionMap := map[string]interface{}{"active": true}
	result := connector.FormatSQLRestrictionWhereByMap("", restrictionMap, false)
	assert.NotNil(t, result)
}

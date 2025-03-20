package tests

import (
	"errors"
	"sqldb-ws/domain"
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock domain struct

func TestSetToken_Success(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("Call", mock.Anything, mock.Anything, utils.UPDATE, mock.Anything).Return(utils.Results{
		utils.Record{"success": true},
	}, nil)

	result, err := domain.SetToken(false, "user1", "token")
	assert.NoError(t, err)
	assert.Equal(t, utils.Results{utils.Record{"success": true}}, result)
}

func TestSetToken_Error(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("Call", mock.Anything, mock.Anything, utils.UPDATE, mock.Anything).Return(nil, errors.New("database error"))

	result, err := domain.SetToken(false, "user1", "token")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestIsLogged_Success(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(utils.Results{{"link_id": "1"}}, nil)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false, mock.Anything).Return(utils.Results{{"name": "user1"}}, nil)

	result, err := domain.IsLogged(false, "user1", "token")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestIsLogged_ErrorFetchingNotifications(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(nil, errors.New("fetch error"))

	result, err := domain.IsLogged(false, "user1", "token")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestIsLogged_ErrorFetchingUser(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(utils.Results{{"link_id": "1"}}, nil)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false, mock.Anything).Return(nil, errors.New("user fetch error"))

	result, err := domain.IsLogged(false, "user1", "token")
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestGetQueryFilter_ValidUser(t *testing.T) {
	filter := domain.GetQueryFilter("testuser")
	assert.Contains(t, filter, "name")
	assert.Contains(t, filter, "email")
}

func TestIsLogged_InvalidLinkID(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(utils.Results{{"link_id": "invalid"}}, nil)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false, mock.Anything).Return(utils.Results{{"name": "user1"}}, nil)

	result, err := domain.IsLogged(false, "user1", "token")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestIsLogged_NoNotifications(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(utils.Results{}, nil)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false, mock.Anything).Return(utils.Results{{"name": "user1"}}, nil)

	result, err := domain.IsLogged(false, "user1", "token")
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestIsLogged_NoUserFound(t *testing.T) {
	mockDomain := new(MockDomain)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false).Return(utils.Results{{"link_id": "1"}}, nil)
	mockDomain.On("SuperCall", mock.Anything, mock.Anything, utils.SELECT, false, mock.Anything).Return(utils.Results{}, nil)

	result, err := domain.IsLogged(false, "user1", "token")
	assert.Error(t, err)
	assert.Nil(t, result)
}

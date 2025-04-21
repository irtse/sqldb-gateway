package controller_test

import (
	"os"
	"sqldb-ws/controllers/controller"
	domain "sqldb-ws/domain"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestMySession_CreateSession(t *testing.T) {
	ctrl := &controller.AbstractController{}
	token := ctrl.MySession("user123", true, false)

	assert.NotEmpty(t, token, "Token should be generated")
	assert.Equal(t, "user123", ctrl.GetSession(controller.SESSIONS_KEY))
	assert.Equal(t, true, ctrl.GetSession(controller.ADMIN_KEY))
}

func TestMySession_DeleteSession(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.SetSession(controller.SESSIONS_KEY, "user123")
	ctrl.SetSession(controller.ADMIN_KEY, true)

	token := ctrl.MySession("user123", true, true)

	assert.Empty(t, token, "Token should be empty after deletion")
	assert.Nil(t, ctrl.GetSession(controller.SESSIONS_KEY))
	assert.Nil(t, ctrl.GetSession(controller.ADMIN_KEY))
}

func TestMySession_TokenMode(t *testing.T) {
	os.Setenv("AUTH_MODE", "token")
	ctrl := &controller.AbstractController{}
	token := ctrl.MySession("user123", false, false)

	assert.NotEmpty(t, token, "Token should be generated in token mode")
	storedToken, _ := domain.SetToken(false, "user123", "")
	assert.Equal(t, token, storedToken)
}

func TestMySession_DeleteToken(t *testing.T) {
	os.Setenv("AUTH_MODE", "token")
	ctrl := &controller.AbstractController{}
	token := ctrl.MySession("user123", false, false)

	assert.NotEmpty(t, token)
	ctrl.MySession("user123", false, true)

	storedToken, _ := domain.SetToken(false, "user123", "")
	assert.Empty(t, storedToken, "Token should be removed after delete flag")
}

func TestIsAuthorized_SessionMode(t *testing.T) {
	os.Setenv("AUTH_MODE", "session")
	ctrl := &controller.AbstractController{}
	ctrl.SetSession(controller.SESSIONS_KEY, "user123")
	ctrl.SetSession(controller.ADMIN_KEY, true)

	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	assert.True(t, superAdmin)
}

func TestIsAuthorized_SessionMode_NoUser(t *testing.T) {
	os.Setenv("AUTH_MODE", "session")
	ctrl := &controller.AbstractController{}

	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.False(t, superAdmin)
}

func TestIsAuthorized_TokenMode_ValidToken(t *testing.T) {
	os.Setenv("AUTH_MODE", "token")
	ctrl := &controller.AbstractController{}
	tokenService := &controller.Token{}
	token, _ := tokenService.Create("user123", true)

	ctrl.Ctx.Request.Header.Set("Authorization", token)
	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.NoError(t, err)
	assert.Equal(t, "user123", userID)
	assert.True(t, superAdmin)
}

func TestIsAuthorized_TokenMode_InvalidToken(t *testing.T) {
	os.Setenv("AUTH_MODE", "token")
	ctrl := &controller.AbstractController{}

	ctrl.Ctx.Request.Header.Set("Authorization", "invalid-token")
	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.False(t, superAdmin)
}

func TestIsAuthorized_MissingAuthorizationHeader(t *testing.T) {
	os.Setenv("AUTH_MODE", "token")
	ctrl := &controller.AbstractController{}

	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.False(t, superAdmin)
}

func TestToken_Create(t *testing.T) {
	tokenService := &controller.Token{}
	token, err := tokenService.Create("user123", true)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestToken_Verify_ValidToken(t *testing.T) {
	tokenService := &controller.Token{}
	tokenStr, _ := tokenService.Create("user123", true)
	token, err := tokenService.Verify(tokenStr)

	assert.NoError(t, err)
	assert.NotNil(t, token)
}

func TestToken_Verify_InvalidToken(t *testing.T) {
	tokenService := &controller.Token{}
	token, err := tokenService.Verify("invalid-token")

	assert.Error(t, err)
	assert.Nil(t, token)
}

func TestToken_ClaimsExtraction(t *testing.T) {
	tokenService := &controller.Token{}
	tokenStr, _ := tokenService.Create("user123", true)
	token, _ := tokenService.Verify(tokenStr)
	claims := token.Claims.(jwt.MapClaims)

	assert.Equal(t, "user123", claims[controller.SESSIONS_KEY])
	assert.True(t, claims[controller.ADMIN_KEY].(bool))
}

func TestIsAuthorized_AuthModeNotAllowed(t *testing.T) {
	os.Setenv("AUTH_MODE", "invalid-mode")
	ctrl := &controller.AbstractController{}

	userID, superAdmin, err := ctrl.IsAuthorized()

	assert.Error(t, err)
	assert.Equal(t, "", userID)
	assert.False(t, superAdmin)
}

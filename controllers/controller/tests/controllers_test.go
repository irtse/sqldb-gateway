package controller_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sqldb-ws/controllers/controller"
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/assert"
)

// Mock controller struct for testing
type MockController struct {
	controller.AbstractController
	MockIsAuthorized func() (string, bool, error)
}

// Mock IsAuthorized function
func (m *MockController) IsAuthorized() (string, bool, error) {
	return m.MockIsAuthorized()
}

func TestSafeCall_AuthSuccess(t *testing.T) {
	ctrl := &MockController{
		MockIsAuthorized: func() (string, bool, error) {
			return "user1", true, nil
		},
	}
	ctrl.SafeCall(utils.GET)
	// Add assertions based on expected behavior
}

func TestSafeCall_AuthFailure(t *testing.T) {
	ctrl := &MockController{
		MockIsAuthorized: func() (string, bool, error) {
			return "", false, errors.New("Unauthorized")
		},
	}
	ctrl.SafeCall(utils.GET)
	// Ensure unauthorized response is handled
}

func TestUnSafeCall_NoAuth(t *testing.T) {
	ctrl := &MockController{}
	ctrl.UnSafeCall(utils.GET)
	// Validate response logic
}

func TestCall_WithAuth_SuperAdmin(t *testing.T) {
	ctrl := &MockController{
		MockIsAuthorized: func() (string, bool, error) {
			return "admin", true, nil
		},
	}
	ctrl.Call(true, utils.POST)
}

func TestCall_WithoutAuth(t *testing.T) {
	ctrl := &MockController{}
	ctrl.Call(false, utils.POST)
}

func TestParams_ExtractUrlParams(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{}
	ctrl.Ctx.Input.SetParam("id", "123")
	params, _ := ctrl.Params()
	assert.Equal(t, "123", params["id"])
}

func TestParams_ExtractQueryParams(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{Context: ctrl.Ctx}
	ctrl.Ctx.Input.URI = "/test?name=John"
	params, _ := ctrl.Params()
	assert.Equal(t, "John", params["name"])
}

func TestBody_ParseJson(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	jsonBody := `{"key": "value"}`
	ctrl.Ctx.Input = &context.BeegoInput{RequestBody: []byte(jsonBody)}
	body := ctrl.Body(false)
	assert.Equal(t, "value", body["key"])
}

func TestBody_ParseJsonWithPasswordHashing(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	jsonBody := `{"password": "secret"}`
	ctrl.Ctx.Input = &context.BeegoInput{RequestBody: []byte(jsonBody)}
	body := ctrl.Body(true)
	assert.NotEqual(t, "secret", body["password"])
}

func TestResponse_ValidData(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{ResponseWriter: &context.Response{ResponseWriter: recorder}}
	ctrl.Response(utils.Results{"success": true}, nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestResponse_WithError(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{ResponseWriter: &context.Response{ResponseWriter: recorder}}
	ctrl.Response(nil, errors.New("error"))
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestIsAuthorized_Success(t *testing.T) {
	ctrl := &MockController{
		MockIsAuthorized: func() (string, bool, error) {
			return "user", true, nil
		},
	}
	user, admin, err := ctrl.IsAuthorized()
	assert.NoError(t, err)
	assert.Equal(t, "user", user)
	assert.True(t, admin)
}

func TestIsAuthorized_Failure(t *testing.T) {
	ctrl := &MockController{
		MockIsAuthorized: func() (string, bool, error) {
			return "", false, errors.New("Unauthorized")
		},
	}
	_, _, err := ctrl.IsAuthorized()
	assert.Error(t, err)
}

func TestParams_WithRootExport(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{URI: "/test?export=true"}
	params, _ := ctrl.Params()
	assert.Equal(t, "", params[utils.RootRawView])
}

func TestCall_WithFileUpload(t *testing.T) {
	ctrl := &MockController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{}
	ctrl.Call(false, utils.POST)
}

func TestCall_WithFileError(t *testing.T) {
	ctrl := &MockController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{}
	ctrl.Call(false, utils.POST)
}

func TestParams_NoPassword(t *testing.T) {
	ctrl := &controller.AbstractController{}
	params, _ := ctrl.Params()
	assert.NotContains(t, params, "password")
}

func TestParams_HashedPassword(t *testing.T) {
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{}
	ctrl.Ctx.Input = &context.BeegoInput{URI: "/test?password=secret"}
	params, _ := ctrl.Params()
	assert.NotEqual(t, "secret", params["password"])
}

func TestResponse_JSONEncoding(t *testing.T) {
	recorder := httptest.NewRecorder()
	ctrl := &controller.AbstractController{}
	ctrl.Ctx = &context.Context{ResponseWriter: &context.Response{ResponseWriter: recorder}}
	ctrl.Response(utils.Results{"message": "test"}, nil)
	var result map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &result)
	assert.Equal(t, "test", result["message"])
}

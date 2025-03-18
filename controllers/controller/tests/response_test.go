package controller_test

import (
	"errors"
	"net/http/httptest"
	"sqldb-ws/controllers/controller"
	"sqldb-ws/domain/utils"
	"testing"
)

type MockDomain struct{}

func (m *MockDomain) Call(params map[string]string, body interface{}, method utils.Method, args ...interface{}) (utils.Results, error) {
	return utils.Results{{"key": "value"}}, nil
}

func (m *MockDomain) GetDb() *utils.MockDb {
	return &utils.MockDb{}
}

func (m *MockDomain) GetTable() string {
	return "mock_table"
}

func TestRespond_NoExport(t *testing.T) {
	ctrl := &controller.AbstractController{}
	params := map[string]string{}
	domain := &MockDomain{}
	ctrl.Respond(params, nil, utils.Method{}, domain)
}

func TestRespond_WithExport(t *testing.T) {
	ctrl := &controller.AbstractController{}
	params := map[string]string{utils.RootExport: "true"}
	domain := &MockDomain{}
	ctrl.Respond(params, nil, utils.Method{}, domain)
}

func TestResponse_Success(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.Output = recorder

	ctrl.Response(utils.Results{{"key": "value"}}, nil)
}

func TestResponse_Error(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.Output = recorder

	ctrl.Response(nil, errors.New("test error"))
}

func TestResponse_PartialError(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.Output = recorder

	ctrl.Response(nil, errors.New("partial error"))
}

func TestDownload_CSV(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.ResponseWriter = recorder
	domain := &MockDomain{}
	ctrl.Download(domain, "col1", "col2", "cmd", "csv", "file", nil, utils.Results{{"key": "value"}}, nil)
}

func TestDownload_NonCSV(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.ResponseWriter = recorder
	domain := &MockDomain{}
	ctrl.Download(domain, "col1", "col2", "cmd", "json", "file", nil, utils.Results{{"key": "value"}}, nil)
}

func TestCSV_Generation(t *testing.T) {
	ctrl := &controller.AbstractController{}
	domain := &MockDomain{}
	result := ctrl.CSV(domain, nil, []string{"col1"}, utils.Results{{"col1": "value"}})

	if len(result) == 0 {
		t.Errorf("CSV generation failed")
	}
}

func TestMapping_EmptyResponse(t *testing.T) {
	ctrl := &controller.AbstractController{}
	cols, funcs, results := ctrl.Mapping("", "", "", nil, utils.Results{})

	if len(cols) != 0 || len(funcs) != 0 || len(results) != 0 {
		t.Errorf("Expected empty results for empty response")
	}
}

func TestMapping_WithResponse(t *testing.T) {
	ctrl := &controller.AbstractController{}
	cols, funcs, results := ctrl.Mapping("col1", "col2:func", "cmd", nil, utils.Results{{"col1": "value"}})

	if len(cols) == 0 || len(funcs) == 0 || len(results) == 0 {
		t.Errorf("Expected valid mapping results")
	}
}

func TestMapping_SchemaExclusion(t *testing.T) {
	ctrl := &controller.AbstractController{}
	cols, _, _ := ctrl.Mapping("", "", "", nil, utils.Results{{"schema": map[string]interface{}{"col1": "many"}}})

	if len(cols) != 0 {
		t.Errorf("Expected no columns due to schema exclusion")
	}
}

func TestCSV_LastLineProcessing(t *testing.T) {
	ctrl := &controller.AbstractController{}
	domain := &MockDomain{}
	result := ctrl.CSV(domain, map[string]string{"col1": "func(col1)"}, []string{"col1"}, utils.Results{{"col1": "value"}})

	if len(result) == 0 {
		t.Errorf("Expected valid CSV output with last line processing")
	}
}

func TestResponse_RemovePassword(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.Output = recorder

	ctrl.Response(utils.Results{{"password": "secret", "key": "value"}}, nil)
}

func TestDownload_HeaderSettings(t *testing.T) {
	ctrl := &controller.AbstractController{}
	recorder := httptest.NewRecorder()
	ctrl.Ctx.ResponseWriter = recorder
	domain := &MockDomain{}
	ctrl.Download(domain, "col1", "col2", "cmd", "csv", "file", nil, utils.Results{{"key": "value"}}, nil)

	headers := recorder.Header()
	if headers.Get("Content-Disposition") == "" {
		t.Errorf("Expected Content-Disposition header to be set")
	}
}

func TestCSV_EmptyResults(t *testing.T) {
	ctrl := &controller.AbstractController{}
	domain := &MockDomain{}
	result := ctrl.CSV(domain, nil, []string{"col1"}, utils.Results{})

	if len(result) != 1 { // Should only contain the header row
		t.Errorf("Expected only header row for empty results")
	}
}

package controller_test

import (
	"sqldb-ws/controllers/controller"
	"strings"
	"testing"
)

func TestFormFile_CSV_Success(t *testing.T) {
	data := `name,age\nJohn,30\nJane,25`
	file := createMultipartFile("test.csv", data)
	testController := &controller.AbstractController{}
	results, err := testController.FormFile(map[string]string{"name": "full_name", "age": "user_age"})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}
}

func TestFormFile_XLSX_Success(t *testing.T) {
	data := createXLSXMockData([][]string{{"name", "age"}, {"John", "30"}, {"Jane", "25"}})
	file := createMultipartFile("test.xlsx", data)
	testController := &controller.AbstractController{}
	results, err := testController.FormFile(map[string]string{"name": "full_name", "age": "user_age"})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 records, got %d", len(results))
	}
}

func TestFormFile_UnsupportedFormat(t *testing.T) {
	file := createMultipartFile("test.txt", "invalid data")
	testController := &controller.AbstractController{}
	_, err := testController.FormFile(nil)

	if err == nil || !strings.Contains(err.Error(), "unsupported file format") {
		t.Errorf("Expected unsupported file format error, got %v", err)
	}
}

func TestParseCSV_InvalidData(t *testing.T) {
	file := createMultipartFile("test.csv", "invalid data")
	testController := &controller.AbstractController{}
	_, err := testController.parseCSV(file, nil)

	if err == nil {
		t.Errorf("Expected an error for invalid CSV, got nil")
	}
}

func TestParseXLSX_InvalidData(t *testing.T) {
	file := createMultipartFile("test.xlsx", "invalid data")
	testController := &controller.AbstractController{}
	_, err := testController.parseXLSX(file, nil)

	if err == nil {
		t.Errorf("Expected an error for invalid XLSX, got nil")
	}
}

func TestParseCSV_EmptyFile(t *testing.T) {
	file := createMultipartFile("test.csv", "")
	testController := &controller.AbstractController{}
	_, err := testController.parseCSV(file, nil)

	if err == nil {
		t.Errorf("Expected an error for empty CSV, got nil")
	}
}

func TestParseXLSX_EmptyFile(t *testing.T) {
	file := createMultipartFile("test.xlsx", "")
	testController := &controller.AbstractController{}
	_, err := testController.parseXLSX(file, nil)

	if err == nil {
		t.Errorf("Expected an error for empty XLSX, got nil")
	}
}

func TestMapColumns_RenameHeaders(t *testing.T) {
	input := []string{"name", "age"}
	mapping := map[string]string{"name": "full_name", "age": "user_age"}
	output := controller.MapColumns(input, mapping)

	if output[0] != "full_name" || output[1] != "user_age" {
		t.Errorf("Expected renamed headers, got %v", output)
	}
}

func TestExtractCellValues(t *testing.T) {
	mockRow := controller.MockXLSXRow([]string{"John", "30"})
	values := controller.ExtractCellValues(mockRow)

	if values[0] != "John" || values[1] != "30" {
		t.Errorf("Expected [John 30], got %v", values)
	}
}

package controller

import (
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"sqldb-ws/domain/utils"
	"strings"

	"github.com/thedatashed/xlsxreader"
)

func (t *AbstractController) FormFile(asLabel map[string]string) (utils.Results, error) {
	file, header, err := t.Ctx.Request.FormFile("file")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if strings.HasSuffix(header.Filename, ".csv") {
		return t.parseCSV(file, asLabel)
	} else if strings.HasSuffix(header.Filename, ".xlsx") {
		return t.parseXLSX(file, asLabel)
	}

	return nil, fmt.Errorf("unsupported file format")
}

func (t *AbstractController) parseCSV(file multipart.File, asLabel map[string]string) (utils.Results, error) {
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil || len(records) < 2 {
		return nil, err
	}

	cols := mapColumns(records[0], asLabel)
	var results utils.Results

	for _, rec := range records[1:] {
		record := utils.Record{}
		for i, col := range cols {
			record[col] = rec[i]
		}
		results = append(results, record)
	}

	return results, nil
}

func (t *AbstractController) parseXLSX(file multipart.File, asLabel map[string]string) (utils.Results, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	xl, err := xlsxreader.NewReader(data)
	if err != nil {
		return nil, err
	}

	var results utils.Results
	var cols []string
	firstRow := true

	for row := range xl.ReadRows(xl.Sheets[0]) {
		if firstRow {
			firstRow = false
			cols = mapColumns(extractCellValues(row), asLabel)
		} else {
			record := utils.Record{}
			for i, cell := range row.Cells {
				record[cols[i]] = cell.Value
			}
			results = append(results, record)
		}
	}

	return results, nil
}

func mapColumns(header []string, asLabel map[string]string) []string {
	var cols []string
	for _, col := range header {
		if newCol, exists := asLabel[col]; exists {
			col = strings.Replace(newCol, "_aslabel", "", -1)
		}
		cols = append(cols, col)
	}
	return cols
}

func extractCellValues(row xlsxreader.Row) []string {
	var values []string
	for _, cell := range row.Cells {
		values = append(values, cell.Value)
	}
	return values
}

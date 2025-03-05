package connector

import (
	"database/sql"
	"fmt"
	"strconv"
)

/*
* RowResultToMap converts a row result to a map.
 */
func (db *Database) RowResultToMap(rows *sql.Rows,
	columnNames []string,
	columnType map[string]string) (map[string]interface{}, error) {
	columnPointers := make([]interface{}, len(columnNames))
	for i := range columnPointers {
		columnPointers[i] = new(interface{})
	}
	if err := rows.Scan(columnPointers...); err != nil {
		return nil, err
	}
	rowMap := map[string]interface{}{}
	for i, colName := range columnNames {
		rowMap[colName] = db.parseColumnValue(columnType[colName], columnPointers[i].(*interface{}))
	}
	return rowMap, nil
}

/*
* parseColumnValue converts the column value to the appropriate type.
 */
func (db *Database) parseColumnValue(colType string, val *interface{}) interface{} {
	if val == nil || *val == nil {
		return nil
	}
	strVal := fmt.Sprintf("%v", *val)
	switch colType {
	case "MONEY", "NUMERIC", "DECIMAL", "DOUBLE", "FLOAT":
		if num, err := strconv.ParseFloat(strVal, 64); err == nil {
			return num
		}
	case "TIMESTAMP", "DATE":
		if len(strVal) > 10 {
			return strVal[:10]
		}
	}
	return strVal
}

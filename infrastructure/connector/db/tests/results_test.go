package connector_test

import (
	"database/sql"
	"sqldb-ws/infrastructure/connector"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRowResultToMap(t *testing.T) {
	columns := []string{"id", "name", "price"}
	columnTypes := map[string]string{"id": "INT", "name": "TEXT", "price": "FLOAT"}

	mockRows := &sql.Rows{} // Ideally, use a mock SQL library or interface
	db := &connector.Database{}

	result, err := db.RowResultToMap(mockRows, columns, columnTypes)
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestParseColumnValue(t *testing.T) {
	db := &connector.Database{}

	var val interface{}
	val = []uint8("123.45")
	assert.Equal(t, 123.45, db.ParseColumnValue("FLOAT", &val))

	val = nil
	assert.Nil(t, db.ParseColumnValue("TEXT", &val))
}

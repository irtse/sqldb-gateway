package utils_test

import (
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddMap(t *testing.T) {
	base := map[string]interface{}{"a": 1}
	add := map[string]interface{}{"b": 2}
	result := utils.AddMap(base, add)
	assert.Equal(t, map[string]interface{}{"a": 1, "b": 2}, result)
}

func TestToListStr(t *testing.T) {
	input := []interface{}{123, "abc", nil}
	result := utils.ToListStr(input)
	assert.Equal(t, []string{"123", "abc", ""}, result)
}

func TestToResult(t *testing.T) {
	input := []map[string]interface{}{{"key": "value"}}
	result := utils.ToResult(input)
	assert.Len(t, result, 1)
	assert.Equal(t, "value", result[0]["key"])
}

func TestToRecord(t *testing.T) {
	input := map[string]interface{}{"key": "value"}
	result := utils.ToRecord(input)
	assert.Equal(t, input, result)
}

func TestRecord_Add(t *testing.T) {
	rec := utils.Record{}
	rec.Add("key", "value", func() bool { return true })
	assert.Equal(t, "value", rec["key"])
}

func TestRecord_Copy(t *testing.T) {
	rec := utils.Record{"key": "value"}
	copy := rec.Copy()
	assert.Equal(t, rec, copy)
}

func TestRecord_GetString(t *testing.T) {
	rec := utils.Record{"key": "value"}
	assert.Equal(t, "value", rec.GetString("key"))
	assert.Equal(t, "", rec.GetString("missing"))
}

func TestGetString(t *testing.T) {
	rec := map[string]interface{}{"key": "value"}
	assert.Equal(t, "value", utils.GetString(rec, "key"))
}

func TestGetInt(t *testing.T) {
	rec := map[string]interface{}{"key": "123"}
	assert.Equal(t, int64(123), utils.GetInt(rec, "key"))
}

func TestRecord_GetInt(t *testing.T) {
	rec := utils.Record{"key": "123"}
	assert.Equal(t, int64(123), rec.GetInt("key"))
}

func TestRecord_GetFloat(t *testing.T) {
	rec := utils.Record{"key": "123.45"}
	assert.Equal(t, 123.45, rec.GetFloat("key"))
}

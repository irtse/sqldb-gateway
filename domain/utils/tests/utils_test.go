package utils_test

import (
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareEnum_NonEnumString(t *testing.T) {
	result := utils.PrepareEnum("test_string")
	assert.Equal(t, "test_string", result)
}

func TestPrepareEnum_EnumString(t *testing.T) {
	result := utils.PrepareEnum("enum('A', 'B', 'C')")
	expected := "enum__A_B_C"
	assert.Equal(t, expected, result)
}

func TestTransformType(t *testing.T) {
	result := utils.TransformType("enum('X', 'Y', 'Z')")
	expected := "enum__x_y_z"
	assert.Equal(t, expected, result)
}

func TestToMap_ValidMap(t *testing.T) {
	input := map[string]interface{}{"key": "value"}
	result := utils.ToMap(input)
	assert.Equal(t, input, result)
}

func TestToMap_InvalidType(t *testing.T) {
	input := "not_a_map"
	result := utils.ToMap(input)
	assert.Empty(t, result)
}

func TestToList_ValidList(t *testing.T) {
	input := []interface{}{1, 2, 3}
	result := utils.ToList(input)
	assert.Equal(t, input, result)
}

func TestToList_InvalidType(t *testing.T) {
	input := "not_a_list"
	result := utils.ToList(input)
	assert.Empty(t, result)
}

func TestToInt64_ValidIntString(t *testing.T) {
	result := utils.ToInt64("123")
	assert.Equal(t, int64(123), result)
}

func TestToInt64_InvalidString(t *testing.T) {
	result := utils.ToInt64("invalid")
	assert.Equal(t, int64(0), result)
}

func TestToString_NilInput(t *testing.T) {
	result := utils.ToString(nil)
	assert.Equal(t, "", result)
}

func TestToString_ValidInput(t *testing.T) {
	result := utils.ToString(123)
	assert.Equal(t, "123", result)
}

func TestCompare_EqualValues(t *testing.T) {
	assert.True(t, utils.Compare("test", "test"))
}

func TestCompare_DifferentValues(t *testing.T) {
	assert.False(t, utils.Compare("test", "other"))
}

func TestCompare_NilValue(t *testing.T) {
	assert.False(t, utils.Compare(nil, "test"))
}

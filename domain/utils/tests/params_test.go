package utils_test

import (
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams_GetAsArgs(t *testing.T) {
	params := utils.Params{Values: map[string]string{"key": "value"}}
	assert.Equal(t, []string{"value"}, params.GetAsArgs("key"))
	assert.Empty(t, params.GetAsArgs("missing"))
}

func TestParams_Copy(t *testing.T) {
	params := utils.Params{Values: map[string]string{"key": "value"}}
	copy := params.Copy()
	assert.Equal(t, params, copy)
}

func TestParams_Enrich(t *testing.T) {
	params := utils.Params{Values: map[string]string{"key": "old"}}
	enriched := params.Enrich(map[string]interface{}{"key": "new", "extra": "data"})
	assert.Equal(t, "new", enriched.Values["key"])
	assert.Equal(t, "data", enriched.Values["extra"])
}

func TestParams_Delete(t *testing.T) {
	params := utils.Params{Values: map[string]string{"key": "value", "remove": "this"}}
	filtered := params.Delete(func(k string) bool { return k == "remove" })
	assert.NotContains(t, filtered, "remove")
}

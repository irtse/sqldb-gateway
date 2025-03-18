package tests

import (
	"testing"

	domain "sqldb-ws/domain"
	permissions "sqldb-ws/domain/permission"
	"sqldb-ws/domain/utils"

	"github.com/stretchr/testify/assert"
)

func TestDomainInitialization(t *testing.T) {
	d := domain.Domain(true, "admin", nil)
	assert.NotNil(t, d)
	assert.Equal(t, "admin", d.User)
	assert.True(t, d.SuperAdmin)
}

func TestVerifyAuth_NoArgs(t *testing.T) {
	permsService := &permissions.PermDomainService{}
	d := domain.Domain(false, "user", permsService)
	assert.False(t, d.VerifyAuth("table", "column", "read", utils.SELECT))
}

func TestVerifyAuth_WithArgs(t *testing.T) {
	permsService := &permissions.PermDomainService{}
	d := domain.Domain(false, "user", permsService)
	assert.False(t, d.VerifyAuth("table", "column", "read", utils.SELECT, "extra"))
}

func TestHandleRecordAttributes_Empty(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	record := utils.Record{"is_empty": true, "is_list": false, "own_view": true}
	d.HandleRecordAttributes(record)
	assert.True(t, d.Empty)
	assert.False(t, d.LowerRes)
	assert.True(t, d.Own)
}

func TestIsOwn_PermissionCheck(t *testing.T) {
	permsService := &permissions.PermDomainService{}
	d := domain.Domain(false, "user", permsService)
	assert.False(t, d.IsOwn(true, false, utils.SELECT))
}

func TestIsOwn_NoPermissionCheck(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	assert.False(t, d.IsOwn(false, false, utils.SELECT))
}

func TestGetDb_ReturnsNilInitially(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	assert.Nil(t, d.GetDb())
}

func TestCreateSuperCall_FailsWithoutParams(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.CreateSuperCall(nil, nil)
	assert.Error(t, err)
}

func TestUpdateSuperCall_FailsWithoutParams(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.UpdateSuperCall(nil, nil)
	assert.Error(t, err)
}

func TestDeleteSuperCall_FailsWithoutParams(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.DeleteSuperCall(nil)
	assert.Error(t, err)
}

func TestSuperCall_NotAuthorized(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.SuperCall(nil, nil, utils.DELETE, false)
	assert.Error(t, err)
}

func TestCall_NoParams(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.Call(nil, nil, utils.SELECT)
	assert.Error(t, err)
}

func TestOnBooleanValue_EnableKey(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	d.Params = utils.Params{"testKey": "enable"}
	var flag bool
	d.OnBooleanValue("testKey", func(b bool) { flag = b })
	assert.True(t, flag)
}

func TestCall_NoTableName(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	params := utils.Params{}
	_, err := d.Call(params, nil, utils.SELECT)
	assert.Error(t, err)
}

func TestCall_UnauthorizedMethod(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	params := utils.Params{utils.RootTableParam: "users"}
	_, err := d.Call(params, nil, utils.DELETE)
	assert.Error(t, err)
}

func TestInvoke_UnknownMethod(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.Invoke(nil, utils.Method("UNKNOWN"))
	assert.Error(t, err)
}

func TestInvoke_NoService(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	_, err := d.Invoke(nil, utils.CREATE)
	assert.Error(t, err)
}

func TestClearDeprecatedDatas_NoSchema(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	d.ClearDeprecatedDatas("unknown_table")
	// No error expected, just ensuring function runs without failure
}

func TestClearDeprecatedDatas_ValidSchema(t *testing.T) {
	d := domain.Domain(false, "user", nil)
	d.ClearDeprecatedDatas("valid_table")
	// Function should execute without errors
}

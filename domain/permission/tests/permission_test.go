package permission_test

import (
	"sqldb-ws/domain/permission"
	"sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPermDomainService(t *testing.T) {
	db := &connector.Database{}
	service := permission.NewPermDomainService(db, "testUser", true, false)

	assert.NotNil(t, service)
	assert.Equal(t, "testUser", service.User)
	assert.True(t, service.IsSuperAdmin)
	assert.False(t, service.Empty)
}

func TestUserSelectQuery(t *testing.T) {
	db := &connector.Database{}
	service := permission.NewPermDomainService(db, "testUser", false, false)

	query := service.UserSelectQuery()
	assert.Contains(t, query, "SELECT")
	assert.Contains(t, query, "testUser")
}

func TestEntitySelectQuery(t *testing.T) {
	db := &connector.Database{}
	service := permission.NewPermDomainService(db, "testUser", false, false)

	query := service.EntitySelectQuery()
	assert.Contains(t, query, "SELECT")
}

func TestBuildFilterOwnPermsQueryRestriction(t *testing.T) {
	db := &connector.Database{}
	service := permission.NewPermDomainService(db, "testUser", false, false)

	restriction := service.BuildFilterOwnPermsQueryRestriction()
	assert.NotNil(t, restriction)
}

func TestProcessPermissionRecord_ValidData(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", false, false)
	data := map[string]interface{}{
		models.NAMEKEY: "tableName:columnName",
		"read":         "ALL",
		"write":        true,
		"update":       true,
		"delete":       false,
	}
	service.ProcessPermissionRecord(data)
	assert.Equal(t, "ALL", service.Perms["tableName"]["columnName"].Read)
}

func TestProcessPermissionRecord_InvalidData(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", false, false)
	data := map[string]interface{}{models.NAMEKEY: "invalidData"}
	service.ProcessPermissionRecord(data)
	assert.Empty(t, service.Perms)
}

func TestMapPerm(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", false, false)
	perm1 := permission.Perms{Read: "OWN", Create: true, Update: false, Delete: false}
	perm2 := permission.Perms{Read: "ALL", Create: false, Update: true, Delete: true}
	result := service.MapPerm(perm1, perm2)
	assert.True(t, result.Update)
	assert.True(t, result.Delete)
}

func TestIsOwnPermission(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", false, false)
	assert.False(t, service.IsOwnPermission("some_table", true, utils.SELECT))
}

func TestPermsCheck_SuperAdmin(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", true, false)
	assert.True(t, service.PermsCheck("table", "col", "ALL", utils.SELECT))
}

func TestLocalPermsCheck_NoPerms(t *testing.T) {
	service := permission.NewPermDomainService(nil, "", false, false)
	assert.False(t, service.LocalPermsCheck("table", "col", "", utils.UPDATE, ""))
}

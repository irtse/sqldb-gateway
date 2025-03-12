package permission

import (
	"encoding/json"
	"slices"
	schserv "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	conn "sqldb-ws/infrastructure/connector"
	"strings"
	"sync"
)

// DONE - ~ 240 LINES - PARTIALLY TESTED
type Perms struct {
	Read   string `json:"read"`
	Create bool   `json:"write"`
	Update bool   `json:"update"`
	Delete bool   `json:"delete"`
}

type PermDomainService struct {
	mutexPerms   sync.RWMutex
	Perms        map[string]map[string]Perms
	IsSuperAdmin bool
	Empty        bool
	User         string
	db           *conn.Database
}

func NewPermDomainService(db *conn.Database, user string, isSuperAdmin bool, empty bool) *PermDomainService {
	return &PermDomainService{
		mutexPerms:   sync.RWMutex{},
		Perms:        map[string]map[string]Perms{},
		IsSuperAdmin: isSuperAdmin,
		Empty:        empty,
		db:           db,
		User:         user,
	}
}

func (p *PermDomainService) PermsBuilder() {
	filterOwnPermsQueryRestriction := p.BuildFilterOwnPermsQueryRestriction()
	datas, _ := p.db.SelectQueryWithRestriction(ds.DBPermission.Name, filterOwnPermsQueryRestriction, false)
	if len(datas) == 0 {
		return
	}

	p.mutexPerms.Lock()
	defer p.mutexPerms.Unlock()

	for _, record := range datas {
		p.ProcessPermissionRecord(record)
	}
}

func (p *PermDomainService) BuildFilterOwnPermsQueryRestriction() map[string]interface{} {
	return map[string]interface{}{
		"id": conn.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
			ds.DBRole.Name + "_id": p.db.BuildSelectQueryWithRestriction(
				ds.DBRoleAttribution.Name,
				map[string]interface{}{
					ds.DBUser.Name + "_id":   p.UserSelectQuery(),
					ds.DBEntity.Name + "_id": p.EntitySelectQuery(),
				}, true, ds.DBRole.Name+"_id"),
		}, false),
	}
}

func (p *PermDomainService) UserSelectQuery() string {
	return p.db.BuildSelectQueryWithRestriction(
		ds.DBUser.Name,
		map[string]interface{}{
			"name":  conn.Quote(p.User),
			"email": conn.Quote(p.User),
		}, true, "id")
}

func (p *PermDomainService) EntitySelectQuery() string {
	return p.db.BuildSelectQueryWithRestriction(
		ds.DBEntityUser.Name,
		map[string]interface{}{
			ds.DBUser.Name + "_id": p.UserSelectQuery(),
		}, true, ds.DBEntity.Name+"_id",
	)
}

func (p *PermDomainService) ProcessPermissionRecord(record map[string]interface{}) {
	names := strings.Split(utils.ToString(record[sm.NAMEKEY]), ":")
	if len(names) < 2 {
		return
	}
	tName, n := names[0], names[1]
	var perms Perms
	b, _ := json.Marshal(record)
	json.Unmarshal(b, &perms)

	if p.Perms[tName] == nil {
		p.Perms[tName] = make(map[string]Perms)
	}

	perm := p.Perms[tName][n]
	if slices.Index(sm.READLEVELACCESS, perms.Read) > slices.Index(sm.READLEVELACCESS, perm.Read) {
		perm.Read = perms.Read
	}
	perm = p.MapPerm(perm, perms)

	p.Perms[tName][n] = perm
}

func (p *PermDomainService) MapPerm(perm Perms, perms Perms) Perms {
	perm.Create = perms.Create
	perm.Update = perms.Update
	perm.Delete = perms.Delete
	return perm
}

func (p *PermDomainService) exception(tableName string, force bool, method utils.Method) bool {
	if !force {
		return false
	}
	return slices.Contains(ds.OWNPERMISSIONEXCEPTION, tableName) ||
		slices.Contains(ds.AllPERMISSIONEXCEPTION, tableName) ||
		(slices.Contains(ds.PERMISSIONEXCEPTION, tableName) && method == utils.SELECT) ||
		(slices.Contains(ds.PUPERMISSIONEXCEPTION, tableName) && method == utils.UPDATE) ||
		(slices.Contains(ds.POSTPERMISSIONEXCEPTION, tableName) && method == utils.CREATE)
}

func (p *PermDomainService) IsOwnPermission(tableName string, force bool, method utils.Method) bool {
	if p.exception(tableName, !force, method) || method != utils.SELECT {
		return slices.Contains(ds.OWNPERMISSIONEXCEPTION, tableName)
	}
	if len(p.Perms) == 0 {
		p.PermsBuilder()
	}
	p.mutexPerms.Lock()
	defer p.mutexPerms.Unlock()
	if tPerms, ok := p.Perms[tableName]; ok {
		return tPerms[tableName].Read == sm.LEVELOWN
	}
	return false
}

// can redact a view based on perms.
func (p *PermDomainService) PermsCheck(tableName string, colName string, level string, method utils.Method) bool {
	return p.LocalPermsCheck(tableName, colName, level, method, "")
}
func (p *PermDomainService) LocalPermsCheck(tableName string, colName string, level string, method utils.Method, destID string) bool {
	// Super admin override or exception handling
	if p.IsSuperAdmin || p.exception(tableName, level == "" || level == "<nil>" || level == sm.LEVELNORMAL, method) {
		return true
	}
	// Build permissions if empty
	if len(p.Perms) == 0 {
		p.PermsBuilder()
	}
	// Retrieve permissions
	p.mutexPerms.Lock()
	perms := p.getPermissions(tableName, colName)
	p.mutexPerms.Unlock()

	// Handle SELECT method permissions
	if method == utils.SELECT {
		return p.hasReadAccess(level, perms.Read)
	}
	// Handle UPDATE and CREATE permissions
	if (method == utils.UPDATE && perms.Update) || (method == utils.CREATE && perms.Create) {
		return p.checkUpdateCreatePermissions(tableName, destID)
	}
	// Handle DELETE permissions
	return method == utils.DELETE && perms.Delete
}

func (p *PermDomainService) getPermissions(tableName, colName string) Perms {
	if tPerms, ok := p.Perms[tableName]; ok {
		if cPerms, ok2 := tPerms[colName]; ok2 && colName != "" {
			return cPerms
		}
		return p.aggregatePermissions(tPerms, tableName)
	}
	return Perms{}
}

func (p *PermDomainService) aggregatePermissions(tPerms map[string]Perms, tableName string) Perms {
	perms := p.Perms[tableName][tableName]
	for _, perm := range tPerms {
		p.MapPerm(perm, perms)
	}
	return perms
}

func (p *PermDomainService) hasReadAccess(level, readPerm string) bool {
	if slices.Contains(sm.READLEVELACCESS, level) && level != sm.LEVELNORMAL {
		return p.compareAccessLevels(level, readPerm)
	}
	return readPerm == sm.LEVELNORMAL || readPerm == sm.LEVELOWN
}

func (p *PermDomainService) compareAccessLevels(level, readPerm string) bool {
	levelCount, _ := p.accessLevelIndex(level)
	compareCount, foundCompare := p.accessLevelIndex(readPerm)
	return compareCount >= levelCount && foundCompare
}

func (p *PermDomainService) accessLevelIndex(targetLevel string) (int, bool) {
	count := 0
	found := false
	for _, l := range sm.READLEVELACCESS {
		if l == targetLevel {
			found = true
			break
		} else if !found {
			count++
		}
	}
	return count, found
}

func (p *PermDomainService) checkUpdateCreatePermissions(tableName, destID string) bool {
	if p.Empty || destID == "" {
		return true
	}
	schema, _ := schserv.GetSchema(tableName)
	res, err := p.db.SimpleMathQuery("COUNT", ds.DBRequest.Name, map[string]interface{}{
		utils.RootDestTableIDParam: destID,
		ds.SchemaDBField:           utils.ToString(schema.ID),
	}, false)
	if err != nil || len(res) == 0 || res[0]["result"] == nil || utils.ToInt64(res[0]["result"]) == 0 {
		return false
	}
	res, err = p.db.SimpleMathQuery("COUNT", ds.DBTask.Name, map[string]interface{}{
		utils.RootDestTableIDParam: destID,
		ds.UserDBField:             p.UserSelectQuery(),
		ds.EntityDBField:           p.EntitySelectQuery(),
	}, true)
	return err == nil && len(res) > 0 && res[0]["result"] != nil && utils.ToInt64(res[0]["result"]) > 0
}

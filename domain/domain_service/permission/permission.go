package permission

import (
	"encoding/json"
	"fmt"
	"slices"
	"sqldb-ws/domain/schema"
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
	mutexPerms        sync.RWMutex
	AlreadyCheckPerms map[string]map[utils.Method]bool
	Perms             map[string]map[string]Perms
	IsSuperAdmin      bool
	Empty             bool
	User              string
	db                *conn.Database
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

func (p *PermDomainService) PermsBuilder(userID string) {
	filterOwnPermsQueryRestriction := p.BuildFilterOwnPermsQueryRestriction(userID)
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

func (p *PermDomainService) BuildFilterOwnPermsQueryRestriction(userID string) map[string]interface{} {
	return map[string]interface{}{
		"id": p.db.BuildSelectQueryWithRestriction(
			ds.DBRolePermission.Name,
			map[string]interface{}{
				ds.DBRole.Name + "_id": p.db.BuildSelectQueryWithRestriction(
					ds.DBRoleAttribution.Name,
					map[string]interface{}{
						ds.DBUser.Name + "_id":   userID,
						ds.DBEntity.Name + "_id": p.EntitySelectQuery(userID),
					}, true, ds.DBRole.Name+"_id"),
			}, false, ds.DBPermission.Name+"_id"),
	}
}

func (p *PermDomainService) EntitySelectQuery(myUserID string) string {
	return p.db.BuildSelectQueryWithRestriction(
		ds.DBEntityUser.Name,
		map[string]interface{}{
			ds.DBUser.Name + "_id": myUserID,
		}, true, ds.DBEntity.Name+"_id",
	)
}

func (p *PermDomainService) ProcessPermissionRecord(record map[string]interface{}) {
	names := strings.Split(utils.ToString(record[sm.NAMEKEY]), ":")
	if len(names) < 2 {
		return
	}
	tName, n := names[0], names[1]
	if len(names) < 4 {
		n = names[0]
	}
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

func (p *PermDomainService) exception(tableName string, force bool, method utils.Method, isOwn bool) bool {
	if !force {
		return false
	}
	return slices.Contains(ds.OWNPERMISSIONEXCEPTION, tableName) && isOwn ||
		slices.Contains(ds.AllPERMISSIONEXCEPTION, tableName) ||
		(slices.Contains(ds.PERMISSIONEXCEPTION, tableName) && method == utils.SELECT) ||
		(slices.Contains(ds.PUPERMISSIONEXCEPTION, tableName) && method == utils.UPDATE) ||
		(slices.Contains(ds.POSTPERMISSIONEXCEPTION, tableName) && method == utils.CREATE)
}

func (p *PermDomainService) IsOwnPermission(tableName string, force bool, method utils.Method, myUserID string, isOwn bool) bool {
	if p.exception(tableName, !force, method, isOwn) || method != utils.SELECT {
		return slices.Contains(ds.OWNPERMISSIONEXCEPTION, tableName)
	}
	if len(p.Perms) == 0 {
		p.PermsBuilder(myUserID)
	}
	p.mutexPerms.Lock()
	defer p.mutexPerms.Unlock()
	if tPerms, ok := p.Perms[tableName]; ok {
		return tPerms[tableName].Read == sm.LEVELOWN
	}
	return false
}

// can redact a view based on perms.
func (p *PermDomainService) PermsCheck(tableName string, colName string, level string, method utils.Method, myUserID string, isOwn bool) bool {
	return p.LocalPermsCheck(tableName, colName, level, method, "", myUserID, isOwn)
}
func (p *PermDomainService) LocalPermsCheck(tableName string, colName string, level string, method utils.Method, destID string, myUserID string, isOwn bool) bool {
	// Super admin override or exception handling
	if p.IsSuperAdmin || p.exception(tableName, level == "" || level == "<nil>" || level == sm.LEVELNORMAL, method, isOwn) {
		return true
	}

	// Build permissions if empty
	if len(p.Perms) == 0 {
		p.PermsBuilder(myUserID)
	}
	// Retrieve permissions
	p.mutexPerms.Lock()
	perms := p.getPermissions(tableName, colName)
	p.mutexPerms.Unlock()
	// Handle SELECT method permissions
	schema, err := schserv.GetSchema(tableName)
	if err != nil {
		return false
	}
	accesGranted := true
	if already, ok := p.AlreadyCheckPerms[tableName+":"+colName]; ok {
		if granted, ok := already[method]; ok {
			fmt.Println("ALREADY GRANTED", granted)

			return granted
		}
	}
	if method == utils.SELECT && !p.hasReadAccess(level, perms.Read) {
		accesGranted = p.getShare(schema, destID, "read_access", true)
	}
	// Handle UPDATE and CREATE permissions
	if method == utils.CREATE && !perms.Create {
		accesGranted = p.getShare(schema, destID, "create_access", true)
	}
	if method == utils.UPDATE && !perms.Update {
		if !p.checkUpdateCreatePermissions(tableName, destID, myUserID) {
			accesGranted = p.getShare(schema, destID, "update_access", true)
		}
	}
	if method == utils.DELETE && !perms.Delete {
		accesGranted = p.getShare(schema, destID, "delete_access", true)
	}
	if p.AlreadyCheckPerms[tableName+":"+colName] == nil {
		p.AlreadyCheckPerms[tableName+":"+colName] = map[utils.Method]bool{}
	}
	p.AlreadyCheckPerms[tableName+":"+colName][method] = accesGranted
	// Handle DELETE permissions
	return accesGranted
}

func (p *PermDomainService) getShare(schema sm.SchemaModel, destID string, key string, val bool) bool {
	if destID == "" {
		return false
	}
	res, err := p.db.SelectQueryWithRestriction(ds.DBShare.Name, map[string]interface{}{
		ds.UserDBField: p.db.BuildSelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			"name":  p.User,
			"email": p.User,
		}, true, "id"),
		ds.SchemaDBField:    schema.ID,
		ds.DestTableDBField: destID,
		key:                 val,
	}, false)
	return err == nil && len(res) > 0
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

func (p *PermDomainService) checkUpdateCreatePermissions(tableName, destID string, myUserID string) bool {
	if p.Empty || destID == "" {
		return true
	}
	sch, e := schema.GetSchema(tableName)
	if e != nil {
		return false
	}
	if res, err := p.db.SimpleMathQuery("COUNT", ds.DBDataAccess.Name, map[string]interface{}{
		ds.SchemaDBField:           sch.ID,
		utils.RootDestTableIDParam: destID,
		ds.UserDBField:             myUserID,
		ds.EntityDBField:           p.EntitySelectQuery(myUserID),
		"write":                    true,
	}, true); err == nil && len(res) > 0 && res[0]["result"] != nil && utils.ToInt64(res[0]["result"]) > 0 {
		if res, err := p.db.SimpleMathQuery("COUNT", ds.DBRequest.Name, map[string]interface{}{
			ds.SchemaDBField:           sch.ID,
			utils.RootDestTableIDParam: destID,
			"is_close":                 false,
		}, true); err == nil && len(res) > 0 && res[0]["result"] != nil && utils.ToInt64(res[0]["result"]) > 0 {
			return true
		}
	}
	res, err := p.db.SimpleMathQuery("COUNT", ds.DBTask.Name, map[string]interface{}{
		utils.SpecialIDParam: p.db.BuildSelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
			ds.UserDBField:   myUserID,
			ds.EntityDBField: p.EntitySelectQuery(myUserID),
		}, true),
		ds.SchemaDBField:           sch.ID,
		utils.RootDestTableIDParam: destID,
		"is_close":                 false,
	}, false)
	return err == nil && len(res) > 0 && res[0]["result"] != nil && utils.ToInt64(res[0]["result"]) > 0
}

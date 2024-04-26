package domain

import (
	"fmt"
	"sync"
	"slices"
	"strings"
	"encoding/json"
	"sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type Perms struct {
	Read   string 	`json:"read"`
	Create bool 	`json:"write"`
	Update bool 	`json:"update"`
	Delete bool 	`json:"delete"`
}
var mutexPerms  = sync.RWMutex{}
func (d *MainService) PermsBuilder() {
	d.Perms = map[string]map[string]Perms{}
	d.Db.SQLRestriction = "id IN (SELECT " + schserv.DBPermission.Name + "_id FROM " 
	d.Db.SQLRestriction += schserv.DBRolePermission.Name + " WHERE " + schserv.DBRole.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT " + schserv.DBRole.Name + "_id FROM " 
	d.Db.SQLRestriction += schserv.DBRoleAttribution.Name + " WHERE " + schserv.DBUser.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT id FROM " + schserv.DBUser.Name + " WHERE " 
	d.Db.SQLRestriction += "name=" + conn.Quote(d.GetUser()) + " OR email=" + conn.Quote(d.GetUser()) + ") OR " + schserv.DBEntity.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT " + schserv.DBEntity.Name + "_id FROM "
	d.Db.SQLRestriction += schserv.DBEntityUser.Name + " WHERE " + schserv.DBUser.Name +"_id IN ("
	d.Db.SQLRestriction += "SELECT id FROM " + schserv.DBUser.Name + " WHERE "
	d.Db.SQLRestriction += "name=" + conn.Quote(d.GetUser()) + " OR email=" + conn.Quote(d.GetUser()) + "))))"
	d.Db.SQLView = ""
	datas, err := d.Db.SelectResults(schserv.DBPermission.Name)
	if err != nil || len(datas) == 0 { return }
	d.Db.SQLRestriction = ""
	for _, record := range datas {
		names := strings.Split(fmt.Sprintf("%v", record[schserv.NAMEKEY]), ":")
		names = names[:len(names)-1]
		if len(names) == 0 { continue }
		tName := names[0]
		mutexPerms.Lock()
		var perms Perms
		b, _ := json.Marshal(record)
		json.Unmarshal(b, &perms)
		n := ""
		if len(names) > 1 { n = names[1] } else { n = names[0] }		
		if p, ok := d.Perms[tName]; !ok || p == nil { d.Perms[tName] = map[string]Perms{}; }
		if _, ok := d.Perms[tName][n]; !ok {  d.Perms[tName][n]=perms }
		p := d.Perms[tName][n]
		if slices.Index(schserv.READLEVELACCESS, perms.Read) > slices.Index(schserv.READLEVELACCESS, p.Read) { p.Read = perms.Read }
		if perms.Create { p.Create = true }
		if perms.Update { p.Update = true }
		if perms.Delete { p.Delete = true }
		mutexPerms.Unlock()
	}
}
func (d *MainService) exception(tableName string, force bool, method utils.Method) bool {
	if force {
		if slices.Contains(schserv.OWNPERMISSIONEXCEPTION, tableName) { return true }
		if slices.Contains(schserv.AllPERMISSIONEXCEPTION, tableName) { return true }
		if slices.Contains(schserv.PERMISSIONEXCEPTION, tableName) && method == utils.SELECT { return true }
		if slices.Contains(schserv.PUPERMISSIONEXCEPTION, tableName) && method == utils.UPDATE { return true }
		if slices.Contains(schserv.POSTPERMISSIONEXCEPTION, tableName) && method == utils.CREATE { return true }
	}
	return false
}

func (d *MainService) IsOwnPermission(tableName string, force bool, method utils.Method) bool {
	if d.exception(tableName, !force, method) || method != utils.SELECT { return slices.Contains(schserv.OWNPERMISSIONEXCEPTION, tableName) }
	if len(d.Perms) == 0 { d.PermsBuilder() }
	mutexPerms.Lock()
	defer mutexPerms.Unlock()
	if tPerms, ok := d.Perms[tableName]; ok { 
		return tPerms[tableName].Read == schserv.LEVELOWN
	}
	return false
}
// can redact a view based on perms. 
func (d *MainService) PermsCheck(tableName string, colName string, level string, method utils.Method) bool {
	if d.SuperAdmin && method != utils.SELECT || method == utils.SELECT && d.IsSuperCall() { return true }
	if d.exception(tableName, level == "" || fmt.Sprintf("%v", level) == "<nil>" || level == schserv.LEVELNORMAL, method) { return true }
	if len(d.Perms) == 0 { d.PermsBuilder() }
	var perms Perms
	mutexPerms.Lock()
	if tPerms, ok := d.Perms[tableName]; ok {
		if cPerms, ok2 := tPerms[colName]; ok2 && colName != "" && level != "" { perms = cPerms 
		} else { 
			perms = d.Perms[tableName][tableName]
			if colName == "" {
				for _, p := range tPerms {
					if p.Create { perms.Create = true }
					if p.Update { perms.Update = true }
					if p.Delete { perms.Delete = true }
				}
			}
		}
	}
	mutexPerms.Unlock()
	if method == utils.SELECT {
		isTableNormal := perms.Read == "" || perms.Read == "<nil>"
		if !isTableNormal && slices.Contains(schserv.READLEVELACCESS, level) && level != schserv.LEVELNORMAL {
			levelCount := 0; found := false;
			compareCount := 0; foundCompare := false
			for _, l := range schserv.READLEVELACCESS {
				if l == level && !found { found=true } else if !found { levelCount++; }
				if l == perms.Read && !foundCompare { foundCompare=true } else if !foundCompare { compareCount++; }
			}
			return compareCount >= levelCount && foundCompare
		}
		return isTableNormal || perms.Read == schserv.LEVELNORMAL || perms.Read == schserv.LEVELOWN
	}
	return (method == utils.CREATE && perms.Create) || (method == utils.UPDATE && perms.Update) || (method == utils.DELETE && perms.Delete)
}
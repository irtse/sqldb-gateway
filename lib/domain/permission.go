package domain

import (
	"fmt"
	"sync"
	"slices"
	"strings"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
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
	d.Db.SQLRestriction = "id IN (SELECT " + entities.DBPermission.Name + "_id FROM " 
	d.Db.SQLRestriction += entities.DBRolePermission.Name + " WHERE " + entities.DBRole.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT " + entities.DBRole.Name + "_id FROM " 
	d.Db.SQLRestriction += entities.DBRoleAttribution.Name + " WHERE " + entities.DBUser.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT id FROM " + entities.DBUser.Name + " WHERE " 
	d.Db.SQLRestriction += "name=" + conn.Quote(d.GetUser()) + " OR email=" + conn.Quote(d.GetUser()) + ") OR " + entities.DBEntity.Name + "_id IN ("
	d.Db.SQLRestriction += "SELECT " + entities.DBEntity.Name + "_id FROM "
	d.Db.SQLRestriction += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
	d.Db.SQLRestriction += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
	d.Db.SQLRestriction += "name=" + conn.Quote(d.GetUser()) + " OR email=" + conn.Quote(d.GetUser()) + "))))"
	datas, err := d.Db.SelectResults(entities.DBPermission.Name)
	if err != nil || len(datas) == 0 { return }
	d.Db.SQLRestriction = ""
	for _, record := range datas {
		names := strings.Split(fmt.Sprintf("%v", record[entities.NAMEATTR]), ":")
		names = names[:len(names)-1]
		if len(names) == 0 { continue }
		tName := names[0]
		
		var perms Perms
		b, _ := json.Marshal(record)
		json.Unmarshal(b, &perms)
		n := ""
		if len(names) > 1 { n = names[1] 
		} else { n = names[0] }		
		
		mutexPerms.Lock()
		if p, ok := d.Perms[tName]; !ok || p == nil { d.Perms[tName] = map[string]Perms{}; }
		p := d.Perms[tName][n]
		if _, ok := d.Perms[tName][n]; !ok { 
			d.Perms[tName][n]=perms
		} else {
			if slices.Index(entities.READLEVELACCESS, perms.Read) > slices.Index(entities.READLEVELACCESS, p.Read) { 
				p.Read = perms.Read 
			}
			if perms.Create { p.Create = true }
			if perms.Update { p.Update = true }
			if perms.Delete { p.Delete = true }
			d.Perms[tName][n]=p
		}
		mutexPerms.Unlock()
	}
}
// can redact a view based on perms. 
func (d *MainService) PermsCheck(tableName string, colName string, level string, method tool.Method) bool {
	if d.SuperAdmin && method != tool.SELECT || method == tool.SELECT && d.IsSuperCall() { return true }
	if level == "" || fmt.Sprintf("%v", level) == "<nil>" || level == entities.LEVELNORMAL {
		if method == tool.SELECT {
			for _, exception := range entities.PERMISSIONEXCEPTION { // permission exception allows any read
				if tableName == exception.Name { return true }
			}
		}
		if method == tool.UPDATE {
			for _, exception := range entities.PUPERMISSIONEXCEPTION { // permission exception allows any read
				if tableName == exception.Name { return true }
			}
		}
		if method == tool.CREATE {
			for _, exception := range entities.POSTPERMISSIONEXCEPTION { // permission exception allows any read
				if tableName == exception.Name { return true }
			}
		}
	}
	if len(d.Perms) == 0 { d.PermsBuilder() }
	var perms Perms
	found := false
	if tPerms, ok := d.Perms[tableName]; ok {
		found = true
		if cPerms, ok2 := tPerms[colName]; ok2 && colName != "" && level != "" { 
			perms = cPerms 
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
	if !found { return false }
	if method == tool.SELECT {
		isTableNormal := perms.Read == "" || perms.Read == "<nil>"
		if !isTableNormal && slices.Contains(entities.READLEVELACCESS, level) && level != entities.LEVELNORMAL {
			levelCount := 0; found := false;
			compareCount := 0; foundCompare := false
			for _, l := range entities.READLEVELACCESS {
				if l == level && !found { found=true 
				} else if !found { levelCount++; }
				if l == perms.Read && !foundCompare { foundCompare=true 
				} else if !foundCompare { compareCount++; }
			}
			return compareCount >= levelCount && foundCompare
		}
		return isTableNormal || perms.Read == entities.LEVELNORMAL
	}
	var permsRecord tool.Record
	b, _ := json.Marshal(perms)
	json.Unmarshal(b, &permsRecord)
	return permsRecord[method.String()].(bool)
}
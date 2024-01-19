package service

import (
	"errors"
	"strings"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)
var ADMINROLE = "admin"
var WRITEROLE = "manager"
var CREATEROLE = "creator"
var UPDATEROLE = "updater"
var READERROLE = "reader"

var DBrole = "dbrole"
var DBuser = "dbuser"
var DBentity = "dbentity"
var DBentityuser = "dbentityuser"
var DBroleentity = "dbroleentity"
var DBrolepermission = "dbrolepermission"
var PERMS = []string{entities.CREATEPERMS, entities.UPDATEPERMS, entities.DELETEPERMS, entities.READPERMS}

var MAIN_PERMS=map[string]map[string]bool{
	ADMINROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: true, entities.DELETEPERMS: true, entities.READPERMS: true, },
    WRITEROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: true, entities.DELETEPERMS: false, entities.READPERMS: true, },
	CREATEROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: false, entities.DELETEPERMS: false, entities.READPERMS: true, },
	UPDATEROLE: map[string]bool{ entities.CREATEPERMS : false, entities.UPDATEPERMS: true, entities.DELETEPERMS: false, entities.READPERMS: true, },
	READERROLE: map[string]bool{ entities.CREATEPERMS : false, entities.UPDATEPERMS: false, entities.DELETEPERMS: false, entities.READPERMS: true, },
}

type Info struct { 
	Info    string				`json:"info"`
	Name    string 				`json:"name"`
	Results tool.Results		`json:"results"`
}

type PermissionInfo struct {
	WarningUpdateField    []string
	PartialResults        string
	Perms     		      map[string]tool.Record
	Row		      		  *TableRowInfo
	db        		      *conn.Db
	InfraService
}

func (p *PermissionInfo) generatePerms(res tool.Results) {
	del := []string{}
	for _, r := range res[1:] {
		if r[entities.COLNAMEATTR] == nil { p.generatePerm(r[entities.TABLENAMEATTR].(string), r, del)
		} else { p.generatePerm(r[entities.TABLENAMEATTR].(string) + ":" + r[entities.COLNAMEATTR].(string), r, del) }
	}
}

func (p *PermissionInfo) generatePerm(name string, record tool.Record, del []string) {
	if permission, ok := p.Perms[name]; ok {
		for _, perm :=  range PERMS {
			if valid, ok1 := record[perm]; ok1 && valid.(bool) && !permission[perm].(bool) { 
				del = append(del, permission[entities.NAMEATTR].(string) )
				permission[entities.NAMEATTR]=record[entities.NAMEATTR] 
				permission[perm]=true 
			}
		}
	} else { p.Perms[name]= record }
}

func (p *PermissionInfo) Template() (interface{}, error) { return p.Get() }
// todo view (columns sort of)
func (p *PermissionInfo) Verify(name string) (string, bool) {
	if p.SuperAdmin { return name, true }
	view := []string{}
	authorized := false
	if tperms, ok := p.Perms[name]; ok {
		if valid, ok2 := tperms[p.Method.String()]; ok2 && valid.(bool) { authorized = true }
	}
	delKeyRec := []string{}
	for k, _ := range p.Record {
		if fperms, ok3 := p.Perms[name + ":" + k]; ok3 {
			if valid3, ok4 := fperms[tool.SELECT.String()]; ok4 && valid3.(bool) { view = append(view, k) }
			if valid2, ok4 := fperms[p.Method.String()]; ok4 { 
				if valid2.(bool) { authorized = true 
				} else { 
					if p.Method == tool.UPDATE || p.Method == tool.CREATE { 
						p.WarningUpdateField = append(p.WarningUpdateField, k)
					} else { delKeyRec = append(delKeyRec, k)  }
				}
			}
		} else { view = append(view, k) }
	}
	if len(delKeyRec) > 0 {
		p.PartialResults = "partial results, only treated : "
		for _, del := range delKeyRec { delete(p.Record, del) }
		for k, _ := range p.Record { p.PartialResults += k + " " }
	}
	cols := ""
	for _, v := range view {
		if columns, ok := p.Params[tool.RootColumnsParam]; ok {
			if strings.Contains(columns, v) { cols += v + "," }
		} else { cols += v + "," }
	}
	if len(cols) > 0 { p.Params[tool.RootColumnsParam] = cols[:len(cols) - 1] }
	return name, authorized
}

func (p *PermissionInfo) Get() (tool.Results, error) { return nil, errors.New("not implemented") }

func (p *PermissionInfo) CreateOrUpdate() (tool.Results, error) {
	if p.Method == tool.UPDATE { return p.Update() 
    } else { return p.Create() }
}

func (p *PermissionInfo) Create() (tool.Results, error) {
	v := Validator[Info]()
	v.data = Info{}
	info, err := v.ValidateStruct(p.Record)
	if err != nil { return nil, errors.New("Not a proper struct to create a column - expect <Info> Scheme " + err.Error()) }
	for role, mainPerms := range MAIN_PERMS {
		if info.Info != "" && (role == ADMINROLE || role == CREATEROLE) { continue }
		params := tool.Params{ tool.RootRowsParam : "all", }
		n := info.Name + ":"
		if info.Info != "" { n += info.Info + ":" }
		n += role
		rec := tool.Record{
			entities.NAMEATTR : conn.Quote(n),
			entities.TABLENAMEATTR : conn.Quote(info.Name) ,
		}
		if info.Info == "" { rec[entities.COLNAMEATTR] = "NULL" 
	    } else { rec[entities.COLNAMEATTR] = conn.Quote(info.Info) }
		for perms, value := range mainPerms { rec[perms]=value }
		p.Row.SpecializedFill(params, rec, tool.CREATE)
		p.Row.Verified=false
		res, err := p.Row.CreateOrUpdate() 
		if err != nil { continue }
		p.Results = append(p.Results, res...)
		if err != nil { return nil, err }
	}
	return p.Results, nil
}

func (p *PermissionInfo) Update() (tool.Results, error) {
	v := Validator[Info]()
	v.data = Info{}
	info, err := v.ValidateStruct(p.Record)
	if err != nil { return nil, errors.New("Not a proper struct to update a column - expect <Info> Scheme " + err.Error()) }
	for _, t := range info.Results { 
		rec := tool.Record{}
		if name, ok := t[entities.NAMEATTR]; ok {
			if info.Info == "" { rec[entities.TABLENAMEATTR] = conn.Quote(name.(string))
			} else { rec[entities.COLNAMEATTR] = conn.Quote(name.(string)) }
			p.Row.SpecializedFill(tool.Params{}, rec, tool.CREATE)
			p.Row.Verified=false
			return p.Row.CreateOrUpdate()
		}
	}
	return nil, errors.New("no permissions to update")
}

func (p *PermissionInfo) Delete() (tool.Results, error) {
	v := Validator[Info]()
	v.data = Info{}
	info, err := v.ValidateStruct(p.Record)
	if err != nil { return nil, errors.New("Not a proper struct to delete a column - expect <Info> Scheme " + err.Error()) }
	params := tool.Params{ entities.TABLENAMEATTR :  info.Name, }
	p.Row.SpecializedFill(params, tool.Record{}, tool.DELETE)
	return p.Row.Delete()
}
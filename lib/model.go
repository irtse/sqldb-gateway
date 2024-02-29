package lib

import (
	"fmt"
	"strconv"
	"sqldb-ws/lib/entities"
)
var ADMINROLE = "admin"
var WRITEROLE = "manager"
var CREATEROLE = "creator"
var UPDATEROLE = "updater"
var READERROLE = "reader"
var PERMS = []string{entities.CREATEPERMS, entities.UPDATEPERMS, entities.DELETEPERMS, entities.READPERMS}

var MAIN_PERMS=map[string]map[string]bool{
	ADMINROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: true, entities.DELETEPERMS: true, },
    WRITEROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: true, entities.DELETEPERMS: false, },
	CREATEROLE: map[string]bool{ entities.CREATEPERMS : true, entities.UPDATEPERMS: false, entities.DELETEPERMS: false, },
	UPDATEROLE: map[string]bool{ entities.CREATEPERMS : false, entities.UPDATEPERMS: true, entities.DELETEPERMS: false, },
	READERROLE: map[string]bool{ entities.CREATEPERMS : false, entities.UPDATEPERMS: false, entities.DELETEPERMS: false, },
}

// API COMMON Models 
type Method int64
const(
	SELECT Method = 1
	CREATE Method = 2
	UPDATE Method = 3
	DELETE Method = 4
)
func (s Method) String() string {
	switch s {
		case SELECT: return "read"
		case CREATE: return "write"
		case UPDATE: return "update"
		case DELETE: return "delete"
	}
	return "unknown"
}

func (s Method) Method() string {
	switch s {
		case SELECT: return "get"
		case CREATE: return "post"
		case UPDATE: return "put"
		case DELETE: return "delete"
	}
	return "unknown"
}

type Results []Record  
type Record map[string]interface{}

func (ar *Record) GetString(column string) string {
	return fmt.Sprintf("%v", (*ar)[column])
}

func (ar *Record) GetInt(column string) int {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.Atoi(str)
	return val
}

func (ar *Record) GetFloat(column string) float64 {
	str := fmt.Sprintf("%v", (*ar)[column])
	val, _ := strconv.ParseFloat(str, 64)
	return val
}

var DATATYPE = []string {
	"TINYINT",
	"SMALLINT",
	"MEDIUMINT",
	"INT",
	"MONEY",
	"INTEGER",
	"BIGINT",
	"FLOAT",
	"DOUBLE",
	"TIME",
	"DATE",
	"BOOLEAN",
	"TIMESTAMP",
	"VARCHAR",
	"TINYTEXT",
	"SMALLTEXT",
	"MEDIUMTEXT",
	"TEXT",
	"TINYBLOB",
	"SMALLBLOB",
	"MEDIUMBLOB",
	"BLOB", // ??? 
	"ENUM",
	"ONETOMANY", // TODO FOR REAL
	"MANYTOMANY", // TODO FOR REAL
}
/*
{
   "name" : "contractual",
   "type" : "boolean",
   "required" : false,
   "read_level": "normal",
   "readonly": false,
   "index": 0,
   "dbschema_id": 309,
   "label" : "contractual",
   "description": "contractual status of the formalized data"
}
*/
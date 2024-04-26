package schema
import "strings"
var NAMEKEY = "name"
var LABELKEY = "label"
var TYPEKEY = "type"

var FOREIGNTABLEKEY = "foreign_table"
var CONSTRAINTKEY = "constraints"
var LINKKEY = "link_id"

var STARTKEY = "start_date"
var ENDKEY = "end_date"

var STATEPENDING = "pending"
var STATEPROGRESSING = "progressing"
var STATEDISMISS= "dismiss"
var STATECOMPLETED = "completed"

var LEVELADMIN = "admin"
var LEVELMODERATOR = "moderator"
var LEVELRESPONSIBLE = "responsible"
var LEVELNORMAL = "normal"
var LEVELOWN = "own"
var READLEVELACCESS = []string{ LEVELOWN, LEVELNORMAL, LEVELRESPONSIBLE, LEVELMODERATOR, LEVELADMIN, }

type DataType int

const (
    SMALLINT DataType = iota + 1
	INTEGER
	BIGINT
	FLOAT8
	DECIMAL
	TIME
	DATE
	TIMESTAMP
	BOOLEAN
	SMALLVARCHAR
	MEDIUMVARCHAR
	VARCHAR
	BIGVARCHAR
	TEXT
	ENUMOPERATOR
	ENUMSEPARATOR
	ENUMLEVEL
	ENUMLEVELCOMPLETE
	ENUMSTATE
	ENUMURGENCY
	ONETOMANY
	MANYTOMANY
)
func DataTypeToEnum() string {
	enum := "enum("
	for _, val := range DataTypeList() { 
		enum += "'" +  enumName(strings.ToLower(val)) + "', " 
	}
	return enum[:len(enum)-2] + ")"
}

func enumName(name string) string {
	if (strings.Contains(name, "enum")) { 
		return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), ",", "_"), "'", ""), "(", "__"), ")", ""), " ", "")
	}
	return name
}

func DataTypeList() []string {
    return []string{"SMALLINT", "INTEGER", "BIGINT", "DOUBLE PRECISION", "DECIMAL", "TIME",
	"DATE", "TIMESTAMP", "BOOLEAN", "VARCHAR(32)", "VARCHAR(64)", "VARCHAR(128)", "VARCHAR(255)",
	"TEXT", "VARCHAR(6)", "ENUM('and', 'or')",
	"ENUM('"+ LEVELADMIN + "', '"+ LEVELMODERATOR + "', '"+ LEVELRESPONSIBLE + "', '"+ LEVELNORMAL + "')",
	"ENUM('"+ LEVELADMIN + "', '"+ LEVELMODERATOR + "', '"+ LEVELRESPONSIBLE + "', '"+ LEVELNORMAL + "', '"+ LEVELOWN + "')",
	"ENUM('" + STATEPENDING + "', '" + STATEPROGRESSING + "', '" + STATEDISMISS + "', '" + STATECOMPLETED + "')",
	"ENUM('low', 'normal', 'high')", "ONETOMANY", "MANYTOMANY"}
}

func (s DataType) String() string {  return strings.ToLower(DataTypeList()[s-1]) }

var CREATEPERMS = "write"
var UPDATEPERMS = "update"
var DELETEPERMS = "delete"
var READPERMS = "read"

var ADMINROLE = "admin"
var WRITEROLE = "manager"
var CREATEROLE = "creator"
var UPDATEROLE = "updater"
var READERROLE = "reader"
var PERMS = []string{CREATEPERMS, UPDATEPERMS, DELETEPERMS, READPERMS}

var MAIN_PERMS=map[string]map[string]bool{
	ADMINROLE: map[string]bool{ CREATEPERMS : true, UPDATEPERMS: true, DELETEPERMS: true, },
    WRITEROLE: map[string]bool{ CREATEPERMS : true, UPDATEPERMS: true, DELETEPERMS: false, },
	CREATEROLE: map[string]bool{ CREATEPERMS : true, UPDATEPERMS: false, DELETEPERMS: false, },
	UPDATEROLE: map[string]bool{ CREATEPERMS : false, UPDATEPERMS: true, DELETEPERMS: false, },
	READERROLE: map[string]bool{ CREATEPERMS : false, UPDATEPERMS: false, DELETEPERMS: false, },
}

package lib

// API COMMON INTERFACE
// Defined infrastructure service functions
type InfraServiceItf interface {
	Verify(string)              					(string, bool)
	Save(restriction... string) 			        (error)
	Count(restriction... string)  					(Results, error)
	Get(restriction... string)  					(Results, error)
	CreateOrUpdate(restriction... string)        	(Results, error)
	Delete(restriction... string)                	(Results, error)
	Template(restriction... string)               	(interface{}, error) 
	GenerateFromTemplate(string) error
}
var EXCEPTION_FUNC = []string{"Count"}
// Defined domain service functions
type DomainITF interface {
	SetParams(params Params)
	SetExternalSuperAdmin(external bool)
	CountNewDataAccess(tableName string, filter string, countParams Params) ([]string, int64)
	PermsSuperCall(params Params, record Record, method Method, funcName string, args... interface{}) (Results, error)
	SuperCall(params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	Call(params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	IsSuperCall() bool
	GetUser() string
	IsShallowed() bool 
	GetEmpty() bool
	SetEmpty(empty bool)
	SetLowerRes(empty bool)
	BuildPath(tableName string, rows string, extra... string) string
	GeneratePathFilter(path string, record Record, params Params) (string, Params)
	IsSuperAdmin() bool
	ViewDefinition(tableName string, innerRestriction... string) (string, string)
	Schema(record Record, p bool) (Results, error)
	GetParams() Params
	ByEntityUser(tableName string, extra ...string) (string)
	PostTreat(results Results, tableName string) Results
	PermsCheck(tableName string, colName string, level string, method Method) bool
} 

type DbITF interface {
	Close()
	GetSQLView()        string 
	GetSQLOrder()       string 
	GetSQLRestriction() string 	
}
// Defined specialized service functions
type SpecializedServiceInfo interface { GetName() string }
type SpecializedService interface {
	Entity() SpecializedServiceInfo
	SetDomain(d DomainITF)
	WriteRowAutomation(record Record, tableName string)
	VerifyRowAutomation(record Record, create bool) (Record, bool, bool)
	DeleteRowAutomation(results Results, tableName string)
	UpdateRowAutomation(results Results, record Record) 
	PostTreatment(results Results, tableName string, dest_id... string) Results
	ConfigureFilter(tableName string) (string, string)
}

type AbstractSpecializedService struct { Domain DomainITF }
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }


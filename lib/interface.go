package lib

// API COMMON INTERFACE
// Defined infrastructure service functions
type InfraServiceItf interface {
	SetAuth(bool)
	Verify(string)              					(string, bool)
	Save(restriction... string) 			        (error)
	Get(restriction... string)  					(Results, error)
	CreateOrUpdate(restriction... string)        	(Results, error)
	Delete(restriction... string)                	(Results, error)
	Import(string, ...string)          	(Results, error)
	Template(restriction... string)               	(interface{}, error) 
	GenerateFromTemplate(string) error
}
// Defined domain service functions
type DomainITF interface {
	PermsSuperCall(params Params, record Record, method Method, funcName string, args... interface{}) (Results, error)
	SuperCall(params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	Call(params Params, rec Record, m Method, auth bool, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	IsSuperCall() bool
	GetUser() string
	IsShallowed() bool 
	IsRawView() bool
	BuildPath(tableName string, rows string, extra... string) string
	GeneratePathFilter(path string, record Record, params Params) (string, Params)
	IsSuperAdmin() bool
	GetPermission() InfraServiceItf
	DeleteRow(tableName string, results Results)
	WriteRow(tableName string, record Record)
	ViewDefinition(tableName string, innerRestriction... string) (string, string)
	Schema(record Record, p bool) (Results, error)
	GetParams() Params
	PostTreat(results Results, tableName string, shallow bool,  additonnalRestriction ...string) Results
} 

type DbITF interface {
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

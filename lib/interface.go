package lib

// API COMMON INTERFACE
// Defined infrastructure service functions
type InfraServiceItf interface {
	SetAuth(bool)
	Verify(string)              (string, bool)
	Save() 			        	(error)
	Get()                   	(Results, error)
	CreateOrUpdate()        	(Results, error)
	Delete()                	(Results, error)
	Link()        				(Results, error)
	UnLink()                	(Results, error)
	Import(string)          	(Results, error)
	Template()               	(interface{}, error) 
	GenerateFromTemplate(string) error
}
// Defined domain service functions
type DomainITF interface {
	SuperCall(params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	Call(params Params, rec Record, m Method, auth bool, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	GetUser() string
	IsShallowed() bool 
	IsRawView() bool
	BuildPath(tableName string, rows string, extra... string) string
	GeneratePathFilter(path string, record Record, params Params) (string, Params)
	IsSuperAdmin() bool
	GetPermission() InfraServiceItf
	DeleteRow(tableName string, results Results)
	WriteRow(tableName string, record Record)
	ViewDefinition(tableName string, params Params) (string, string)
	Schema(record Record) (Results, error)
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
	VerifyRowAutomation(record Record, create bool) (Record, bool)
	DeleteRowAutomation(results Results, tableName string)
	UpdateRowAutomation(results Results, record Record) 
	PostTreatment(results Results, tableName string) Results
	ConfigureFilter(tableName string, params Params) (string, string)
}

type AbstractSpecializedService struct { Domain DomainITF }
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }

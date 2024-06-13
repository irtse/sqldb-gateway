package utils

import ( 
	conn "sqldb-ws/lib/infrastructure/connector"
	infrastructure "sqldb-ws/lib/infrastructure/service" 
)

type SpecializedServiceITF interface {
	SetDomain(d DomainITF)
	Entity() SpecializedServiceInfo
	PostTreatment(results Results, tableName string, dest_id... string) Results
	infrastructure.InfraSpecializedServiceItf
}
type AbstractSpecializedService struct { Domain DomainITF }
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }

type SpecializedService struct { AbstractSpecializedService }
func (s *SpecializedService) PostTreatment(results Results, tableName string, dest_id... string) Results { 
	return s.Domain.PostTreat(results, tableName, true) }
func (s *SpecializedService) ConfigureFilter(tableName string, innerestr... string) (string, string, string, string) { 
	return s.Domain.ViewDefinition(tableName, innerestr...) 
}

type SpecializedServiceInfo interface { GetName() string }
type DomainITF interface {
	GetMethod() Method
	SetOwn(own bool)
	SetExternalSuperAdmin(external bool)
	CountNewDataAccess(tableName string, filter string, countParams Params) ([]string, int64)
	SpecialSuperCall(params Params, record Record, method Method, args... interface{}) (Results, error)
	PermsSuperCall(params Params, record Record, method Method, args... interface{}) (Results, error)
	SuperCall(params Params, rec Record, m Method, args... interface{}) (Results, error)
	Call(params Params, rec Record, m Method, args... interface{}) (Results, error)
	GetDb() *conn.Db 
    SetIsCustom(isCustom bool)
	IsSuperCall() bool
	GetUser() string
	IsShallowed() bool 
	GetEmpty() bool
	SetEmpty(empty bool)
	SetLowerRes(empty bool)
	BuildPath(tableName string, rows string, extra... string) string
	GetFilter(filterID string, viewfilterID string, schemaID string) (string, string, string, string, string)
	IsSuperAdmin() bool
	GetAutoload() bool
	ValidateBySchema(data Record, tableName string) (Record, error)
	ViewDefinition(tableName string, innerRestriction... string) (string, string, string, string)
	GetParams() Params
	PostTreat(results Results, tableName string, isWorflow bool) Results
	PermsCheck(tableName string, colName string, level string, method Method) bool
} 

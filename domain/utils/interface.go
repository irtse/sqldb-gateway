package utils

import (
	conn "sqldb-ws/infrastructure/connector"
	infrastructure "sqldb-ws/infrastructure/service"
)

type SpecializedServiceITF interface {
	SetDomain(d DomainITF)
	Entity() SpecializedServiceInfo
	TransformToGenericView(results Results, tableName string, dest_id ...string) Results
	infrastructure.InfraSpecializedServiceItf
}
type SpecializedServiceInfo interface{ GetName() string }

type DomainITF interface {
	// Main Procedure of services at Domain level.
	SuperCall(params Params, record Record, method Method, isOwn bool, args ...interface{}) (Results, error)
	CreateSuperCall(params Params, rec Record, args ...interface{}) (Results, error)
	UpdateSuperCall(params Params, rec Record, args ...interface{}) (Results, error)
	DeleteSuperCall(params Params, args ...interface{}) (Results, error)
	Call(params Params, rec Record, m Method, args ...interface{}) (Results, error)

	// Main accessor defined by DomainITF interface
	GetAutoload() bool
	GetDb() *conn.Database
	GetMethod() Method
	GetTable() string
	GetUser() string
	GetEmpty() bool
	GetParams() Params

	// Main accessor defined by DomainITF interface
	SetExternalSuperAdmin(external bool)
	HandleRecordAttributes(record Record)

	// Main accessor defined by DomainITF interface
	IsOwn(checkPerm bool, force bool, method Method) bool
	IsSuperCall() bool
	IsSuperAdmin() bool
	IsShallowed() bool
	IsLowerResult() bool

	VerifyAuth(tableName string, colName string, level string, method Method, dest ...string) bool
}

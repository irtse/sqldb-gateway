package lib

import ()

type DomainITF interface {
	SafeCall(admin bool, user string, params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
	UnSafeCall(user string, params Params, rec Record, m Method, funcName string, args... interface{}) (Results, error)
    SetIsCustom(isCustom bool)
	GetUser() string
}
type SpecializedServiceInfo interface { GetName() string }
type SpecializedService interface {
	Entity() SpecializedServiceInfo
	SetDomain(d DomainITF)
	WriteRowAutomation(record Record)
	VerifyRowAutomation(record Record, create bool) (Record, bool)
	DeleteRowAutomation(results Results)
	UpdateRowAutomation(results Results, record Record) 
}

type AbstractSpecializedService struct { 
	Domain           DomainITF 
}
func (s *AbstractSpecializedService) SetDomain(d DomainITF) {  s.Domain = d  }

type CustomService struct { AbstractSpecializedService }
func (s *CustomService) UpdateRowAutomation(results Results, record Record) {}
func (s *CustomService) WriteRowAutomation(record Record) {}
func (s *CustomService) DeleteRowAutomation(results Results) { }
func (s *CustomService) Entity() SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyRowAutomation(record Record, create bool) (Record, bool) { return record, true }
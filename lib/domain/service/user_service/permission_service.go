package user_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type PermissionService struct { tool.AbstractSpecializedService }

func (s *PermissionService) Entity() tool.SpecializedServiceInfo { return entities.DBPermission }
func (s *PermissionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *PermissionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *PermissionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *PermissionService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *PermissionService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName) 
}
func (s *PermissionService) ConfigureFilter(tableName string) (string, string) {
	rows, ok := s.Domain.GetParams()[tool.RootRowsParam]
	ids, ok2 := s.Domain.GetParams()[tool.SpecialIDParam]
	if (ok && fmt.Sprintf("%v", rows) != tool.ReservedParam) || (ok2 && ids != "") {
		return s.Domain.ViewDefinition(tableName)
	}
	restr := "id IN (SELECT " + entities.DBPermission.Name + "_id FROM " 
	restr += entities.DBRolePermission.Name + " WHERE " + entities.DBRole.Name + "_id IN ("
	restr += "SELECT " + entities.DBRole.Name + "_id FROM " 
	restr += entities.DBRoleAttribution.Name + " WHERE " + entities.DBUser.Name + "_id IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE " 
	restr += "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ") OR " + entities.DBEntity.Name + "_id IN ("
	restr += "SELECT " + entities.DBEntity.Name + "_id FROM "
	restr += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
	restr += "name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + "))))"
	return s.Domain.ViewDefinition(tableName, restr)
}
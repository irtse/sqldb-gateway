package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type RoleService struct { tool.AbstractSpecializedService }

func (s *RoleService) Entity() tool.SpecializedServiceInfo { return entities.DBRole }
func (s *RoleService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *RoleService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *RoleService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *RoleService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *RoleService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *RoleService) ConfigureFilter(tableName string) (string, string) {
	restr := "id IN (SELECT "+ entities.RootID(entities.DBRole.Name) + " FROM " +  entities.DBRoleAttribution.Name + " " 
	restr += "WHERE " + entities.RootID(entities.DBEntity.Name) + " IN ("
	restr += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")) "
	restr += "OR " + entities.RootID(entities.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE name=" + conn.Quote(s.Domain.GetUser()) + " OR email=" + conn.Quote(s.Domain.GetUser()) + ")) "
	return s.Domain.ViewDefinition(tableName, restr)
}
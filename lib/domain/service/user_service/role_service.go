package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type RoleService struct { tool.AbstractSpecializedService }

func (s *RoleService) Entity() tool.SpecializedServiceInfo { return entities.DBRole }
func (s *RoleService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *RoleService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *RoleService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *RoleService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *RoleService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *RoleService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT "+ entities.RootID(entities.DBRole.Name) + " FROM " +  entities.DBRoleAttribution.Name + " " 
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')) "
	params[tool.RootSQLFilterParam] += "OR " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')) "
	return s.Domain.ViewDefinition(tableName, params)
}
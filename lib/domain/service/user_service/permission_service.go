package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type PermissionService struct { tool.AbstractSpecializedService }

func (s *PermissionService) Entity() tool.SpecializedServiceInfo { return entities.DBPermission }
func (s *PermissionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *PermissionService) DeleteRowAutomation(results tool.Results) { }
func (s *PermissionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *PermissionService) WriteRowAutomation(record tool.Record) { }
func (s *PermissionService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName) 
}
func (s *PermissionService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT " + entities.DBPermission.Name + "_id FROM " 
	params[tool.RootSQLFilterParam] += entities.DBRolePermission.Name + " WHERE " + entities.DBRole.Name + "_id IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.DBRole.Name + "_id FROM " 
	params[tool.RootSQLFilterParam] += entities.DBRoleAttribution.Name + " WHERE " + entities.DBUser.Name + "_id IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE " 
	params[tool.RootSQLFilterParam] += entities.DBUser.Name + ".login = '" + s.Domain.GetUser() + "') OR " + entities.DBEntity.Name + "_id IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.DBEntity.Name + "_id FROM "
	params[tool.RootSQLFilterParam] += entities.DBEntityUser.Name + " WHERE " + entities.DBUser.Name +"_id IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE "
	params[tool.RootSQLFilterParam] += entities.DBUser.Name + ".login = '" + s.Domain.GetUser()  + "'))))"
	return tool.ViewDefinition(s.Domain, tableName, params)
}
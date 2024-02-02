package user_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type PermissionService struct { tool.AbstractSpecializedService }

func (s *PermissionService) Entity() tool.SpecializedServiceInfo { return entities.DBPermission }
func (s *PermissionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	return record, true, false }
func (s *PermissionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *PermissionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *PermissionService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *PermissionService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
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
	return s.Domain.ViewDefinition(tableName, params)
}
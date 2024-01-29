package user_service

import (
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type RoleAttributionService struct { tool.AbstractSpecializedService }

func (s *RoleAttributionService) Entity() tool.SpecializedServiceInfo { return entities.DBRoleAttribution }
func (s *RoleAttributionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	params := tool.Params{ tool.RootTableParam : entities.DBRoleAttribution.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, }
	currentTime := time.Now()
	params[tool.RootSQLFilterParam]= "'" + currentTime.Format("2000-01-01") + "' < start_date OR " 
	params[tool.RootSQLFilterParam]+= "'" + currentTime.Format("2000-01-01") + "' > end_date"
	s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", )	
	return record, true 
}
func (s *RoleAttributionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *RoleAttributionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *RoleAttributionService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *RoleAttributionService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *RoleAttributionService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	params[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	params[tool.RootSQLFilterParam] += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')"
	return tool.ViewDefinition(s.Domain, tableName, params)
}
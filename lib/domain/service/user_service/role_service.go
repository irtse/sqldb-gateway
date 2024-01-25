package user_service

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type RoleService struct { tool.AbstractSpecializedService }

func (s *RoleService) Entity() tool.SpecializedServiceInfo { return entities.DBRole }
func (s *RoleService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *RoleService) DeleteRowAutomation(results tool.Results) { }
func (s *RoleService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *RoleService) WriteRowAutomation(record tool.Record) { }
func (s *RoleService) PostTreatment(results tool.Results) tool.Results {
	res := tool.Results{} // TODO COMPLEXIFY WITH ENTITY ID
	params := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
		tool.RootRowsParam : tool.ReservedParam,
		"login" : s.Domain.GetUser() }
	users, err := s.Domain.SuperCall( params, tool.Record{}, tool.SELECT, "Get", )	
	if err != nil || len(users) == 0 { return res }
	for _, record := range results {
		params = tool.Params{ tool.RootTableParam : entities.DBEntity.Name, 
			                  tool.RootRowsParam : tool.ReservedParam, }
	    ents, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, false, "Get", )	
		ids := ""
		for _, entity := range ents { ids += fmt.Sprintf("%v", entity[tool.SpecialIDParam]) + "," }
		params = tool.Params{ tool.RootTableParam : entities.DBRoleAttribution.Name, 
				                  tool.RootRowsParam : tool.ReservedParam,
				                  entities.RootID(entities.DBRole.Name): fmt.Sprintf("%v", record[tool.SpecialIDParam]), }
		params[tool.RootSQLFilterParam] = "(" + entities.RootID(entities.DBUser.Name) + "="+ fmt.Sprintf("%v", users[0][tool.SpecialIDParam]) +" "
		params[tool.RootSQLFilterParam] += "OR "+ entities.RootID(entities.DBEntity.Name) + " IN (" + ids[:len(ids) - 1] + ") )"
		affectedRole, err := s.Domain.Call( params, tool.Record{}, tool.SELECT, false, "Get", )	
		if err == nil && len(affectedRole) > 0 { res = append(res, record) }
	}
	return res
}

func (s *RoleService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}
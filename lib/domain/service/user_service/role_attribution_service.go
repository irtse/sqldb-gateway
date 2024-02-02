package user_service

import (
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type RoleAttributionService struct { tool.AbstractSpecializedService }

func (s *RoleAttributionService) Entity() tool.SpecializedServiceInfo { return entities.DBRoleAttribution }
func (s *RoleAttributionService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	params := tool.Params{ tool.RootTableParam : entities.DBRoleAttribution.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, }
	currentTime := time.Now()
	sqlFilter := conn.Quote(currentTime.Format("2000-01-01")) + " < start_date OR " 
	sqlFilter += conn.Quote(currentTime.Format("2000-01-01")) + " > end_date"
	s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", sqlFilter)	
	return record, true, false 
}
func (s *RoleAttributionService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *RoleAttributionService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *RoleAttributionService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *RoleAttributionService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *RoleAttributionService) ConfigureFilter(tableName string) (string, string) {
	restr := entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")" 
	restr += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	restr += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name + " "
	restr += "WHERE " + entities.RootID(entities.DBUser.Name) + " IN ("
	restr += "SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.GetUser()) + ")"
	return s.Domain.ViewDefinition(tableName, restr)
}
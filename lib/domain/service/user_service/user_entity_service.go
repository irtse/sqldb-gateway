package user_service

import (
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
)

type UserEntityService struct { tool.AbstractSpecializedService }

func (s *UserEntityService) Entity() tool.SpecializedServiceInfo { return entities.DBEntityUser }
func (s *UserEntityService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	params := tool.Params{ tool.RootTableParam : entities.DBEntityUser.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, }
	currentTime := time.Now()
	params[tool.RootSQLFilterParam]= "'" + currentTime.Format("2000-01-01") + "' < start_date OR " 
	params[tool.RootSQLFilterParam]+= "'" + currentTime.Format("2000-01-01") + "' > end_date"
	s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", )	
	return record, true , false
}
func (s *UserEntityService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *UserEntityService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *UserEntityService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *UserEntityService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *UserEntityService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = entities.RootID(entities.DBUser.Name) + " IN (SELECT id FROM " + entities.DBUser.Name + " WHERE login='" + s.Domain.GetUser() + "')" 
	return s.Domain.ViewDefinition(tableName, params)
}
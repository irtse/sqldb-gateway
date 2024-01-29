package user_service

import (
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type HierarchyService struct { tool.AbstractSpecializedService }

func (s *HierarchyService) Entity() tool.SpecializedServiceInfo { return entities.DBHierarchy }
func (s *HierarchyService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	params := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, }
	currentTime := time.Now()
	params[tool.RootSQLFilterParam]= "'" + currentTime.Format("2000-01-01") + "' < start_date OR " 
	params[tool.RootSQLFilterParam]+= "'" + currentTime.Format("2000-01-01") + "' > end_date"
	s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", )	
	return record, true 
}
func (s *HierarchyService) DeleteRowAutomation(results tool.Results) { }
func (s *HierarchyService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *HierarchyService) WriteRowAutomation(record tool.Record) { }
func (s *HierarchyService) PostTreatment(results tool.Results, tableName string) tool.Results { 	
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *HierarchyService) ConfigureFilter(tableName string, params tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
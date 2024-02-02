package user_service

import (
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type HierarchyService struct { tool.AbstractSpecializedService }

func (s *HierarchyService) Entity() tool.SpecializedServiceInfo { return entities.DBHierarchy }
func (s *HierarchyService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool, bool) { 
	params := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, 
	                       tool.RootRowsParam : tool.ReservedParam, }
	currentTime := time.Now()
	sqlFilter := conn.Quote(currentTime.Format("2000-01-01")) + " < start_date OR " 
	sqlFilter += conn.Quote(currentTime.Format("2000-01-01")) + " > end_date"
	s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", sqlFilter)	
	return record, true , false
}
func (s *HierarchyService) DeleteRowAutomation(results tool.Results, tableName string) { }
func (s *HierarchyService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *HierarchyService) WriteRowAutomation(record tool.Record, tableName string) { }
func (s *HierarchyService) PostTreatment(results tool.Results, tableName string, dest_id... string) tool.Results { 	
	return s.Domain.PostTreat( results, tableName, false) 
}
func (s *HierarchyService) ConfigureFilter(tableName string) (string, string) {
	return s.Domain.ViewDefinition(tableName)
}	
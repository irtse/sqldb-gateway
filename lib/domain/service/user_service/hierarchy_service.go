package user_service

import (
	"fmt"
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type HierarchyService struct { tool.AbstractSpecializedService }

func (s *HierarchyService) Entity() tool.SpecializedServiceInfo { return entities.DBHierarchy }
func (s *HierarchyService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *HierarchyService) DeleteRowAutomation(results tool.Results) { }
func (s *HierarchyService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *HierarchyService) WriteRowAutomation(record tool.Record) { }
func (s *HierarchyService) Filter(results tool.Results) tool.Results {
	res := tool.Results{}
	for _, record := range results {
		found := true
		if date, ok := record["start_date"]; ok && date != nil && date != "" {
			today := time.Now() 
			start, err := time.Parse("2000-01-01", date.(string))
			if err == nil && start.Before(today) { found=false }
			if date2, ok := record["end_date"]; ok && date != nil && date != "" && found {
				end, err := time.Parse("2000-01-01", date2.(string))
				if err == nil && today.Before(end) { found=false}
			}
		}
		if !found {
			params := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, tool.RootRowsParam : tool.ReservedParam, }
			if entityId, ok2 := record[entities.RootID(entities.DBEntity.Name)]; ok2 {
				params[entities.RootID(entities.DBEntity.Name)]= fmt.Sprintf("%d", entityId.(int64))
			}
			if userId, ok3 := record[entities.RootID(entities.DBUser.Name)]; ok3 {
				params[entities.RootID(entities.DBUser.Name)]= fmt.Sprintf("%d", userId.(int64))
			}
			if entityParentId, ok4 := record["parent_" + entities.RootID(entities.DBEntity.Name)]; ok4 {
				params["parent_" + entities.RootID(entities.DBEntity.Name)]= fmt.Sprintf("%d", entityParentId.(int64))
			}
			s.Domain.SafeCall(true, "", params, tool.Record{}, tool.DELETE, "Delete", )	
		} else { res = append(res, record) }
	}
	return res
}
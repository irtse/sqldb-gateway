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
func (s *HierarchyService) PostTreatment(results tool.Results) tool.Results {
	res := tool.Results{}
	for _, record := range results {
		found := true
		if start, ok := record["start_date"]; ok && start != nil && start != "" {
			if !start.(time.Time).Before(time.Now()) { found=false }
			if end, ok := record["end_date"]; ok && end != nil && end != "" && found {
				if !time.Now().Before(end.(time.Time)) { found=false }
			}
		}
		if !found {
			params := tool.Params{ tool.RootTableParam : entities.DBHierarchy.Name, tool.RootRowsParam : tool.ReservedParam, }
			if entityId, ok2 := record[entities.RootID(entities.DBEntity.Name)]; ok2 {
				params[entities.RootID(entities.DBEntity.Name)]= fmt.Sprintf("%v", entityId)
			}
			if userId, ok3 := record[entities.RootID(entities.DBUser.Name)]; ok3 {
				params[entities.RootID(entities.DBUser.Name)]= fmt.Sprintf("%v", userId)
			}
			if entityParentId, ok4 := record["parent_" + entities.RootID(entities.DBEntity.Name)]; ok4 {
				params["parent_" + entities.RootID(entities.DBEntity.Name)]= fmt.Sprintf("%v", entityParentId)
			}
			s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", )	
		} else { res = append(res, record) }
	}
	return res
}

func (s *HierarchyService) ConfigureFilter(tableName string, params tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
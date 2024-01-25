package user_service

import (
	"fmt"
	"time"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type UserEntityService struct { tool.AbstractSpecializedService }

func (s *UserEntityService) Entity() tool.SpecializedServiceInfo { return entities.DBEntityUser }
func (s *UserEntityService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *UserEntityService) DeleteRowAutomation(results tool.Results) { }
func (s *UserEntityService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *UserEntityService) WriteRowAutomation(record tool.Record) { }
func (s *UserEntityService) PostTreatment(results tool.Results) tool.Results {
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
			params := tool.Params{ tool.RootTableParam : entities.DBEntityUser.Name, 
				                   tool.RootRowsParam : tool.ReservedParam, }
			if entityId, ok2 := record[entities.RootID(entities.DBEntity.Name)]; ok2 {
				params[entities.RootID(entities.DBEntity.Name)]= fmt.Sprintf("%v", entityId)
			}
			if userId, ok3 := record[entities.RootID(entities.DBUser.Name)]; ok3 {
				params[entities.RootID(entities.DBUser.Name)]= fmt.Sprintf("%v", userId)
			}
			s.Domain.SuperCall( params, tool.Record{}, tool.DELETE, "Delete", )	
		} else { res = append(res, record) }
	}
	return res
}

func (s *UserEntityService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}
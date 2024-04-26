package task_service

import (
	utils "sqldb-ws/lib/domain/utils"
	schserv "sqldb-ws/lib/domain/schema"
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
type WorkflowService struct { 
	utils.AbstractSpecializedService 
	infrastructure.InfraSpecializedService
}

func (s *WorkflowService) Entity() utils.SpecializedServiceInfo { return schserv.DBWorkflow }
func (s *WorkflowService) PostTreatment(results utils.Results, tableName string, dest_id... string) utils.Results { 
	res := utils.Results{}	
	for _, rec := range results { // filter by allowed schemas
		schema, err := schserv.GetSchemaByID(int64(rec[schserv.RootID(schserv.DBSchema.Name)].(float64)))
		if err == nil && s.Domain.PermsCheck(schema.Name, "", "", utils.CREATE) { res = append(res, rec) }
	}
	return s.Domain.PostTreat(res, tableName, true) 
}
func (s *WorkflowService) ConfigureFilter(tableName string) (string, string, string, string) { return s.Domain.ViewDefinition(tableName) }
func (s *WorkflowService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, bool, bool) { 
	if s.Domain.GetMethod() != utils.DELETE {
		rec, err := s.Domain.ValidateBySchema(record, tablename)
		if err != nil && !s.Domain.GetAutoload() { return rec, false, false } else { rec = record }
		return rec, true, false 
	}
	return record, true, true
}
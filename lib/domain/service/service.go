package service

import ( 
	"sqldb-ws/lib/domain/utils" 
	task "sqldb-ws/lib/domain/service/task_service" 
	schema "sqldb-ws/lib/domain/service/schema_service" 
	favorite "sqldb-ws/lib/domain/service/favorite_service" 
	infrastructure "sqldb-ws/lib/infrastructure/service"
)
// export all specialized services available per domain
var SERVICES = []utils.SpecializedServiceITF{
	&schema.SchemaService{}, 
	&schema.SchemaFields{}, 
	&schema.ViewService{}, 
	&task.RequestService{}, 
	&task.TaskService{},
	&task.WorkflowService{},
	&favorite.FilterService{},
}
// funct to get specialized service depending on table reached
func SpecializedService(name string) utils.SpecializedServiceITF {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &CustomService{}
}
// Default Specialized Service. 
type CustomService struct { 
	utils.SpecializedService
	infrastructure.InfraSpecializedService
}
func (s *CustomService) Entity() utils.SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyRowAutomation(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) { 
	if s.Domain.GetMethod() != utils.DELETE {
		rec, err := s.Domain.ValidateBySchema(record, tablename)
		if err != nil && !s.Domain.GetAutoload() { return rec, err, false } else { rec = record }
		return rec, nil, false 
	}
	return record, nil, true
}
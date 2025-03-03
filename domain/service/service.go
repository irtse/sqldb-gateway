package service

import (
	favorite "sqldb-ws/domain/service/favorite_service"
	schema "sqldb-ws/domain/service/schema_service"
	task "sqldb-ws/domain/service/task_service"
	servutils "sqldb-ws/domain/service/utils"
	"sqldb-ws/domain/utils"
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
		if service.Entity().GetName() == name {
			return service
		}
	}
	return &CustomService{}
}

// Default Specialized Service.
type CustomService struct {
	servutils.SpecializedService
}

func (s *CustomService) Entity() utils.SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	return servutils.CheckAutoLoad(tablename, record, s.Domain)
}
func (s *CustomService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	// EMPTY AND PROUD TO BE
}
func (s *CustomService) SpecializedUpdateRow(results []map[string]interface{}, record map[string]interface{}) {
	// EMPTY AND PROUD TO BE
}
func (s *CustomService) SpecializedCreateRow(record map[string]interface{}, tableName string) {
	// EMPTY AND PROUD TO BE
}

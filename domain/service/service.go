package service

import (
	favorite "sqldb-ws/domain/service/favorite_service"
	schema "sqldb-ws/domain/service/schema_service"
	task "sqldb-ws/domain/service/task_service"
	user "sqldb-ws/domain/service/user_service"
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
	&favorite.DashboardService{},
	&user.DelegationService{},
	&user.ShareService{},
	&user.UserService{},
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
	if _, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.SpecializedService.VerifyDataIntegrity(record, tablename)
	} else {
		return record, err, false
	}

}

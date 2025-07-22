package service

import (
	"sqldb-ws/domain/specialized_service/email_service"
	favorite "sqldb-ws/domain/specialized_service/favorite_service"
	schema "sqldb-ws/domain/specialized_service/schema_service"
	task "sqldb-ws/domain/specialized_service/task_service"
	user "sqldb-ws/domain/specialized_service/user_service"
	servutils "sqldb-ws/domain/specialized_service/utils"
	"sqldb-ws/domain/utils"
)

// export all specialized services available per domain
var SERVICES = []func() utils.SpecializedServiceITF{
	schema.NewSchemaService,
	schema.NewSchemaFieldsService,
	schema.NewViewService,
	task.NewWorkflowService,
	task.NewTaskService,
	task.NewRequestService,
	favorite.NewFilterService,
	//&favorite.DashboardService{},
	user.NewDelegationService,
	user.NewShareService,
	user.NewUserService,
	email_service.NewEmailResponseService,
	email_service.NewEmailSendedService,
	email_service.NewEmailSendedUserService,
}

// funct to get specialized service depending on table reached
func SpecializedService(name string) utils.SpecializedServiceITF {
	for _, service := range SERVICES {
		if service().Entity().GetName() == name {
			return service()
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

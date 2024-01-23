package service

import ( 
	tool "sqldb-ws/lib" 
	task "sqldb-ws/lib/domain/service/task_service" 
	user "sqldb-ws/lib/domain/service/user_service"
	schema "sqldb-ws/lib/domain/service/schema_service" 
)

var SERVICES = []tool.SpecializedService{
	&schema.SchemaService{}, 
	&schema.SchemaFields{}, 
	&task.TaskAssigneeService{}, 
	&task.TaskVerifyerService{}, 
	&task.TaskService{},
	&user.UserEntityService{},
	&user.HierarchyService{},
	&user.RoleAttributionService{},
}

func SpecializedService(name string) tool.SpecializedService {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &tool.CustomService{}
}

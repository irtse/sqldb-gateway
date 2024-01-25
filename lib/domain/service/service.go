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
	&schema.ViewService{},
	&schema.ActionService{},
	&task.TaskAssigneeService{}, 
	&task.TaskVerifyerService{}, 
	&task.TaskService{},
	&user.UserEntityService{},
	&user.HierarchyService{},
	&user.RoleAttributionService{},
	&user.RoleService{},
	&user.EntityService{},
}

func SpecializedService(name string) tool.SpecializedService {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &CustomService{}
}

type CustomService struct { tool.AbstractSpecializedService }
func (s *CustomService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *CustomService) WriteRowAutomation(record tool.Record) {}
func (s *CustomService) DeleteRowAutomation(results tool.Results) { }
func (s *CustomService) Entity() tool.SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *CustomService) PostTreatment(results tool.Results) tool.Results { 	return results }
func (s *CustomService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
// to set up ConfigureFilter
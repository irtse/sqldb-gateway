package service

import ( 
	"fmt"
	tool "sqldb-ws/lib" 
	"sqldb-ws/lib/infrastructure/entities" 
	task "sqldb-ws/lib/domain/service/task_service" 
	user "sqldb-ws/lib/domain/service/user_service"
	schema "sqldb-ws/lib/domain/service/schema_service" 
)

var SERVICES = []tool.SpecializedService{
	&schema.SchemaService{}, 
	&schema.SchemaFields{}, 
	&schema.ViewService{},
	&schema.ActionService{},
	&schema.ViewActionService{},
	&task.TaskAssigneeService{}, 
	&task.TaskVerifyerService{}, 
	&task.TaskService{},
	&task.TaskWatcherService{},
	&user.UserEntityService{},
	&user.HierarchyService{},
	&user.RoleAttributionService{},
	&user.RoleService{},
	&user.EntityService{},
	&user.PermissionService{},
	&user.UserService{},
	&user.UserEntryService{},
}

func SpecializedService(name string) tool.SpecializedService {
	for _, service := range SERVICES {
		if service.Entity().GetName() == name { return service }
	}
	return &CustomService{}
}

type CustomService struct { tool.AbstractSpecializedService }
func (s *CustomService) UpdateRowAutomation(results tool.Results, record tool.Record) {

}
func (s *CustomService) WriteRowAutomation(record tool.Record, tableName string) {
	tool.WriteRow(s.Domain, tableName, record)
}
func (s *CustomService) DeleteRowAutomation(results tool.Results, tableName string) {
	tool.DeleteRow(s.Domain, tableName, results)
}
func (s *CustomService) Entity() tool.SpecializedServiceInfo { return nil }
func (s *CustomService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { return record, true }
func (s *CustomService) PostTreatment(results tool.Results, tableName string) tool.Results { 
	return tool.PostTreat(s.Domain, results, tableName, false) 
}
func (s *CustomService) ConfigureFilter(tableName string, params  tool.Params) (string, string) {
	params[tool.RootSQLFilterParam] = "id IN (SELECT " + fmt.Sprintf("%v",  entities.RootID("dest_table")) + " FROM " + entities.DBUserEntry.Name 
	params[tool.RootSQLFilterParam] += " WHERE " + fmt.Sprintf("%v", entities.RootID(entities.DBSchema.Name))  + " IN ("
	params[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBSchema.Name + " WHERE name=" + tableName + ") "
	params[tool.RootSQLFilterParam] += "AND " + fmt.Sprintf("%v",  entities.RootID(entities.DBUser.Name)) + " IN (SELECT id FROM " + entities.DBUser.Name 
	params[tool.RootSQLFilterParam] += " WHERE login='" + s.Domain.GetUser() + "'))"
	return tool.ViewDefinition(s.Domain, tableName, params)
}	
// to set up ConfigureFilter
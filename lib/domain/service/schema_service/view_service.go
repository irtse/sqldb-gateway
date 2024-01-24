package schema_service

import (
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type HierarchyService struct { tool.AbstractSpecializedService }

func (s *HierarchyService) Entity() tool.SpecializedServiceInfo { return entities.DBView }
func (s *HierarchyService) VerifyRowAutomation(record tool.Record, create bool) (tool.Record, bool) { 
	if _, ok := record[entities.RootID(entities.DBSchema.Name)]; !ok { return record, false }
	params := tool.Params{ tool.RootTableParam :  record[entities.RootID(entities.DBSchema.Name)].(string), }
	schemas, err := s.Domain.SafeCall(true, "",
									params, 
									tool.Record{}, 
									tool.SELECT, 
									"Get")
	if err != nil && len(schemas) == 0 { return record, false }
	for _, scheme := range schemas {
		if _, valid := scheme[tool.RootColumnsParam]; !valid { return record, false }
		_, validUserId := scheme[tool.RootColumnsParam].(tool.Record)[entities.RootID(entities.DBUser.Name)]
		_, validEntityId := scheme[tool.RootColumnsParam].(tool.Record)[entities.RootID(entities.DBEntity.Name)]
		if !validUserId && !validEntityId { return record, false } 
	}
	return record, true 
}
func (s *HierarchyService) DeleteRowAutomation(results tool.Results) { }
func (s *HierarchyService) UpdateRowAutomation(results tool.Results, record tool.Record) {}
func (s *HierarchyService) WriteRowAutomation(record tool.Record) { }
func (s *HierarchyService) PostTreatment(results tool.Results) tool.Results { return results }

func (s *HierarchyService) ConfigureFilter(tableName string, params tool.Params) (string, string) { return "", "" }	
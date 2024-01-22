package domain

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
	conn "sqldb-ws/lib/infrastructure/connector"
)

type TaskAttributionService struct { Domain tool.DomainITF }

func (s *TaskAttributionService) SetDomain(d tool.DomainITF) { s.Domain = d }
func (s *TaskAttributionService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskAttribution }
func (s *TaskAttributionService) VerifyRowWorkflow(record tool.Record, create bool) (tool.Record, bool) { 
	var res tool.Results
	if taskID, ok := record[entities.RootID(entities.DBTask.Name)]; ok && taskID != nil {
		if userID, ok := record[entities.RootID(entities.DBUser.Name)]; ok && userID != nil {
			res, _ = s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : entities.DBTaskAttribution.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%d", record[entities.RootID(entities.DBTask.Name)].(int64)),
						 entities.RootID(entities.DBUser.Name): fmt.Sprintf("%d", record[entities.RootID(entities.DBUser.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		} else if entityID, ok := record[entities.RootID(entities.DBEntity.Name)]; ok && entityID != nil {
			res, _ = s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : entities.DBTaskAttribution.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBEntity.Name): fmt.Sprintf("%d", record[entities.RootID(entities.DBEntity.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		}
	}
	return record, res == nil || len(res) == 0
}
func (s *TaskAttributionService) DeleteRowWorkflow(results tool.Results) { }
func (s *TaskAttributionService) UpdateRowWorkflow(results tool.Results, record tool.Record) {
	for _, res := range results {
		s.WriteRowWorkflow(res)
	}
}
func (s *TaskAttributionService) WriteRowWorkflow(record tool.Record) { 
	if isAssignee, ok := record["assignee"]; !ok || isAssignee == false { return }
	paramsNew := tool.Params{ tool.RootTableParam : entities.DBUser.Name, 
							  tool.RootRowsParam: tool.ReservedParam, }
	paramsNew[tool.RootSQLFilterParam] += "id IN (SELECT id FROM " + entities.DBHierarchy.Name + " WHERE "
	paramsNew[tool.RootSQLFilterParam] += entities.RootID(entities.DBUser.Name) + " IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT id FROM " + entities.DBUser.Name + " WHERE login=" + conn.Quote(s.Domain.(*MainService).User) + ")"
	paramsNew[tool.RootSQLFilterParam] += " OR " + entities.RootID(entities.DBEntity.Name) + " IN ("
	paramsNew[tool.RootSQLFilterParam] += "SELECT " + entities.RootID(entities.DBEntity.Name) + " FROM " + entities.DBEntityUser.Name
	paramsNew[tool.RootSQLFilterParam] += " WHERE " + entities.RootID(entities.DBUser.Name) + "=" + conn.Quote(s.Domain.(*MainService).User) + "))"
	// TODO CHECK IF HIERARCHY FROM ENTITY OR USER
	hierarchy, err := s.Domain.SafeCall(true, "", 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
					)
	if err == nil {
		for _, upper := range hierarchy {
			s.Domain.SafeCall(true, "",
							tool.Params{ 
								entities.RootID(entities.DBUser.Name) : fmt.Sprintf("%d", upper["id"].(int64)),
								entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%d", record[entities.RootID(entities.DBTask.Name)].(int64)),
							},
							tool.Record{ entities.RootID(entities.DBUser.Name) : upper["id"],
										 entities.RootID(entities.DBTask.Name) : record[entities.RootID(entities.DBTask.Name)].(int64),
										 "assignee" : false,
										 "watcher" : true,
										 "verifyer" : false,
									   },
								  tool.CREATE,
								  "CreateOrUpdate",
							)
		}
	}
	
}
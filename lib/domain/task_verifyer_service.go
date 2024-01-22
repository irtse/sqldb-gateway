package domain

import (
	"fmt"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/infrastructure/entities"
)

type TaskVerifyerService struct { Domain tool.DomainITF }

func (s *TaskVerifyerService) SetDomain(d tool.DomainITF) { s.Domain = d }
func (s *TaskVerifyerService) Entity() tool.SpecializedServiceInfo { return entities.DBTaskVerifyer }
func (s *TaskVerifyerService) VerifyRowWorkflow(record tool.Record, create bool) (tool.Record, bool) { 
	var res tool.Results
	if taskID, ok := record[entities.RootID(entities.DBTask.Name)]; ok && taskID != nil {
		if userID, ok := record[entities.RootID(entities.DBUser.Name)]; ok && userID != nil {
			res, _ = s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBTask.Name) : fmt.Sprintf("%d", record[entities.RootID(entities.DBTask.Name)].(int64)),
						 entities.RootID(entities.DBUser.Name): fmt.Sprintf("%d", record[entities.RootID(entities.DBUser.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		} else if entityID, ok := record[entities.RootID(entities.DBEntity.Name)]; ok && entityID != nil {
			res, _ = s.Domain.SafeCall(true, "",
			tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
						 tool.RootRowsParam : tool.ReservedParam, 
						 entities.RootID(entities.DBEntity.Name): fmt.Sprintf("%d", record[entities.RootID(entities.DBEntity.Name)].(int64)) }, 
			tool.Record{}, 
			tool.SELECT, 
			"Get")
		}
	}
	return record, res == nil || len(res) == 0
}
func (s *TaskVerifyerService) DeleteRowWorkflow(results tool.Results) { }
func (s *TaskVerifyerService) UpdateRowWorkflow(results tool.Results, record tool.Record) {
	if state, ok := record["state"]; ok && state != "complete" { return }
	for _, rec := range results {
		if id, ok2 := rec[entities.RootID(entities.DBTask.Name)]; ok2 {
			paramsNew := tool.Params{ tool.RootTableParam : entities.DBTaskVerifyer.Name, 
				                      tool.RootRowsParam: tool.ReservedParam, }
			paramsNew[entities.RootID(entities.DBTask.Name)] = fmt.Sprintf("%d", id.(int64))
			paramsNew[tool.RootSQLFilterParam] = "state != 'complete'"
			unfinished, err := s.Domain.SafeCall(true, "", 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
					)
			if len(unfinished) > 0 || err != nil  { continue }
			paramsNew = tool.Params{ tool.RootTableParam : entities.DBTaskAssignee.Name, 
				                      tool.RootRowsParam: tool.ReservedParam, }
			paramsNew[entities.RootID(entities.DBTask.Name)] = fmt.Sprintf("%d", id.(int64))
			paramsNew[tool.RootSQLFilterParam] = "state != 'complete'"
			unfinishedAssign, err := s.Domain.SafeCall(true, "", 
						paramsNew, 
						tool.Record{},
						tool.SELECT,
						"Get",
					)
			if len(unfinishedAssign) > 0  || err != nil { continue }
			// TODO when all is verified
			s.Domain.SafeCall(true, "",
							tool.Params{ 
								tool.RootTableParam : entities.DBTask.Name,
								tool.RootRowsParam : tool.ReservedParam,
							},
							tool.Record{ "id" : id, "state" : "close" },
							tool.UPDATE,
							"CreateOrUpdate",
						)
		}
	}
}
func (s *TaskVerifyerService) WriteRowWorkflow(record tool.Record) {}
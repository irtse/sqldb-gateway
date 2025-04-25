package task_service

import (
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	servutils "sqldb-ws/domain/specialized_service/utils"
	utils "sqldb-ws/domain/utils"
)

// DONE - UNDER 100 LINES - NOT TESTED
type WorkflowService struct {
	servutils.AbstractSpecializedService
}

func (s *WorkflowService) Entity() utils.SpecializedServiceInfo { return ds.DBWorkflow }

func (s *WorkflowService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	res := utils.Results{}
	for _, rec := range results { // filter by allowed schemas
		schema, err := schema.GetSchemaByID(utils.ToInt64(rec[SchemaDBField]))
		if err == nil && s.Domain.VerifyAuth(schema.Name, "", "", utils.CREATE) {
			if !(!schema.HasField(sm.NAMEKEY) && !s.Domain.IsSuperAdmin()) {
				res = append(res, rec)
			}
		}
	}
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(res, tableName, true, s.Domain.GetParams().Copy())
}

func (s *WorkflowService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

func (s *WorkflowService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
	} else {
		return record, err, false
	}
}

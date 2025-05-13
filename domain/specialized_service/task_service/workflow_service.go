package task_service

import (
	"fmt"
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
	fmt.Println("TH1", len(results))
	res := utils.Results{}
	for _, rec := range results { // filter by allowed schemas
		schema, err := schema.GetSchemaByID(utils.ToInt64(rec[SchemaDBField]))
		if err == nil && s.Domain.VerifyAuth(schema.Name, "", "", utils.CREATE) {
			if !(!schema.HasField(sm.NAMEKEY) && !s.Domain.IsSuperAdmin()) {
				res = append(res, rec)
			}
		}
	}
	fmt.Println("TH2", len(res))
	rr := view_convertor.NewViewConvertor(s.Domain).TransformToView(res, tableName, true, s.Domain.GetParams().Copy())
	if _, ok := s.Domain.GetParams().Get(utils.SpecialIDParam); ok && len(results) == 1 && len(rr) == 1 {
		r := results[0]
		if i, ok := r[ds.FilterDBField]; ok {
			schema := rr[0]["schema"].(map[string]interface{})
			newSchema := map[string]interface{}{}
			if fields, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBSchemaField.Name,
				map[string]interface{}{
					utils.SpecialIDParam: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBFilterField.Name,
						map[string]interface{}{
							ds.FilterDBField: i,
						}, false, ds.SchemaFieldDBField)}, false); err == nil {
				for _, f := range fields {
					newSchema[utils.GetString(f, "name")] = schema[utils.GetString(f, "name")]
				}
			}
			rr[0]["schema"] = newSchema
		}
	}
	fmt.Println("TH3", rr)
	return rr
}

func (s *WorkflowService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	s1, s2, s3, s4 := filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
	return s1, s2, s3, s4
}

func (s *WorkflowService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if rec, err, ok := servutils.CheckAutoLoad(tablename, record, s.Domain); ok {
		return s.AbstractSpecializedService.VerifyDataIntegrity(rec, tablename)
	} else {
		return record, err, false
	}
}

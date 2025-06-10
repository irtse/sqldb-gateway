package utils

import (
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/domain_service/view_convertor"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/service"
	"strings"
)

type AbstractSpecializedService struct {
	service.InfraSpecializedService
	Domain utils.DomainITF
}

func (s *AbstractSpecializedService) SpecializedCreateRow(record map[string]interface{}, tablename string) {
	if s, ok := s.Domain.GetParams().Get(utils.RootRawView); ok && s == "enable" {
		return
	}
	sch, err := sch.GetSchema(tablename)
	if err == nil {
		triggers.NewTrigger(s.Domain).Trigger(sch, record, utils.CREATE)
	}
}

func (s *AbstractSpecializedService) SpecializedUpdateRow(res []map[string]interface{}, record map[string]interface{}) {
	if s, ok := s.Domain.GetParams().Get(utils.RootRawView); ok && s == "enable" {
		return
	}
	sch, err := sch.GetSchema(s.Domain.GetTable())
	if err == nil {
		for _, rec := range res {
			triggers.NewTrigger(s.Domain).Trigger(sch, record, utils.UPDATE)
			if s.Domain.GetTable() == ds.DBRequest.Name || s.Domain.GetTable() == ds.DBTask.Name {
				continue
			}
			if reqs, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
				ds.DestTableDBField: rec[utils.SpecialIDParam],
				ds.SchemaDBField:    sch.ID,
			}, false); err == nil && len(reqs) == 0 {
				if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflowSchema.Name, map[string]interface{}{
					ds.WorkflowDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
						ds.SchemaDBField: sch.ID,
					}, false, utils.SpecialIDParam),
				}, false); err == nil && len(res) > 0 {
					s.Domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name).RootRaw(), map[string]interface{}{
						ds.WorkflowDBField:  res[0][ds.WorkflowDBField],
						ds.DestTableDBField: rec[utils.SpecialIDParam],
						ds.SchemaDBField:    sch.ID,
					})
				}
			}
		}
	}
}

func (s *AbstractSpecializedService) VerifyDataIntegrity(record map[string]interface{}, tablename string) (map[string]interface{}, error, bool) {
	if s.Domain.GetAutoload() {
		return record, nil, true
	}
	if sch, err := sch.GetSchema(tablename); err != nil {
		return record, errors.New("no schema found"), false
	} else {
		if e, ok := record[ds.EntityDBField]; ok && e == nil && sch.HasField(ds.EntityDBField) {
			if res, err := s.Domain.GetDb().CreateQuery(ds.DBEntity.Name, map[string]interface{}{
				"name": record["name"],
			}, func(s string) (string, bool) { return "", true }); err == nil {
				record[ds.EntityDBField] = res
			}
		}
		for k, v := range record {
			if f, err := sch.GetField(k); err == nil && f.Transform != "" {
				if f.Transform == "lowercase" {
					record[k] = strings.ToLower(utils.ToString(v))
				} else if f.Transform == "uppercase" {
					record[k] = strings.ToUpper(utils.ToString(v))
				}
			}
		}
		if _, ok := record["is_draft"]; !ok || !utils.GetBool(record, "is_draft") {
			if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBConsentResponse.Name, map[string]interface{}{
				ds.ConsentDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBConsent.Name, map[string]interface{}{
					ds.SchemaDBField: sch.ID,
					"optionnal":      false,
				}, false, "id"),
				"is_consenting": false,
			}, false); err == nil && len(res) > 0 {
				return record, errors.New("should consent"), false
			}
		}
	}
	return record, nil, true
}

func (s *AbstractSpecializedService) SetDomain(d utils.DomainITF) { s.Domain = d }

type SpecializedService struct{ AbstractSpecializedService }

func (s *SpecializedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}
func (s *SpecializedService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

func (s *SpecializedService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for _, r := range results {
		if schema, err := sch.GetSchema(tableName); err == nil {
			for _, db := range []string{ds.DBRequest.Name, ds.DBTask.Name, ds.DBNotification.Name, ds.DBDataAccess.Name, ds.DBShare.Name} {
				s.Domain.GetDb().DeleteQueryWithRestriction(db, map[string]interface{}{
					ds.SchemaDBField:    schema.ID,
					ds.DestTableDBField: utils.GetInt(r, utils.SpecialIDParam),
				}, false)
			}
		}
	}
}

func CheckAutoLoad(tablename string, record utils.Record, domain utils.DomainITF) (utils.Record, error, bool) {
	if domain.GetMethod() != utils.DELETE {
		rec, err := sch.ValidateBySchema(record, tablename, domain.GetMethod(), domain.VerifyAuth)
		return rec, err, err == nil
	}
	return record, nil, false
}

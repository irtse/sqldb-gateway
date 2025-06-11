package utils

import (
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/service"
	"strings"
	"time"
)

type AbstractSpecializedService struct {
	service.InfraSpecializedService
	Domain     utils.DomainITF
	ManyToMany map[string][]map[string]interface{}
	OneToMany  map[string][]map[string]interface{}
}

func (s *AbstractSpecializedService) Entity() utils.SpecializedServiceInfo {
	return nil
}

func (s *AbstractSpecializedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}

func (s *AbstractSpecializedService) SpecializedCreateRow(record map[string]interface{}, tablename string) {
	if s, ok := s.Domain.GetParams().Get(utils.RootRawView); ok && s == "enable" {
		return
	}

	sch, err := sch.GetSchema(tablename)
	if err == nil {
		for schemaName, mm := range s.ManyToMany {
			field, err := sch.GetField(schemaName)
			if err != nil {
				continue
			}
			if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
				for _, m := range mm {
					if m[utils.SpecialIDParam] != nil && m[ds.RootID(ff.Name)] == nil {
						m[ds.RootID(ff.Name)] = m[utils.SpecialIDParam]
						delete(m, utils.SpecialIDParam)
					} else if m[utils.SpecialIDParam] == nil && m[ds.RootID(ff.Name)] == nil {
						continue
					}
					m[ds.RootID(tablename)] = record[utils.SpecialIDParam]
					s.Domain.CreateSuperCall(utils.AllParams(ff.Name), m)
				}
			}
		}
		for schemaName, om := range s.OneToMany {
			field, err := sch.GetField(schemaName)
			if err != nil {
				continue
			}
			if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
				for _, m := range om {
					m[ds.RootID(tablename)] = record[utils.SpecialIDParam]
					s.Domain.CreateSuperCall(utils.AllParams(ff.Name), m)
				}
			}
		}
		triggers.NewTrigger(s.Domain).Trigger(&sch, record, utils.CREATE)
	}
}

func (s *AbstractSpecializedService) SpecializedUpdateRow(res []map[string]interface{}, record map[string]interface{}) {
	if s, ok := s.Domain.GetParams().Get(utils.RootRawView); ok && s == "enable" {
		return
	}
	sch, err := sch.GetSchema(s.Domain.GetTable())
	if err == nil {
		for _, rec := range res {
			for schemaName, mm := range s.ManyToMany {
				field, err := sch.GetField(schemaName)
				if err != nil {
					continue
				}
				if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
					s.Domain.GetDb().DeleteQueryWithRestriction(ff.Name, map[string]interface{}{
						ds.RootID(s.Domain.GetTable()): rec[utils.SpecialIDParam],
					}, false)
					for _, m := range mm {
						if m[utils.SpecialIDParam] != nil && m[ds.RootID(ff.Name)] == nil {
							m[ds.RootID(ff.Name)] = m[utils.SpecialIDParam]
							delete(m, utils.SpecialIDParam)
						} else if m[utils.SpecialIDParam] == nil && m[ds.RootID(ff.Name)] == nil {
							continue
						}
						m[ds.RootID(s.Domain.GetTable())] = rec[utils.SpecialIDParam]
						s.Domain.GetDb().ClearQueryFilter().CreateQuery(ff.Name, m, func(s string) (string, bool) { return "", true })
					}
				}
			}
			for schemaName, om := range s.OneToMany {
				field, err := sch.GetField(schemaName)
				if err != nil {
					continue
				}
				if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
					s.Domain.GetDb().DeleteQueryWithRestriction(ff.Name, map[string]interface{}{
						ds.RootID(s.Domain.GetTable()): rec[utils.SpecialIDParam],
					}, false)
					for _, m := range om {
						m[ds.RootID(s.Domain.GetTable())] = rec[utils.SpecialIDParam]
						s.Domain.GetDb().ClearQueryFilter().CreateQuery(ff.Name, m, func(s string) (string, bool) { return "", true })
					}
				}
			}
			triggers.NewTrigger(s.Domain).Trigger(&sch, record, utils.UPDATE)
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
		currentTime := time.Now()
		sqlFilter := "'" + currentTime.Format("2000-01-01") + "' < start_date OR "
		sqlFilter += "'" + currentTime.Format("2000-01-01") + "' > end_date"
		p := utils.AllParams(tablename).RootRaw()
		s.Domain.SuperCall(p, utils.Record{}, utils.DELETE, false, sqlFilter)
		if s.Domain.GetMethod() == utils.CREATE || s.Domain.GetMethod() == utils.UPDATE { // stock oneToMany and ManyToMany
			for _, field := range sch.Fields {
				if strings.ToUpper(field.Type) == sm.MANYTOMANY.String() && record[field.Name] != nil {
					if s.ManyToMany[field.Name] == nil {
						s.ManyToMany[field.Name] = []map[string]interface{}{}
					}
					for _, mm := range utils.ToList(record[field.Name]) {
						s.ManyToMany[field.Name] = append(s.ManyToMany[field.Name], utils.ToMap(mm))
					}
					delete(record, field.Name)
				} else if strings.ToUpper(field.Type) == sm.ONETOMANY.String() && record[field.Name] != nil {
					if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
						if s.OneToMany[ff.Name] == nil {
							s.OneToMany[ff.Name] = []map[string]interface{}{}
						}
						for _, mm := range utils.ToList(record[field.Name]) {
							s.OneToMany[field.Name] = append(s.OneToMany[field.Name], utils.ToMap(mm))
						}
					}
					delete(record, field.Name)
				}
			}
		}

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

func (s *AbstractSpecializedService) SetDomain(d utils.DomainITF) utils.SpecializedServiceITF {
	s.Domain = d
	return s
}

type SpecializedService struct{ AbstractSpecializedService }

func (s *SpecializedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	return view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
}
func (s *SpecializedService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

func (s *SpecializedService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	if schema, err := sch.GetSchema(tableName); err == nil { // CASCADE DEL ON NON REF DATA fusion of schema + dest table id
		if res, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBSchema.Name, map[string]interface{}{}, true); err == nil {
			for _, scheme := range res {
				for _, r := range results {
					if sch, err := sch.GetSchema(utils.GetString(scheme, "name")); err == nil && sch.HasField(ds.SchemaDBField) && sch.HasField(ds.DestTableDBField) {
						s.Domain.GetDb().DeleteQueryWithRestriction(sch.Name, map[string]interface{}{
							ds.SchemaDBField:    schema.ID,
							ds.DestTableDBField: utils.GetInt(r, utils.SpecialIDParam),
						}, false)
					}
				}
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

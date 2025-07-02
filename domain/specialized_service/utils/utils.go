package utils

import (
	"errors"
	"sqldb-ws/domain/domain_service/filter"
	"sqldb-ws/domain/domain_service/triggers"
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/schema"
	sch "sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/schema/models"
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
	t := view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
	if scheme, err := sch.GetSchema(tableName); err == nil {
		if s.Domain.GetMethod() == utils.CREATE && len(results) == 1 && utils.GetBool(results[0], "is_draft") {
			for _, tt := range t {
				tt["inner_redirection"] = utils.BuildPath(scheme.ID, utils.GetString(results[0], "id"))
			}
		}

		if s.Domain.GetMethod() == utils.DELETE && scheme.ViewIDOnDelete != "" {
			for _, tt := range t {
				tt["inner_redirection"] = utils.BuildPath(ds.DBView.Name, scheme.ViewIDOnDelete)
			}
		}
		if s.Domain.GetIsDraftToPublished() && len(results) == 1 {
			if rr, err := s.Domain.GetDb().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.RequestDBField: s.Domain.GetDb().BuildSelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
					ds.DestTableDBField: results[0][utils.SpecialIDParam],
					ds.SchemaDBField:    scheme.ID,
					"is_close":          false,
				}, false, utils.SpecialIDParam),
			}, false); err == nil && len(rr) > 0 {
				if ss, err := sch.GetSchema(ds.DBTask.Name); err == nil {
					for _, tt := range t {
						tt["inner_redirection"] = utils.BuildPath(ss.ID, utils.GetString(rr[0], utils.SpecialIDParam))
					}
				}
			}
		}
	}
	return t
}

func (s *AbstractSpecializedService) SpecializedCreateRow(record map[string]interface{}, tablename string) {
	if sch, err := sch.GetSchema(tablename); err == nil {
		for schemaName, mm := range s.ManyToMany {
			field, err := sch.GetField(schemaName)
			if err != nil {
				continue
			}
			if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
				for _, m := range mm {
					if m[utils.SpecialIDParam] != nil && m[ds.RootID(ff.Name)] == nil {
						if ff.HasField(ds.RootID(ff.Name)) {
							m[ds.RootID(ff.Name)] = m[utils.SpecialIDParam]
							delete(m, utils.SpecialIDParam)
						} else {
							for _, fff := range ff.Fields {
								if !strings.Contains(fff.Name, ff.Name) && fff.GetLink() > 0 {
									m[fff.Name] = m[utils.SpecialIDParam]
									delete(m, utils.SpecialIDParam)
									break
								}
							}
						}
					} else if m[utils.SpecialIDParam] == nil && m[ds.RootID(ff.Name)] == nil {
						continue
					}
					m[ds.RootID(tablename)] = record[utils.SpecialIDParam]
					s.Domain.CreateSuperCall(utils.AllParams(ff.Name).RootRaw(), m)
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
					s.Domain.CreateSuperCall(utils.AllParams(ff.Name).RootRaw(), m)
				}
			}
		}
		triggers.NewTrigger(s.Domain).Trigger(&sch, record, utils.CREATE)
	}
}

func (s *AbstractSpecializedService) SpecializedUpdateRow(res []map[string]interface{}, record map[string]interface{}) {
	sche, err := sch.GetSchema(s.Domain.GetTable())
	if err == nil {

		for _, rec := range res {
			for _, field := range sche.Fields {
				if field.GetLink() == 0 {
					continue
				}
				if sch2, err := sch.GetSchemaByID(field.GetLink()); err == nil && strings.Contains(strings.ToUpper(field.Type), "MANY") {
					s.Domain.DeleteSuperCall(utils.AllParams(sch2.Name).Enrich(map[string]interface{}{
						ds.RootID(s.Domain.GetTable()): rec[utils.SpecialIDParam],
					}).RootRaw())
				}
			}
			for schemaName, mm := range s.ManyToMany {
				field, err := sche.GetField(schemaName)
				if err != nil {
					continue
				}
				if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
					for _, m := range mm {
						if m[utils.SpecialIDParam] != nil && m[ds.RootID(ff.Name)] == nil {
							if ff.HasField(ds.RootID(ff.Name)) {
								m[ds.RootID(ff.Name)] = m[utils.SpecialIDParam]
								delete(m, utils.SpecialIDParam)
							} else {
								for _, fff := range ff.Fields {
									if !strings.Contains(fff.Name, ff.Name) && fff.GetLink() > 0 {
										m[fff.Name] = m[utils.SpecialIDParam]
										delete(m, utils.SpecialIDParam)
										break
									}
								}
							}
						} else if m[utils.SpecialIDParam] == nil && m[ds.RootID(ff.Name)] == nil {
							continue
						}
						m[ds.RootID(s.Domain.GetTable())] = record[utils.SpecialIDParam]
						s.Domain.CreateSuperCall(utils.AllParams(ff.Name).RootRaw(), m)
					}
				}
			}
			for schemaName, om := range s.OneToMany {
				field, err := sche.GetField(schemaName)
				if err != nil {
					continue
				}
				if ff, err := schema.GetSchemaByID(field.GetLink()); err == nil {
					for _, m := range om {
						m[ds.RootID(s.Domain.GetTable())] = record[utils.SpecialIDParam]
						s.Domain.CreateSuperCall(utils.AllParams(ff.Name).RootRaw(), m)
					}
				}
			}
			triggers.NewTrigger(s.Domain).Trigger(&sche, record, utils.UPDATE)
			if s.Domain.GetTable() == ds.DBRequest.Name || s.Domain.GetTable() == ds.DBTask.Name {
				continue
			}
			if reqs, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
				ds.DestTableDBField: rec[utils.SpecialIDParam],
				ds.SchemaDBField:    sche.ID,
			}, false); err == nil && len(reqs) == 0 {
				if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBWorkflow.Name, map[string]interface{}{
					ds.SchemaDBField: sche.ID,
				}, false); err == nil && len(res) > 0 {
					s.Domain.CreateSuperCall(utils.AllParams(ds.DBRequest.Name).RootRaw(), map[string]interface{}{
						ds.WorkflowDBField:  res[0][utils.SpecialIDParam],
						ds.DestTableDBField: rec[utils.SpecialIDParam],
						ds.SchemaDBField:    sche.ID,
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
		if sch.HasField("start_date") && sch.HasField("end_date") {
			sqlFilter := "'" + currentTime.Format("2000-01-01") + "' < start_date OR "
			sqlFilter += "'" + currentTime.Format("2000-01-01") + "' > end_date"
			db := s.Domain.GetDb()
			db.ClearQueryFilter().SQLRestriction = sqlFilter
			db.DeleteQueryWithRestriction(tablename, map[string]interface{}{}, false)
		}

		if s.Domain.GetMethod() == utils.CREATE || s.Domain.GetMethod() == utils.UPDATE { // stock oneToMany and ManyToMany
			s.ManyToMany = map[string][]map[string]interface{}{}
			s.OneToMany = map[string][]map[string]interface{}{}
			for _, field := range sch.Fields {
				if strings.Contains(strings.ToUpper(field.Type), strings.ToUpper(sm.MANYTOMANY.String())) && record[field.Name] != nil {
					if s.ManyToMany[field.Name] == nil {
						s.ManyToMany[field.Name] = []map[string]interface{}{}
					}
					for _, mm := range utils.ToList(record[field.Name]) {
						s.ManyToMany[field.Name] = append(s.ManyToMany[field.Name], utils.ToMap(mm))
					}
					delete(record, field.Name)
				} else if strings.Contains(strings.ToUpper(field.Type), strings.ToUpper(sm.ONETOMANY.String())) && record[field.Name] != nil {
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
			if e, ok := record[ds.EntityDBField]; ok && e == nil && sch.HasField(ds.EntityDBField) {
				if res, err := s.Domain.GetDb().CreateQuery(ds.DBEntity.Name, map[string]interface{}{
					"name": record["name"],
				}, func(s string) (string, bool) { return "", true }); err == nil {
					record[ds.EntityDBField] = res
				}
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
		/*if _, ok := record["is_draft"]; ok && !utils.GetBool(record, "is_draft") {
			if res, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBConsent.Name, map[string]interface{}{
				ds.SchemaDBField: sch.ID,
				"optionnal":      false,
			}, false); err == nil {
				for _, r := range res {
					if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBConsentResponse.Name, map[string]interface{}{
						ds.ConsentDBField:   r[utils.SpecialIDParam],
						ds.DestTableDBField: record[utils.SpecialIDParam],
						ds.SchemaDBField:    sch.ID,
						"is_consenting":     true,
					}, false); err == nil && len(rr) == 0 {
						return record, errors.New("should consent"), false
					}
				}
			}
		}*/
	}
	return record, nil, true
}

func (s *AbstractSpecializedService) SetDomain(d utils.DomainITF) utils.SpecializedServiceITF {
	s.Domain = d
	return s
}

type SpecializedService struct{ AbstractSpecializedService }

func (s *SpecializedService) TransformToGenericView(results utils.Results, tableName string, dest_id ...string) utils.Results {
	t := view_convertor.NewViewConvertor(s.Domain).TransformToView(results, tableName, true, s.Domain.GetParams().Copy())
	if scheme, err := sch.GetSchema(tableName); err == nil {
		if s.Domain.GetMethod() == utils.CREATE && len(results) == 1 && utils.GetBool(results[0], "is_draft") {
			for _, tt := range t {
				tt["inner_redirection"] = utils.BuildPath(scheme.ID, utils.GetString(results[0], "id"))
			}
		}
		if s.Domain.GetMethod() == utils.DELETE && scheme.ViewIDOnDelete != "" {
			for _, tt := range t {
				tt["inner_redirection"] = utils.BuildPath(ds.DBView.Name, scheme.ViewIDOnDelete)
			}
		}
		if s.Domain.GetIsDraftToPublished() && len(results) == 1 {
			if rr, err := s.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
				ds.RequestDBField: s.Domain.GetDb().ClearQueryFilter().BuildSelectQueryWithRestriction(ds.DBRequest.Name, map[string]interface{}{
					ds.DestTableDBField: results[0][utils.SpecialIDParam],
					ds.SchemaDBField:    scheme.ID,
					"is_close":          false,
				}, false, utils.SpecialIDParam),
			}, false); err == nil && len(rr) > 0 {
				if ss, err := sch.GetSchema(ds.DBTask.Name); err == nil {
					for _, tt := range t {
						tt["inner_redirection"] = utils.BuildPath(ss.ID, utils.GetString(rr[0], utils.SpecialIDParam))
					}
				}
			}
		}
	}
	return t
}
func (s *SpecializedService) GenerateQueryFilter(tableName string, innerestr ...string) (string, string, string, string) {
	return filter.NewFilterService(s.Domain).GetQueryFilter(tableName, s.Domain.GetParams().Copy(), innerestr...)
}

func (s *SpecializedService) SpecializedDeleteRow(results []map[string]interface{}, tableName string) {
	for _, sch := range models.SchemaRegistry {
		for _, r := range results {
			if r[ds.SchemaDBField] != nil && r[ds.DestTableDBField] != nil {
				s.Domain.DeleteSuperCall(utils.AllParams(sch.Name).Enrich(map[string]interface{}{
					ds.SchemaDBField:    r[ds.SchemaDBField],
					ds.DestTableDBField: r[ds.DestTableDBField],
				}), false)
			}
			if sch.HasField(ds.SchemaDBField) && sch.HasField(ds.DestTableDBField) {
				s.Domain.GetDb().ClearQueryFilter().DeleteQueryWithRestriction(sch.Name, map[string]interface{}{
					ds.SchemaDBField:    sch.ID,
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

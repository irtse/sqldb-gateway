package triggers

import (
	"fmt"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	conn "sqldb-ws/infrastructure/connector/db"
	"strings"
)

type TriggerService struct {
	Domain utils.DomainITF
}

func NewTrigger(domain utils.DomainITF) *TriggerService {
	return &TriggerService{
		Domain: domain,
	}
}

func (t *TriggerService) GetViewTriggers(record utils.Record, method utils.Method, fromSchema *sm.SchemaModel, toSchemaID, destID int64) []sm.ManualTriggerModel {
	if _, ok := t.Domain.GetParams().Get(utils.SpecialIDParam); method == utils.DELETE || (!ok && method == utils.SELECT) {
		return []sm.ManualTriggerModel{}
	}
	if utils.UPDATE == method && t.Domain.GetIsDraftToPublished() {
		method = utils.CREATE
	}
	mt := []sm.ManualTriggerModel{}
	if res, err := t.GetTriggers("manual", method, fromSchema.ID); err == nil {
		for _, r := range res {
			typ := utils.GetString(r, "type")
			switch typ {
			case "mail":
				if t, err := t.GetViewMailTriggers(record, fromSchema, utils.GetString(r, "description"), utils.GetString(r, "name"),
					utils.GetInt(r, utils.SpecialIDParam), toSchemaID, destID); err == nil {
					mt = append(mt, t...)
				}
			}
		}
	}
	return mt
}

func (t *TriggerService) GetTriggers(mode string, method utils.Method, fromSchemaID string) ([]map[string]interface{}, error) {
	if method == utils.SELECT {
		method = utils.CREATE
	}
	return t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTrigger.Name, map[string]interface{}{
		"on_" + method.String(): true,
		"mode":                  conn.Quote(mode),
		ds.SchemaDBField:        fromSchemaID,
	}, false)
}

func (t *TriggerService) Trigger(fromSchema *sm.SchemaModel, record utils.Record, method utils.Method) {
	if t.Domain.GetAutoload() {
		return
	}
	if res, err := t.GetTriggers("auto", method, fromSchema.ID); err == nil {
		fmt.Println("FOUND TRIGGERS", len(res), fromSchema.Name)
		for i, r := range res {
			typ := utils.GetString(r, "type")
			fmt.Println("START TRIGGER", i, typ, r)
			switch typ {
			case "mail":
				t.triggerMail(record, fromSchema,
					utils.GetInt(r, utils.SpecialIDParam),
					utils.GetInt(record, ds.SchemaDBField),
					utils.GetInt(record, ds.DestTableDBField))
			case "sms":
				break
			case "teams notification":
				break
			case "data":
				t.triggerData(record, fromSchema,
					utils.GetInt(r, utils.SpecialIDParam),
					utils.GetInt(record, ds.SchemaDBField),
					utils.GetInt(record, ds.DestTableDBField))
			}
		}
	}
}
func (t *TriggerService) ParseMails(toSplit string) []map[string]interface{} {
	splitted := ""
	if len(strings.Split(toSplit, ";")) > 0 {
		splitted = strings.ReplaceAll(strings.Join(strings.Split(toSplit, ";"), ","), " ", "")
	} else if len(strings.Split(toSplit, ",")) > 0 {
		splitted = strings.ReplaceAll(toSplit, " ", "")
	} else if len(strings.Split(toSplit, " ")) > 0 {
		splitted = strings.ReplaceAll(strings.Join(strings.Split(toSplit, ","), ","), " ", "")
	}
	if len(splitted) > 0 {
		s := []string{}
		for _, ss := range strings.Split(splitted, ",") {
			s = append(s, conn.Quote(ss))
		}
		if res, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			"email": s,
		}, false); err == nil {
			return res
		}
	}
	return []map[string]interface{}{}
}

// send_mail_to should be on request + task
func (t *TriggerService) handleOverrideEmailTo(record, dest map[string]interface{}) []map[string]interface{} {
	if record["send_mail_to"] != nil { // it's a particular default field that detect overriding {
		return t.ParseMails(utils.GetString(record, "send_mail_to"))
	} else if dest["send_mail_to"] != nil {
		return t.ParseMails(utils.GetString(dest, "send_mail_to"))
	} else if userID, ok := record[ds.UserDBField]; ok {
		if usto, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: userID,
		}, false); err == nil && len(usto) > 0 {
			return []map[string]interface{}{usto[0]}
		}
	}
	return []map[string]interface{}{}
}

func (t *TriggerService) triggerMail(record utils.Record, fromSchema *sm.SchemaModel, triggerID, toSchemaID, destID int64) {
	for _, mail := range t.TriggerManualMail("auto", record, fromSchema, triggerID, toSchemaID, destID) {
		t.Domain.CreateSuperCall(utils.AllParams(ds.DBEmailSended.Name).RootRaw(), mail)
	}
}

func (t *TriggerService) triggerData(record utils.Record, fromSchema *sm.SchemaModel, triggerID, toSchemaID, destID int64) {
	if toSchemaID < 0 || destID < 0 {
		toSchemaID = utils.ToInt64(fromSchema.ID)
		destID = utils.GetInt(record, utils.SpecialIDParam)
	}
	// PROBLEM WE CAN'T DECOLERATE and action on not a sub data of it. (not a problem for now)

	rules := t.GetTriggerRules(triggerID, fromSchema, toSchemaID, record)
	for _, r := range rules {
		if toSchemaID != utils.GetInt(r, "to_"+ds.SchemaDBField) {
			continue
		}

		toSchema, err := schema.GetSchemaByID(toSchemaID)
		if err != nil {
			continue
		}

		field, err := toSchema.GetFieldByID(utils.GetInt(r, "to_"+ds.SchemaFieldDBField))
		if err != nil {
			continue
		}

		value := utils.GetString(r, "value")
		if value == "" {
			value = utils.GetString(record, field.Name)
		}
		fmt.Println("UPDATE DATA", map[string]interface{}{
			field.Name: value,
		}, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, toSchema.Name)
		t.Domain.GetDb().ClearQueryFilter().UpdateQuery(toSchema.Name, map[string]interface{}{
			field.Name: value,
		}, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, false)
	}
}

func (t *TriggerService) GetTriggerRules(triggerID int64, fromSchema *sm.SchemaModel, toSchemaID int64, record utils.Record) []map[string]interface{} {
	if res, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTriggerCondition.Name, map[string]interface{}{
		ds.TriggerDBField: triggerID,
	}, false); err == nil && len(res) > 0 {
		for _, cond := range res {
			if cond[ds.SchemaFieldDBField] == nil && utils.GetString(record, utils.SpecialIDParam) != utils.GetString(cond, "value") {
				return []map[string]interface{}{}
			}
			if f, err := fromSchema.GetFieldByID(utils.GetInt(cond, ds.SchemaFieldDBField)); err != nil || utils.GetString(record, f.Name) != utils.GetString(cond, "value") {
				return []map[string]interface{}{}
			}
		}
	}
	rules, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTriggerRule.Name, map[string]interface{}{
		ds.TriggerDBField:        triggerID,
		"to_" + ds.SchemaDBField: toSchemaID,
	}, false)
	if err != nil {
		return []map[string]interface{}{}
	}
	return rules
}

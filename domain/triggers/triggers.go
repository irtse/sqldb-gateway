package triggers

import (
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
)

type TriggerService struct {
	Domain utils.DomainITF
}

func NewTrigger(domain utils.DomainITF) *TriggerService {
	return &TriggerService{
		Domain: domain,
	}
}

func (t *TriggerService) Trigger(fromSchema sm.SchemaModel, record utils.Record, method utils.Method) {
	_, gotDestID := record[ds.DestTableDBField]
	schemaID, gotSchema := record[ds.SchemaDBField]
	if !gotDestID || !gotSchema {
		return
	}
	if res, err := t.Domain.GetDb().SelectQueryWithRestriction(ds.DBTrigger.Name, map[string]interface{}{
		"on_" + method.String(): true,
		ds.SchemaDBField:        schemaID,
	}, false); err == nil {
		for _, r := range res {
			typ := utils.GetString(r, "type")
			switch typ {
			case "mail":
				t.triggerMail(record, fromSchema, utils.GetInt(r, ds.TriggerDBField),
					utils.GetInt(record, ds.SchemaDBField), utils.GetInt(record, ds.DestTableDBField))
			case "sms":
				break
			case "teams notification":
				break
			case "data":
				t.triggerData(record, fromSchema,
					utils.GetInt(r, ds.TriggerDBField), utils.GetInt(record, ds.SchemaDBField), utils.GetInt(record, ds.DestTableDBField))
			}
		}
	}
}

func (t *TriggerService) triggerMail(record utils.Record, fromSchema sm.SchemaModel, triggerID, toSchemaID, destID int64) {
	userID, ok := record[ds.UserDBField]
	if !ok {
		return
	}

	mailSchema, err := schema.GetSchema(ds.DBEmailTemplate.Name)
	if err != nil || mailSchema.ID != utils.ToString(toSchemaID) {
		return
	}

	rules := t.getTriggerRules(triggerID, fromSchema.ID, toSchemaID)
	for _, r := range rules {
		if !t.passesCondition(fromSchema, r, record) {
			continue
		}

		mailID := r["value"]
		if mailID == nil {
			continue
		}

		mails, err := t.Domain.GetDb().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
			utils.SpecialIDParam: mailID,
		}, false)
		if err != nil || len(mails) == 0 {
			continue
		}

		usto, err := t.Domain.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: userID,
		}, false)
		if err != nil || len(usto) == 0 {
			continue
		}

		usfrom, err := t.Domain.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: t.Domain.GetUserID(),
		}, false)
		if err != nil || len(usfrom) == 0 {
			continue
		}

		toSchema, err := schema.GetSchemaByID(toSchemaID)
		if err != nil {
			continue
		}

		dest, err := t.Domain.GetDb().SelectQueryWithRestriction(toSchema.Name, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, false)
		if err != nil || len(dest) == 0 {
			continue
		}

		mail := mails[0]
		SendMail(
			utils.GetString(usfrom[0], utils.SpecialIDParam),
			utils.GetString(usto[0], utils.SpecialIDParam),
			utils.GetString(mail, "subject"),
			utils.GetString(mail, "template"),
			dest[0],
		)
	}
}

func (t *TriggerService) triggerData(record utils.Record, fromSchema sm.SchemaModel, triggerID, toSchemaID, destID int64) {
	if record[ds.DestTableDBField] == nil {
		return
	}
	rules := t.getTriggerRules(triggerID, fromSchema.ID, toSchemaID)
	for _, r := range rules {
		if !t.passesCondition(fromSchema, r, record) {
			continue
		}

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

		t.Domain.GetDb().UpdateQuery(toSchema.Name, map[string]interface{}{
			field.Name: value,
		}, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, false)
	}
}

func (t *TriggerService) getTriggerRules(triggerID int64, fromSchemaID string, toSchemaID int64) []map[string]interface{} {
	rules, err := t.Domain.GetDb().SelectQueryWithRestriction(ds.DBTriggerRule.Name, map[string]interface{}{
		ds.TriggerDBField:          triggerID,
		"from_" + ds.SchemaDBField: fromSchemaID,
		"to_" + ds.SchemaDBField:   toSchemaID,
	}, false)
	if err != nil {
		return nil
	}
	return rules
}

func (t *TriggerService) passesCondition(fromSchema sm.SchemaModel, rule utils.Record, record utils.Record) bool {
	ifval, okIfVal := rule["ifvalue"]
	colID, okCol := rule["from_"+ds.SchemaFieldDBField]
	if !okIfVal || !okCol {
		return true
	}

	field, err := fromSchema.GetFieldByID(utils.ToInt64(colID))
	if err != nil || record[field.Name] == ifval {
		return true
	}
	return false
}

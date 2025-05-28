package triggers

import (
	"fmt"
	"net/url"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
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

func (t *TriggerService) GetTriggers(mode string, method utils.Method, fromSchemaID string) ([]map[string]interface{}, error) {
	if method == utils.SELECT {
		method = utils.UPDATE
	}
	return t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBTrigger.Name, map[string]interface{}{
		"on_" + method.String(): true,
		"mode":                  connector.Quote(mode),
		ds.SchemaDBField:        fromSchemaID,
	}, false)
}

func (t *TriggerService) Trigger(fromSchema sm.SchemaModel, record utils.Record, method utils.Method) {
	if t.Domain.GetAutoload() {
		return
	}
	if res, err := t.GetTriggers("auto", method, fromSchema.ID); err == nil {
		for _, r := range res {
			typ := utils.GetString(r, "type")
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
			s = append(s, connector.Quote(ss))
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

func (t *TriggerService) triggerMail(record utils.Record, fromSchema sm.SchemaModel, triggerID, toSchemaID, destID int64) {
	for _, mail := range t.TriggerManualMail("auto", record, fromSchema, triggerID, toSchemaID, destID) {
		t.Domain.CreateSuperCall(utils.AllParams(ds.DBEmailSended.Name).RootRaw(), mail)
	}
}

func (t *TriggerService) triggerData(record utils.Record, fromSchema sm.SchemaModel, triggerID, toSchemaID, destID int64) {
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

		t.Domain.GetDb().ClearQueryFilter().UpdateQuery(toSchema.Name, map[string]interface{}{
			field.Name: value,
		}, map[string]interface{}{
			utils.SpecialIDParam: destID,
		}, false)
	}
}

func (t *TriggerService) GetTriggerRules(triggerID int64, fromSchema sm.SchemaModel, toSchemaID int64, record utils.Record) []map[string]interface{} {
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

func (t *TriggerService) TriggerManualMail(mode string, record utils.Record, fromSchema sm.SchemaModel, triggerID, toSchemaID, destID int64) []utils.Record {
	mailings := []utils.Record{}
	var err error
	var toSchema sm.SchemaModel
	dest := []map[string]interface{}{}
	if toSchemaID < 0 || destID < 0 {
		toSchema = fromSchema
		dest = []map[string]interface{}{record}
		toSchemaID = utils.ToInt64(fromSchema.ID)
		destID = utils.ToInt64(record[utils.SpecialIDParam])
	} else {
		toSchema, err = schema.GetSchemaByID(toSchemaID)
		if err == nil {
			if d, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(toSchema.Name, map[string]interface{}{
				utils.SpecialIDParam: destID,
			}, false); err == nil {
				dest = d
			}
		}
	}

	var toUsers []map[string]interface{}
	if len(dest) > 0 {
		if toUsers = t.handleOverrideEmailTo(record, dest[0]); len(toUsers) == 0 {
			if mode == "auto" {
				return mailings
			}
		}
	} else if toUsers = t.handleOverrideEmailTo(record, map[string]interface{}{}); len(toUsers) == 0 {
		if mode == "auto" {
			return mailings
		}
	}
	mailSchema, err := schema.GetSchema(ds.DBEmailTemplate.Name)
	if err != nil {
		return mailings
	}
	rules := t.GetTriggerRules(triggerID, fromSchema, mailSchema.GetID(), record)
	for _, r := range rules {
		mailID := r["value"]
		if mailID == nil {
			continue
		}
		mails, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailTemplate.Name, map[string]interface{}{
			utils.SpecialIDParam: mailID,
		}, false)
		if err != nil || len(mails) == 0 {
			continue
		}
		mail := mails[0]
		if tmplPath, ok := mail["redirect_on"]; ok { // with redirection only such as outlook
			if len(dest) > 0 {
				d := dest[0]
				path := utils.ToString(tmplPath)
				for k, v := range d {
					if strings.Contains(path, k) {
						path = strings.ReplaceAll(path, k, utils.ToString(v))
					}
				}
			}
			values, err := url.ParseQuery(utils.GetString(mail, "redirect_on"))
			if err == nil {
				SetRedirection(t.Domain.GetDomainID(), values.Encode())
			}
			return mailings
		}

		usfrom, err := t.Domain.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
			utils.SpecialIDParam: t.Domain.GetUserID(),
		}, false)
		if err != nil || len(usfrom) == 0 {
			continue
		}
		destOnResponse := int64(-1)
		if fromSchema.ID == utils.GetString(mail, ds.SchemaDBField+"_on_response") {
			destOnResponse = utils.GetInt(record, utils.SpecialIDParam)
		}
		signature := utils.GetString(mail, "signature")
		if len(toUsers) == 0 {
			if len(dest) > 0 {
				if m, err := ForgeMail(
					usfrom[0],
					utils.Record{}, // always keep a copy
					utils.GetString(mail, "subject"),
					utils.GetString(mail, "template"),
					t.getLinkLabel(toSchema, dest[0]),
					t.Domain,
					utils.GetInt(mail, utils.SpecialIDParam),
					toSchemaID,
					destID,
					destOnResponse,
					t.getFileAttached(toSchema, record),
					signature,
				); err == nil {
					mailings = append(mailings, m)
				}
			} else {
				if m, err := ForgeMail(
					usfrom[0],
					utils.Record{}, // always keep a copy
					utils.GetString(mail, "subject"),
					utils.GetString(mail, "template"),
					utils.Record{},
					t.Domain,
					utils.GetInt(mail, utils.SpecialIDParam),
					toSchemaID,
					destID,
					destOnResponse,
					"",
					signature,
				); err == nil {
					mailings = append(mailings, m)
				}
			}
		}
		for _, to := range toUsers {
			if len(dest) > 0 {
				if fmt.Sprintf("%v", toSchemaID) == utils.GetString(mail, ds.SchemaDBField+"_on_response") {
					destOnResponse = utils.GetInt(dest[0], utils.SpecialIDParam)
				}
				if m, err := ForgeMail(
					usfrom[0],
					to, // always keep a copy
					utils.GetString(mail, "subject"),
					utils.GetString(mail, "template"),
					t.getLinkLabel(toSchema, dest[0]),
					t.Domain,
					utils.GetInt(mail, utils.SpecialIDParam),
					toSchemaID,
					destID,
					destOnResponse,
					t.getFileAttached(toSchema, record),
					signature,
				); err == nil {
					mailings = append(mailings, m)
				}
			} else {
				if m, err := ForgeMail(
					usfrom[0],
					to, // always keep a copy
					utils.GetString(mail, "subject"),
					utils.GetString(mail, "template"),
					map[string]interface{}{},
					t.Domain,
					utils.GetInt(mail, utils.SpecialIDParam),
					-1,
					-1,
					destOnResponse,
					"",
					signature,
				); err == nil {
					mailings = append(mailings, m)
				}
			}
		}
	}
	return mailings
}

func (t *TriggerService) getFileAttached(toSchema sm.SchemaModel, record utils.Record) string {
	for k, v := range record {
		if f, err := toSchema.GetField(k); err == nil && strings.ToLower(f.Type) == "html" {
			return utils.ToString(v)
		}
	}
	return ""
}

func (t *TriggerService) getLinkLabel(toSchema sm.SchemaModel, record utils.Record) utils.Record {
	for _, field := range toSchema.Fields {
		if linkScheme, err := sm.GetSchemaByID(field.GetLink()); err == nil {
			// there is a link... soooo do something
			if res, err := t.Domain.GetDb().SelectQueryWithRestriction(linkScheme.Name, map[string]interface{}{
				utils.SpecialIDParam: record[field.Name],
			}, false); err == nil && len(res) > 0 {
				item := res[0]
				if utils.GetString(item, "label") != "" {
					record[field.Name] = utils.GetString(item, "label")
				}
				if utils.GetString(item, "name") != "" {
					record[field.Name] = utils.GetString(item, "name")
				}
			}
		}
	}
	return record
}

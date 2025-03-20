package view_convertor

import (
	"fmt"
	"slices"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"strings"
)

func (d *ViewConvertor) EnrichWithWorkFlowView(record utils.Record, tableName string, isWorkflow bool) *sm.WorkflowModel {
	if !isWorkflow {
		return nil
	}

	workflow := sm.WorkflowModel{Position: "0", Current: "0", Steps: make(map[string][]sm.WorkflowStepModel)}
	id, requestID, nexts := "", "", []string{}

	switch tableName {
	case ds.DBWorkflow.Name:
		id = record.GetString(utils.SpecialIDParam)
	case ds.DBRequest.Name:
		if t := d.FetchRecord(ds.DBRequest.Name, record.GetString(utils.SpecialIDParam)); len(t) > 0 {
			id = utils.ToString(t[0][ds.RootID(ds.DBWorkflow.Name)])
			requestID = utils.ToString(t[0][utils.SpecialIDParam])
			workflow = d.InitializeWorkflow(t[0])
		} else {
			return nil
		}
	case ds.DBTask.Name:
		if workflow, id, requestID, nexts = d.handleTaskWorkflow(record); id == "" {
			return nil
		}
	default:
		return nil
	}

	if id == "" || id == "<nil>" {
		return nil
	}

	return d.populateWorkflowSteps(&workflow, id, requestID, nexts)
}

func (d *ViewConvertor) InitializeWorkflow(record map[string]interface{}) sm.WorkflowModel {
	return sm.WorkflowModel{
		IsDismiss: record["state"] == "dismiss",
		Current:   utils.ToString(record["current_index"]),
		Position:  utils.ToString(record["current_index"]),
		IsClose:   record["state"] == "completed" || record["state"] == "dismiss" || record["state"] == "refused" || record["state"] == "canceled",
	}
}

func (d *ViewConvertor) handleTaskWorkflow(record utils.Record) (sm.WorkflowModel, string, string, []string) {
	var workflow sm.WorkflowModel
	taskRecord := d.FetchRecord(ds.DBTask.Name, record.GetString(utils.SpecialIDParam))
	if len(taskRecord) == 0 {
		return workflow, "", "", nil
	}

	reqRecord := d.FetchRecord(ds.DBRequest.Name, utils.ToString(taskRecord[0][ds.RootID(ds.DBRequest.Name)]))
	if len(reqRecord) > 0 {
		workflow = d.InitializeWorkflow(reqRecord[0])
	}

	if taskRecord[0][ds.RootID(ds.DBWorkflowSchema.Name)] != nil {
		nexts := d.ParseNextSteps(taskRecord[0])
		requestID := record.GetString(ds.RootID(ds.DBRequest.Name))
		workflow.CurrentDismiss = record["state"] == "dismiss"
		workflow.CurrentClose = record["state"] == "completed" || record["state"] == "dismiss" || record["state"] == "refused" || record["state"] == "canceled"

		schemaRecord := d.FetchRecord(ds.DBWorkflowSchema.Name, utils.ToString(taskRecord[0][ds.RootID(ds.DBWorkflowSchema.Name)]))
		if len(schemaRecord) > 0 {
			workflow.Current = utils.GetString(schemaRecord[0], "index")
			workflow.CurrentHub = utils.Compare(schemaRecord[0]["hub"], true)
			return workflow, utils.ToString(schemaRecord[0][ds.RootID(ds.DBWorkflow.Name)]), requestID, nexts
		}
	}
	return workflow, "", "", nil
}

func (d *ViewConvertor) ParseNextSteps(record map[string]interface{}) []string {
	if record["nexts"] == "all" || record["nexts"] == "" || record["nexts"] == nil {
		return nil
	}
	return strings.Split(utils.ToString(record["nexts"]), ",")
}

func (d *ViewConvertor) populateWorkflowSteps(workflow *sm.WorkflowModel, id, requestID string, nexts []string) *sm.WorkflowModel {
	steps := d.FetchRecord(ds.DBWorkflowSchema.Name, id)
	if len(steps) == 0 {
		return workflow
	}

	workflow.Steps = make(map[string][]sm.WorkflowStepModel)
	for _, step := range steps {
		index := utils.ToString(step["index"])
		newStep := sm.WorkflowStepModel{
			ID:        utils.GetInt(step, utils.SpecialIDParam),
			Name:      utils.ToString(step[sm.NAMEKEY]),
			Optionnal: utils.Compare(step["optionnal"], true),
			IsSet:     !utils.Compare(step["optionnal"], true) || slices.Contains(nexts, utils.ToString(step["wrapped_"+ds.RootID(ds.DBWorkflow.Name)])),
		}

		if workflow.Current != "" {
			d.populateTaskDetails(&newStep, step)
		}

		if wrapped, ok := step["wrapped_"+ds.RootID(ds.DBWorkflow.Name)]; ok {
			newStep.Workflow = d.EnrichWithWorkFlowView(utils.Record{utils.SpecialIDParam: wrapped}, ds.DBWorkflow.Name, true)
		}

		workflow.Steps[index] = append(workflow.Steps[index], newStep)
	}
	workflow.ID = id
	return workflow
}

func (d *ViewConvertor) populateTaskDetails(newStep *sm.WorkflowStepModel, step map[string]interface{}) {
	tasks := d.FetchRecord(ds.DBTask.Name, utils.ToString(step[utils.SpecialIDParam]))
	if len(tasks) > 0 {
		newStep.IsClose = utils.Compare(tasks[0]["is_close"], true)
		newStep.IsCurrent = utils.Compare(tasks[0]["state"], "pending")
		newStep.IsDismiss = utils.Compare(tasks[0]["is_dismiss"], "dismiss")
	}
}

func (d *ViewConvertor) BuildPath(tableName string, rows string, extra ...string) string {
	path := fmt.Sprintf("/%s/%s?rows=%v", utils.MAIN_PREFIX, tableName, rows)
	for _, ext := range extra {
		path += "&" + ext
	}
	return path
}

// TODO : Add hierarchical view in the workflow enrich

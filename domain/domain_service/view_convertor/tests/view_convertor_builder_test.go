package view_convertor_test

import (
	"sqldb-ws/domain/domain_service/view_convertor"
	"sqldb-ws/domain/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnrichWithWorkFlowView_NotWorkflow(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := utils.Record{}
	result := vc.EnrichWithWorkFlowView(record, "some_table", false)
	assert.Nil(t, result)
}

func TestEnrichWithWorkFlowView_EmptyID(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := utils.Record{utils.SpecialIDParam: ""}
	result := vc.EnrichWithWorkFlowView(record, "workflow_table", true)
	assert.Nil(t, result)
}

func TestEnrichWithWorkFlowView_WorkflowInitialization(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := utils.Record{utils.SpecialIDParam: "123"}
	result := vc.EnrichWithWorkFlowView(record, "workflow_table", true)
	assert.NotNil(t, result)
	assert.Equal(t, "0", result.Position)
	assert.Equal(t, "0", result.Current)
}

func TestEnrichWithWorkFlowView_HandlesRequest(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := utils.Record{utils.SpecialIDParam: "456"}
	result := vc.EnrichWithWorkFlowView(record, "request_table", true)
	assert.Nil(t, result) // Simulating an empty fetch record scenario
}

func TestEnrichWithWorkFlowView_HandlesTask(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := utils.Record{utils.SpecialIDParam: "789"}
	result := vc.EnrichWithWorkFlowView(record, "task_table", true)
	assert.Nil(t, result) // Simulating an empty fetch record scenario
}

func TestInitializeWorkflow(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := map[string]interface{}{"state": "completed", "current_index": "2"}
	workflow := vc.InitializeWorkflow(record)
	assert.True(t, workflow.IsClose)
	assert.Equal(t, "2", workflow.Position)
}

func TestParseNextSteps_Empty(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := map[string]interface{}{"nexts": ""}
	result := vc.ParseNextSteps(record)
	assert.Nil(t, result)
}

func TestParseNextSteps_All(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := map[string]interface{}{"nexts": "all"}
	result := vc.ParseNextSteps(record)
	assert.Nil(t, result)
}

func TestParseNextSteps_ValidList(t *testing.T) {
	vc := &view_convertor.ViewConvertor{}
	record := map[string]interface{}{"nexts": "step1,step2,step3"}
	result := vc.ParseNextSteps(record)
	assert.Equal(t, []string{"step1", "step2", "step3"}, result)
}

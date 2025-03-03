package models

import (
	"encoding/json"
	"errors"
	"sqldb-ws/domain/utils"
)

var SchemaRegistry = map[string]SchemaModel{}

type SchemaModel struct { // lightest definition a db table
	ID       int64        `json:"id"`
	Name     string       `json:"name"`
	Label    string       `json:"label"`
	Category string       `json:"category"`
	Fields   []FieldModel `json:"fields"`
}

func (t SchemaModel) Deserialize(rec utils.Record) SchemaModel {
	b, _ := json.Marshal(rec)
	json.Unmarshal(b, &t)
	return t
}
func (t SchemaModel) GetName() string { return t.Name }

func (t SchemaModel) HasField(name string) bool {
	CacheMutex.Lock()
	for _, field := range t.Fields {
		if field.Name == name {
			CacheMutex.Unlock()
			return true
		}
	}
	CacheMutex.Unlock()
	return false
}

func GetSchemaByID(id int64) (SchemaModel, error) {
	CacheMutex.Lock()
	for _, schema := range SchemaRegistry {
		if schema.ID == id {
			CacheMutex.Unlock()
			return schema, nil
		}
	}
	CacheMutex.Unlock()
	return SchemaModel{}, errors.New("no schema corresponding to reference id")
}

func (t SchemaModel) GetTypeAndLinkForField(name string) (string, string, error) {
	field, err := t.GetField(name)
	if err != nil {
		return "", "", err
	}
	foreign, err := GetSchemaByID(field.Link)
	if err != nil {
		return "", "", err
	}
	return field.Type, foreign.Name, nil
}
func (t SchemaModel) GetField(name string) (FieldModel, error) {
	for _, field := range t.Fields {
		if field.Name == name {
			return field, nil
		}
	}
	return FieldModel{}, errors.New("no field corresponding to reference")
}
func (t SchemaModel) GetFieldByID(id int64) (FieldModel, error) {
	for _, field := range t.Fields {
		if field.ID == id {
			return field, nil
		}
	}
	return FieldModel{}, errors.New("no field corresponding to reference")
}

func (v SchemaModel) ToRecord() utils.Record {
	var r utils.Record
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &r)
	return r
}

type FieldModel struct { // definition a db table columns
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Desc         string      `json:"description"`
	Type         string      `json:"type"`
	Index        int64       `json:"index"`
	Placeholder  string      `json:"placeholder"`
	Default      interface{} `json:"default_value"`
	Level        string      `json:"read_level"`
	Readonly     bool        `json:"readonly"`
	Link         int64       `json:"link_id"`
	ForeignTable string      `json:"foreign_table"` // Special case for foreign key
	Constraint   string      `json:"constraints"`   // Special case for constraint on field
	Required     bool        `json:"required"`
}

func (v FieldModel) ToRecord() utils.Record {
	var r utils.Record
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &r)
	return r
}

type ViewModel struct { // lightest struct based on SchemaModel dedicate to view
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Label       string            `json:"label"`
	SchemaID    int64             `json:"schema_id"`
	SchemaName  string            `json:"schema_name"`
	Description string            `json:"description"`
	Path        string            `json:"link_path"`
	Order       []string          `json:"order"`
	Schema      utils.Record      `json:"schema"`
	Items       []ViewItemModel   `json:"items"`
	Actions     []string          `json:"actions"`
	ActionPath  string            `json:"action_path"`
	Readonly    bool              `json:"readonly"`
	Workflow    *WorkflowModel    `json:"workflow"`
	IsWrapper   bool              `json:"is_wrapper"`
	Shortcuts   map[string]string `json:"shortcuts"`
}

func (v ViewModel) ToRecord() utils.Record {
	var r utils.Record
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &r)
	return r
}

type ViewItemModel struct {
	IsEmpty       bool                     `json:"-"`
	Sort          int64                    `json:"-"`
	Path          string                   `json:"link_path"`
	DataPaths     string                   `json:"data_path"`
	ValuePathMany map[string]string        `json:"values_path_many"`
	Values        map[string]interface{}   `json:"values"`
	ValueShallow  map[string]interface{}   `json:"values_shallow"`
	ValueMany     map[string]utils.Results `json:"values_many"`
	HistoryPath   string                   `json:"history_path"`
	Workflow      *WorkflowModel           `json:"workflow"`
	Readonly      bool                     `json:"readonly"`
}

type ViewFieldModel struct { // lightest struct based on FieldModel dedicate to view
	Label        string                 `json:"label" validate:"required"`
	Type         string                 `json:"type" validate:"required"`
	Index        int64                  `json:"index"`
	Description  string                 `json:"description"`
	Placeholder  string                 `json:"placeholder"`
	Default      interface{}            `json:"default_value"`
	Required     bool                   `json:"required"`
	Readonly     bool                   `json:"readonly"`
	LinkPath     string                 `json:"values_path"`
	ActionPath   string                 `json:"action_path"`
	Actions      []string               `json:"actions"`
	DataSchemaID int64                  `json:"data_schema_id"`
	DataSchema   map[string]interface{} `json:"data_schema"`
}

type WorkflowModel struct { // lightest struct based on SchemaModel dedicate to view
	ID             string                         `json:"id"`
	IsDismiss      bool                           `json:"is_dismiss"`
	Current        string                         `json:"current"`
	Position       string                         `json:"position"`
	IsClose        bool                           `json:"is_close"`
	CurrentHub     bool                           `json:"current_hub"`
	CurrentDismiss bool                           `json:"current_dismiss"`
	CurrentClose   bool                           `json:"current_close"`
	Steps          map[string][]WorkflowStepModel `json:"steps"`
}

type WorkflowStepModel struct { // lightest struct based on SchemaModel dedicate to view
	ID        int64          `json:"id"`
	Name      string         `json:"name"`
	Optionnal bool           `json:"optionnal"`
	IsSet     bool           `json:"is_set"`
	IsDismiss bool           `json:"is_dismiss"`
	IsCurrent bool           `json:"is_current"`
	IsClose   bool           `json:"is_close"`
	Workflow  *WorkflowModel `json:"workflow"`
}

type FilterModel struct {
	ID        int64       `json:"id"`
	Name      string      `json:"name"`
	Label     string      `json:"label"`
	Type      string      `json:"type"`
	Index     float64     `json:"index"`
	Value     interface{} `json:"value"`
	Operator  string      `json:"operator"`
	Separator string      `json:"separator"`
	Dir       string      `json:"dir"`
	Width     float64     `json:"width"`
}

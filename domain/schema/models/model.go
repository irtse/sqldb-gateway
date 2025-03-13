package models

import (
	"encoding/json"
	"errors"
	"sqldb-ws/domain/utils"
	"strconv"
	"strings"
)

var SchemaRegistry = map[string]SchemaModel{}

type SchemaModel struct { // lightest definition a db table
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Label    string       `json:"label"`
	Category string       `json:"category"`
	Fields   []FieldModel `json:"fields,omitempty"`
}

func (t SchemaModel) GetID() int64 {
	i, err := strconv.Atoi(t.ID)
	if err != nil {
		return -1
	}
	return int64(i)
}

func (t SchemaModel) Deserialize(rec utils.Record) SchemaModel {
	b, _ := json.Marshal(rec)
	json.Unmarshal(b, &t)
	return t
}
func (t SchemaModel) GetName() string { return t.Name }

func (t SchemaModel) SetField(field map[string]interface{}) SchemaModel {
	newField := FieldModel{}.Map(field)
	if !t.HasField(newField.Name) {
		CacheMutex.Lock()
		defer CacheMutex.Unlock()
		t.Fields = append(t.Fields, *newField)
	} else {
		CacheMutex.Lock()
		defer CacheMutex.Unlock()
		for _, f := range t.Fields {
			if newField.Name != f.Name {
				f = *newField
			}
		}
	}
	SchemaRegistry[t.Name] = t
	return t
}

func (t SchemaModel) HasField(name string) bool {
	CacheMutex.Lock()
	defer CacheMutex.Unlock()
	for _, field := range t.Fields {
		if field.Name == name {
			return true
		}
	}
	return false
}

func GetSchemaByID(id int64) (SchemaModel, error) {
	CacheMutex.Lock()
	for _, schema := range SchemaRegistry {
		if schema.GetID() == id {
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
	foreign, err := GetSchemaByID(field.GetLink())
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
		if field.GetID() == id {
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

func (v SchemaModel) ToSchemaRecord() utils.Record {
	fields := []FieldModel{}
	for _, field := range v.Fields {
		if !strings.Contains(field.Type, "many") {
			fields = append(fields, field)
		}
	}
	var r utils.Record
	b, _ := json.Marshal(SchemaModel{
		ID:       v.ID,
		Name:     v.Name,
		Label:    v.Label,
		Category: v.Category,
		Fields:   fields,
	})
	json.Unmarshal(b, &r)
	return r
}

type FieldModel struct { // definition a db table columns
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Desc         string      `json:"description"`
	Type         string      `json:"type"`
	Index        int64       `json:"index"`
	Placeholder  string      `json:"placeholder"`
	Default      interface{} `json:"default_value"`
	Level        string      `json:"read_level,omitempty"`
	Readonly     bool        `json:"readonly"`
	Link         string      `json:"link_id"`
	ForeignTable string      `json:"-"`           // Special case for foreign key
	Constraint   string      `json:"constraints"` // Special case for constraint on field
	Required     bool        `json:"required"`
}

func (t FieldModel) Map(m map[string]interface{}) *FieldModel {
	return &FieldModel{
		ID:          utils.ToString(m["id"]),
		Name:        utils.ToString(m["name"]),
		Label:       utils.ToString(m["label"]),
		Desc:        utils.ToString(m["description"]),
		Type:        utils.ToString(m["type"]),
		Index:       utils.ToInt64(m["index"]),
		Placeholder: utils.ToString(m["placeholder"]),
		Default:     m["default_value"],
		Level:       utils.ToString(m["read_level"]),
		Readonly:    utils.Compare(m["readonly"], true),
		Link:        utils.ToString(m["link_id"]),
		Constraint:  utils.ToString(m["constraints"]),
		Required:    utils.Compare(m["required"], true),
	}
}

func (t FieldModel) GetID() int64 {
	i, err := strconv.Atoi(t.ID)
	if err != nil {
		return -1
	}
	return int64(i)
}

func (t FieldModel) GetLink() int64 {
	i, err := strconv.Atoi(t.Link)
	if err != nil {
		return -1
	}
	return int64(i)
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

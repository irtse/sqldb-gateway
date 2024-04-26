package schema

import (
	"fmt"
	"sync"
	"errors"
	"strings"
	"encoding/json"
	"sqldb-ws/lib/domain/utils"
	conn "sqldb-ws/lib/infrastructure/connector"
)

var schemaRegistry = map[string]SchemaModel{}

func IsRootDB(name string) bool { 
	if len(name) > 1 { return strings.Contains(name[:2], "db") 
    } else { return false }
}
func RootID(name string) string { 
	if IsRootDB(name) { return name + "_id" } else { return RootName(name) + "_id" }
}
func RootName(name string) string { return "db" + name }

func GetSchemaByFieldID(id int64) (SchemaModel, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	for _, t := range schemaRegistry { 
		for _, field := range t.Fields {
			if field.ID == id { return t, nil }
		}
	}
	return SchemaModel{}, errors.New("no field corresponding to reference")
}

func GetFieldByID(id int64) (FieldModel, error) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	for _, t := range schemaRegistry { 
		for _, field := range t.Fields {
			if field.ID == id { return field, nil }
		}
	}
	return FieldModel{}, errors.New("no field corresponding to reference")
}

type SchemaModel struct { // lightest definition a db table 
	ID 		int64               		 `json:"id"`
	Name    string              		 `json:"name"`
	Label   string						 `json:"label"`
	Fields  []FieldModel 		 		 `json:"fields"`
}
func (t SchemaModel) Deserialize(rec utils.Record) SchemaModel { 
	b, _ := json.Marshal(rec) 
	json.Unmarshal(b, &t)
	return t
}
func (t SchemaModel) GetName() string { return t.Name }

func (t SchemaModel) HasField(name string) bool {
	cacheMutex.Lock()
	for _, field := range t.Fields {
		if field.Name == name { 
			cacheMutex.Unlock()
			return true 
		}
	}
	cacheMutex.Unlock()
	return false
}
func (t SchemaModel) GetField(name string) (FieldModel, error) {
	for _, field := range t.Fields {
		if field.Name == name { return field, nil }
	}
	return FieldModel{}, errors.New("no field corresponding to reference")
}

type FieldModel struct { // definition a db table columns
	ID int64         	`json:"id"`
	Name string         `json:"name"`
	Label string        `json:"label"`
	Desc  string        `json:"description"`
	Type string         `json:"type"`
	Index int64         `json:"index"`
	Placeholder string  `json:"placeholder"`
	Default interface{} `json:"default_value"`
	Level 	string 		`json:"read_level"`
	Readonly bool		`json:"readonly"`
	Link int64 			`json:"link_id"`
	ForeignTable string `json:"foreign_table"` // Special case for foreign key
	Constraint string   `json:"constraints"` // Special case for constraint on field
	Required bool       `json:"required"`
}
var cacheMutex  = sync.Mutex{}
func LoadCache(name string, db *conn.Db) {
	db.ClearFilter()
	if name != utils.ReservedParam { db.SQLRestriction = "name=" + conn.Quote(name) }// filter out system tables
	schemas, err := db.SelectResults(DBSchema.Name) // load schemas from base 
	if err != nil || len(schemas) == 0 { return }  // anything to do if empty
	db.ClearFilter()
	for _, schema := range schemas { // on each schema deserialize + get fields + add to cache
		var newSchema SchemaModel
		data, _:= json.Marshal(schema) // deserialize schema
		json.Unmarshal(data, &newSchema) // deserialize schema
		newSchema.Fields = []FieldModel{} // init fields
		db.SQLRestriction = RootID(DBSchema.Name) + "=" + fmt.Sprintf("%v", newSchema.ID) // filter expected fields
		fields, err := db.SelectResults(DBSchemaField.Name)	// get fields
		db.SQLRestriction = "" // reset restriction
		if err == nil && len(fields) > 0 {
			for _, field := range fields { // on each field deserialize + add to schema fields
				var newField FieldModel 
				data, _:= json.Marshal(field) // deserialize field
				json.Unmarshal(data, &newField) // deserialize field
				newSchema.Fields = append(newSchema.Fields, newField) // add field to schema
			}
		} // anything to do if empty
		cacheMutex.Lock()
		schemaRegistry[newSchema.Name] = newSchema // add schema to cache
		cacheMutex.Unlock()
	}
}

func HasSchema(tableName string) bool { 
	cacheMutex.Lock()
	if _, ok := schemaRegistry[tableName]; !ok {  
		cacheMutex.Unlock()
		return false 
	} else { 
		cacheMutex.Unlock()	
		return true
	}
}

func HasField(tableName string, name string) bool {
	if schema, ok := schemaRegistry[tableName]; !ok { return false } else { 
		return schema.HasField(name) 
	}
	return false
}


func GetSchema(tableName string) (SchemaModel, error) { 
	cacheMutex.Lock()
	if schema, ok := schemaRegistry[tableName]; !ok { 
		cacheMutex.Unlock()
		return SchemaModel{}, errors.New("no schema corresponding to reference name")
	 } else { 
		cacheMutex.Unlock()
		return schema, nil 
	}
}

func GetSchemaByID(id int64) (SchemaModel, error) {
	cacheMutex.Lock()
	for _, schema := range schemaRegistry {
		if schema.ID == id { 
			cacheMutex.Unlock()
			return schema, nil 
		}
	}
	cacheMutex.Unlock()
	return SchemaModel{}, errors.New("no schema corresponding to reference id")
}
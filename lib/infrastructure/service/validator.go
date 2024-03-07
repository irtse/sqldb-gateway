package service

import (
	"errors"
	"encoding/json"
	tool "sqldb-ws/lib"
	"github.com/go-playground/validator"
	"sqldb-ws/lib/entities"
)

type validable interface { entities.TableEntity | entities.ShallowTableEntity | entities.TableColumnEntity | map[string]interface{} }
type DBValidator[T validable] struct { data T }

func Validator[T validable]() *DBValidator[T] {
    c := DBValidator[T]{ }
    return &c
}

func (v *DBValidator[T]) ValidateStruct(data map[string]interface{}) (*T, error) {
	jsonData, err := json.Marshal(data)
	if err != nil { return nil, err }
	json.Unmarshal(jsonData, &v.data)
	validate := validator.New()
	if err := validate.Struct(v.data); err != nil { return nil, err }
	return &v.data, nil
}

func (v *DBValidator[T]) ValidateSchema(data map[string]interface{}, t *TableInfo, reverse bool) (map[string]interface{}, error) {
	schema, err := t.schema(t.Name)
	if err != nil { return nil,  err }
	newData := map[string]interface{}{}
	if reverse {
		found := false
		for _, s := range schema {
			for k, _ := range s.AssColumns {
				if v, ok := data[k]; ok { 
					newData[k]=v
					found = true 
				}
			}
		}
		if !found { return nil, errors.New("Not found field") }
		return newData, nil
	}
	for _, s := range schema {
		for k, v := range s.AssColumns {
			if !v.Null && v.Default == nil {
				if _, ok := data[k]; ok == false && k != tool.SpecialIDParam {
					if v.Label != "" {
						return nil, errors.New("Missing a required field " + v.Label + " (can't see it ? you probably missing permissions)")
					} else {
						return nil, errors.New("Missing a required field " + k + " (can't see it ? you probably missing permissions)")
					}
				} 
			}
			if v, ok := data[k]; ok { newData[k]=v }
		}
	}
	return newData, nil
}
package schema

import (
	"strconv"
	"encoding/json"
	"sqldb-ws/lib/domain/utils"
)

type ViewModel struct { // lightest struct based on SchemaModel dedicate to view
	ID 		 		int64               		`json:"id"`
	Name  		 	string 						`json:"name"`
	Label  		 	string 						`json:"label"`
	SchemaID		int64						`json:"schema_id"`
	SchemaName   	string 						`json:"schema_name"`
	Description  	string 						`json:"description"`
	Path		 	string 						`json:"link_path"`
	Order		 	[]string 					`json:"order"`
	Schema		 	utils.Record 				`json:"schema"`
	Items		 	[]ViewItemModel				`json:"items"`
	Actions		 	[]string				 	`json:"actions"`
	ActionPath		string						`json:"action_path"`
	Readonly		bool						`json:"readonly"`
	Workflow		*WorkflowModel				`json:"workflow"`
	Shortcuts		map[string]string			`json:"shortcuts"`
}

func (v ViewModel) ToRecord() utils.Record { 
	var r utils.Record
	b, _ := json.Marshal(v)
	json.Unmarshal(b, &r)
	return r
}

type ViewItemModel struct {
	IsEmpty		   bool							`json:"-"`
	Sort		   int64						`json:"-"`
	Path 	   	   string					 	`json:"link_path"`
	DataPaths  	   string				        `json:"data_path"`
	ValuePathMany  map[string]string			`json:"values_path_many"`
	Values 	   	   map[string]interface{} 	    `json:"values"`
	ValueShallow   map[string]interface{}		`json:"values_shallow"`
	ValueMany      map[string]utils.Results		`json:"values_many"`
	HistoryPath	   string						`json:"history_path"`
	Workflow  	   *WorkflowModel				`json:"workflow"`
	Readonly	   bool							`json:"readonly"`
}

type ViewFieldModel struct { // lightest struct based on FieldModel dedicate to view
	Label 					string 						 `json:"label" validate:"required"`
	Type 					string 		 				 `json:"type" validate:"required"`
	Index					int64						 `json:"index"`
	Description 			string				 		 `json:"description"`
	Placeholder 			string				 		 `json:"placeholder"`
	Default 				interface{}					 `json:"default_value"`
	Required 				bool 		 			 	 `json:"required"`
	Readonly 				bool 		 			 	 `json:"readonly"`
	LinkPath 				string 		 			 	 `json:"values_path"`
	ActionPath		 		string 		 		 	 	 `json:"action_path"`
	Actions		 			[]string 		 		 	 `json:"actions"`
	DataSchemaID 			int64 		 			 	 `json:"data_schema_id"`
	DataSchema 				map[string]interface{} 		 `json:"data_schema"`
}

func GetTablename(supposedTableName string) (string) {
	i, err := strconv.Atoi(supposedTableName)
	if err != nil { return supposedTableName }
	tablename, err := GetSchemaByID(int64(i))
	if err != nil { return "" }
	return tablename.Name
}

type WorkflowModel struct { // lightest struct based on SchemaModel dedicate to view
	ID 					string               			`json:"id"`
	IsDismiss 			bool               				`json:"is_dismiss"`
	Current  			string 							`json:"current"`
	Position  			string 							`json:"position"`
	IsClose  			bool 							`json:"is_close"`
	CurrentHub  		bool 							`json:"current_hub"`
	CurrentDismiss  	bool 							`json:"current_dismiss"`
	CurrentClose	 	bool 							`json:"current_close"`
	Steps				map[string][]WorkflowStepModel 	`json:"steps"`
}

type WorkflowStepModel struct { // lightest struct based on SchemaModel dedicate to view
	ID 					int64               		`json:"id"`
	Name  				string 						`json:"name"`
	Optionnal	  		bool 						`json:"optionnal"`
	IsSet 				bool               			`json:"is_set"`
	IsDismiss 			bool               			`json:"is_dismiss"`
	IsCurrent 			bool               			`json:"is_current"`
	IsClose  			bool 						`json:"is_close"`
	Workflow  			*WorkflowModel 				`json:"workflow"`
}

type FilterModel struct { 
	ID 					int64               		`json:"id"`
	Name  				string 						`json:"name"`
	Label				string						`json:"label"`
	Type				string						`json:"type"`
	Index				float64						`json:"index"`
	Value 				interface{} 				`json:"value"`
	Operator 			string 						`json:"operator"`
	Separator 			string 						`json:"separator"`
	Dir 				string 						`json:"dir"`
	Width 				float64						`json:"width"`
}
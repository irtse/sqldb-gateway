package entities

import ("strings")

/*
	Here are common resources for every level.
	Entity are usefull struct to organize our Database works.
*/

func IsRootDB(name string) bool { 
	if len(name) > 1 { return strings.Contains(name[:2], "db") 
    } else { return false }
}
func RootID(name string) string { 
	if IsRootDB(name) { return name + "_id" 
    } else { return RootName(name) + "_id" }
}
func RootName(name string) string { return "db" + name }

type TableEntity struct { // define a db table 
	Name    string              		 `json:"name" validate:"required"`
	Label   string						 `json:"label"`
	Columns []TableColumnEntity 		 `json:"columns"`
}

func (t TableEntity) GetName() string { return t.Name }

type ShallowTableEntity struct { // lightest definition a db table 
	Name    string              		 `json:"name"`
	Label   string						 `json:"label"`
	Columns []TableColumnEntity 		 `json:"columns"`
}
func (t ShallowTableEntity) GetName() string { return t.Name }


type TableColumnEntity struct { // definition a db table columns
	Name string         `json:"name" validate:"required"`
	Label string         `json:"label"`
	Type string         `json:"type"`
	Index int64         `json:"-"`
	Default interface{} `json:"default_value"`
	Level 	string 		`json:"read_level"`
	ForeignTable string `json:"link"`
	Readonly bool		`json:"readonly"`
	Constraint string   `json:"constraint"`
	Null bool           `json:"nullable"`
	Comment string      `json:"comment"`
	NewName string      `json:"-"`
}

func (t TableColumnEntity) GetName() string { return t.Name }

type SchemaColumnEntity struct { // struct based on dbschema_columns in root
	ID 						int64 						 `json:"id"`
	Label 					string 						 `json:"label" validate:"required"`
	Name    				string          			 `json:"name" validate:"required"`
	Type 					string 		 				 `json:"type" validate:"required"`
	Index					int64						 `json:"index"`
	Description 			string				 		 `json:"description"`
	Placeholder 			string				 		 `json:"placeholder"`
	Default 				interface{}					 `json:"default_value"`
	Level 					string 		 			 	 `json:"read_level"`
	Required 				bool 		 			 	 `json:"required"`
	Readonly 				bool 		 			 	 `json:"readonly"`
	SchemaId 				int64 		 			 	 `json:"dbschema_id"`
	Link 					string 		 			 	 `json:"link"`
	LinkDir					string 		 			 	 `json:"link_sql_dir"`
	LinkOrder 	 			string 		 			 	 `json:"link_sql_order"`
	LinkView				string 		 			 	 `json:"link_sql_view"`
}
func (t SchemaColumnEntity) GetName() string { return t.Name }

type ShallowSchemaColumnEntity struct {// lightest struct based on dbschema_columns in root
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
	DataSchemaOrder 		[]string 		 		 	 `json:"data_schema_order"`
	DataSchema 				map[string]interface{} 		 `json:"data_schema"`
}
func (t ShallowSchemaColumnEntity) GetName() string { return t.Label }
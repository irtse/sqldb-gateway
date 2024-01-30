package entities

import ("strings")

func IsRootDB(name string) bool { 
	if len(name) > 1 { return strings.Contains(name[:2], "db") 
    } else { return false }
}
func RootID(name string) string { 
	if IsRootDB(name) { return name + "_id" 
    } else { return RootName(name) + "_id" }
}
func RootName(name string) string { return "db" + name }

type TableEntity struct {
	Name    string              		 `json:"name" validate:"required"`
	Columns []TableColumnEntity 		 `json:"columns"`
}

func (t TableEntity) GetName() string { return t.Name }

type ShallowTableEntity struct {
	Name    string              		 `json:"name"`
	Columns []TableColumnEntity 		 `json:"columns"`
}
func (t ShallowTableEntity) GetName() string { return t.Name }


type TableColumnEntity struct {
	Name string         `json:"name" validate:"required"`
	Type string         `json:"type"`
	Index int64         `json:"-"`
	Default interface{} `json:"default_value"`
	Hidden bool         `json:"hidden"`
	ForeignTable string `json:"link"`
	Constraint string   `json:"constraint"`
	Null bool           `json:"nullable"`
	Comment string      `json:"comment"`
	NewName string      `json:"-"`
}

func (t TableColumnEntity) GetName() string { return t.Name }

type TableUpdateEntity struct {
	Name string `json:"name" validate:"required"`
}

func (t TableUpdateEntity) GetName() string { return t.Name }

type LinkEntity struct {
	From    				int64          			 `json:"from_id" validate:"required"`
	To 						int64 		 			 `json:"to_id" validate:"required"`
	Columns 			    map[string]string 	 `json:"columns"`
	Anchor 					string 		 			 `json:"anchor"`
}

func (t LinkEntity) GetName() string { return t.Anchor }

type SchemaColumnEntity struct {
	Label 					string 						 `json:"label" validate:"required"`
	Name    				string          			 `json:"name" validate:"required"`
	Type 					string 		 				 `json:"type" validate:"required"`
	Index					int64						 `json:"index"`
	Description 			string				 		 `json:"description"`
	Placeholder 			string				 		 `json:"placeholder"`
	Default 				interface{}					 `json:"default_value"`
	Required 				bool 		 			 	 `json:"required"`
	Readonly 				bool 		 			 	 `json:"readonly"`
	SchemaId 				int64 		 			 	 `json:"dbschema_id"`
	Link 					string 		 			 	 `json:"link"`
	LinkDir					string 		 			 	 `json:"link_sql_dir"`
	LinkOrder 	 			string 		 			 	 `json:"link_sql_order"`
	LinkColumns 			string 		 			 	 `json:"link_sql_columns"`
	LinkRestriction 		string 		 			 	 `json:"link_sql_restriction"`
	LinkPath 				string 		 			 	 `json:"link_path"`
}
func (t SchemaColumnEntity) GetName() string { return t.Name }

type ShallowSchemaColumnEntity struct {
	Label 					string 						 `json:"label" validate:"required"`
	Type 					string 		 				 `json:"type" validate:"required"`
	Index					int64						 `json:"index"`
	Description 			string				 		 `json:"description"`
	Placeholder 			string				 		 `json:"placeholder"`
	Default 				interface{}					 `json:"default_value"`
	Required 				bool 		 			 	 `json:"required"`
	Readonly 				bool 		 			 	 `json:"readonly"`
	LinkPath 				string 		 			 	 `json:"link_path"`
}
func (t ShallowSchemaColumnEntity) GetName() string { return t.Label }
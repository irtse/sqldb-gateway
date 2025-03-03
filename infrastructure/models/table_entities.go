package models

type TableEntity struct {
	Name       string                       `json:"name"`
	AssColumns map[string]TableColumnEntity `json:"columns"`
	Cols       []string                     `json:"-"`
}

type TableColumnEntity struct { // definition a db table columns
	Name         string      `json:"name" validate:"required"`
	Label        string      `json:"label"`
	Type         string      `json:"type"`
	Index        int64       `json:"-"`
	Default      interface{} `json:"default_value"`
	Level        string      `json:"read_level"`
	ForeignTable string      `json:"link"`
	Readonly     bool        `json:"readonly"`
	Constraint   string      `json:"constraints"`
	Null         bool        `json:"nullable"`
	Comment      string      `json:"comment"`
	NewName      string      `json:"-"`
}

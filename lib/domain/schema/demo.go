package schema

var ConfidentialityLevel = SchemaModel{
	Name : "confidentiality_level",
	Label : "confidentiality level",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}
var confDatas = []string{"public", "confidential project", "IRT confidential", "restricted diffusion", "special restricted diffusion", "classified datas", "authorized profile"}

var FormalizedDataProject = SchemaModel{
	Name : "formalized_data_project",
	Label : "formalized data project",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: RootID(FormalizedData.Name), Type: INTEGER.String(), ForeignTable: FormalizedData.Name, Required : true, Index: 0, Label: "binded formalized data" },
		FieldModel{ Name: RootID(Project.Name), Type: INTEGER.String(), ForeignTable: Project.Name, Required : true, Index: 1, Label: "binded project" },
	},
}

var FormalizedDataStorageType = SchemaModel{ 
	Name : "formalized_data_storage_type",
	Label : "formalized data storage type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: RootID(FormalizedData.Name), Type: INTEGER.String(), ForeignTable: FormalizedData.Name, Required : true, Index: 0, Label: "binded formalized data" },
		FieldModel{ Name: RootID(SupportType.Name), Type: INTEGER.String(), ForeignTable: SupportType.Name, Required : true, Index: 1, Label: "binded storage type" },
	},
}

var Project = SchemaModel{ // todo
	Name : "project",
	Label : "project",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
		FieldModel{ Name: "code", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 1 },
		FieldModel{ Name: RootID(DBEntity.Name), Type: INTEGER.String(), ForeignTable: DBEntity.Name, Required : true, Index: 2, Label: "related entity" },
	},
}

var Protection = SchemaModel{ // TODO
	Name : "protection",
	Label : "protection",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var protArDatas = []string{"france", "UE", "out of the UE"}

var ProtectionArea = SchemaModel{ 
	Name : "protection_area",
	Label : "protection area",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}
var protTypDatas = []string{"SOLEAU envelope", "timestamp", "patent application filed", "patent granted", "APP registration certificate", "no protection"}

var ProtectionType = SchemaModel{ 
	Name : "protection_type",
	Label : "protection type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var RestrictionType = SchemaModel{
	Name : "restriction_type",
	Label : "restriction type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
		FieldModel{ Name: "formalized_data_id", Type: INTEGER.String(), ForeignTable: "formalized_data", Required : true, Index: 1, Label: "related formalized data" },
	},
}

var resFamDatas = []string{"report", "internal technical note", "laboratory workbook", "career guide", "manuscript article", "thesis manuscript", "innovation filing sheet"}

var ResultFamily = SchemaModel{ 
	Name : "result_family",
	Label : "result family",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var resTypDatas = []string{"process", "method", "algorithm", "software", "specification", "conception", "test results", "sample/test vehicle", "demonstrator/proof of concept", "database", "standard", "innovation idea", "irt test bench"}

var ResultType = SchemaModel{ 
	Name : "result_type",
	Label : "result type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var supTypDatas = []string{"IRT server", "external data center", "hard disk", "USB", "paper storager", "physical conservation"}

var SupportType = SchemaModel{ 
	Name : "support_type",
	Label : "support type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var supDatas = []string{"paper", "digital"}

var Support = SchemaModel{ 
	Name : "support",
	Label : "support",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var Valuation = SchemaModel{ // TODO
	Name : "valuation",
	Label : "valuation",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
		FieldModel{ Name: "formalized_data_id", Type: INTEGER.String(), ForeignTable: "formalized_data", Required : true, Index: 1, Label: "related formalized data" },
		FieldModel{ Name: RootID(ValuationType.Name), Type: INTEGER.String(), ForeignTable: ValuationType.Name, Required : true, Index: 2, Label: "related valuation type" },
		FieldModel{ Name: RootID(ValuationFormat.Name), Type: INTEGER.String(), ForeignTable: ValuationFormat.Name, Required : true, Index: 3, Label: "related valuation format" },

	},
}

var valFormDatas = []string{"scientific journal article", "conference presentation", "published thesis dissertation", "training", "BIP contribution to a new project", "license (patent/software)", "service provision"}

var ValuationFormat = SchemaModel{
	Name : "valuation_format",
	Label : "valuation format",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var valTyp = []string{"scientific", "economic"}

var ValuationType = SchemaModel{
	Name : "valuation_type",
	Label : "valuation type",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
	},
}

var FormalizedData = SchemaModel{
	Name : "formalized_data",
	Label : "formalized data",
	Category : "data",
	Fields : []FieldModel{
		FieldModel{ Name: "name", Type: VARCHAR.String(), Constraint: "unique", Required : true, Readonly : true, Index : 0 },
		FieldModel{ Name: "ref", Type: VARCHAR.String(), Required : false, Label: "referencing", Readonly : true, Index : 1 },
		FieldModel{ Name: "first_evaluation", Type: DECIMAL.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 2 },
		FieldModel{ Name: "first_evaluation_date", Type: TIMESTAMP.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 3 },
		FieldModel{ Name: "actualized_evaluation", Type: DECIMAL.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 4 },
		FieldModel{ Name: "capitalization_date", Type: TIMESTAMP.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 5 },
		FieldModel{ Name: "committee_date", Type: TIMESTAMP.String(), Required : true, Level: LEVELRESPONSIBLE, Index: 6, Label: "committee reunion date" },
		FieldModel{ Name: "result_type_id", Type: INTEGER.String(), Required : false, ForeignTable: ResultType.Name, Index: 7,  Label: "result type", },
		FieldModel{ Name: "result_family_id", Type: INTEGER.String(), Required : false, ForeignTable: ResultFamily.Name, Index: 8, Label: "result family", },
		FieldModel{ Name: "support_id", Type: INTEGER.String(), Required : false, ForeignTable: Support.Name, Index: 9, Label: "support" },
		FieldModel{ Name: "confidentiality_level_id", Type: INTEGER.String(), Required : false, Label: "confidentiality level", Level: LEVELRESPONSIBLE, ForeignTable: ConfidentialityLevel.Name, Index: 10 },
		FieldModel{ Name: "contractual", Type: BOOLEAN.String(), Required : false, Level: LEVELRESPONSIBLE, Index: 11 },
		FieldModel{ Name: "storage_area", Type: VARCHAR.String(), Required : false, Readonly : true, Index : 12 },
		FieldModel{ Name: "storage_types", Type: MANYTOMANY.String(), Required : false, ForeignTable: SupportType.Name, Index: 13 },
		FieldModel{ Name: "projects", Type: MANYTOMANY.String(), Required : false, ForeignTable: Project.Name, Index: 14 },
		FieldModel{ Name: "valuations", Type: ONETOMANY.String(), Required : false, Level: LEVELRESPONSIBLE, ForeignTable: Valuation.Name, Index: 15 },
		FieldModel{ Name: "protections", Type: ONETOMANY.String(), Required : false, Level: LEVELRESPONSIBLE, ForeignTable: Protection.Name, Index: 16 },
		FieldModel{ Name: "restriction_types", Type: ONETOMANY.String(), Required : false, Level: LEVELRESPONSIBLE, ForeignTable: RestrictionType.Name, Index: 17 },
	},
}
var DEMODATASENUM = map[string][]string{
	"confidentiality_level" : confDatas,
	"protection_area" : protArDatas,
	"protection_type" : protTypDatas,
	"result_family" : resFamDatas,
	"result_type" : resTypDatas,
	"support_type" : supTypDatas,
	"support" : supDatas,
	"valuation_format" : valFormDatas,
	"valuation_type" : valTyp,
}
var DEMOROOTTABLES = []SchemaModel{ FormalizedData, Valuation, ValuationType, ValuationFormat, Support, SupportType, ResultType, ResultFamily, RestrictionType, ProtectionType, ProtectionArea, Protection, Project, FormalizedDataStorageType, FormalizedDataProject, ConfidentialityLevel }

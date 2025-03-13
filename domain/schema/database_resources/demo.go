package database

import "sqldb-ws/domain/schema/models"

// should set up as json better than a go file...

var ConfidentialityLevel = models.SchemaModel{
	Name:     "confidentiality_level",
	Label:    "confidentiality level",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}
var confDatas = []string{"public", "confidential project", "IRT confidential", "restricted diffusion", "special restricted diffusion", "classified data", "authorized profile"}

var FormalizedDataProject = models.SchemaModel{
	Name:     "formalized_data_project",
	Label:    "formalized data project",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(FormalizedData.Name), Type: models.INTEGER.String(), ForeignTable: FormalizedData.Name, Required: true, Index: 0, Label: "binded formalized data"},
		{Name: RootID(Project.Name), Type: models.INTEGER.String(), ForeignTable: Project.Name, Required: true, Index: 1, Label: "binded project"},
	},
}

var FormalizedDataStorageType = models.SchemaModel{
	Name:     "formalized_data_storage_type",
	Label:    "formalized data storage type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: RootID(FormalizedData.Name), Type: models.INTEGER.String(), ForeignTable: FormalizedData.Name, Required: true, Index: 0, Label: "binded formalized data"},
		{Name: RootID(SupportType.Name), Type: models.INTEGER.String(), ForeignTable: SupportType.Name, Required: true, Index: 1, Label: "binded storage type"},
	},
}

var Project = models.SchemaModel{ // todo
	Name:     "project",
	Label:    "project",
	Category: "global data",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "code", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 1},
		{Name: RootID(DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: DBEntity.Name, Required: true, Index: 2, Label: "related entity"},
	},
}

var Protection = models.SchemaModel{ // TODO
	Name:     "protection",
	Label:    "protection",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var protArDatas = []string{"france", "UE", "out of the UE"}

var ProtectionArea = models.SchemaModel{
	Name:     "protection_area",
	Label:    "protection area",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}
var protTypDatas = []string{"SOLEAU envelope", "timestamp", "patent application filed", "patent granted", "APP registration certificate", "no protection"}

var ProtectionType = models.SchemaModel{
	Name:     "protection_type",
	Label:    "protection type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var RestrictionType = models.SchemaModel{
	Name:     "restriction_type",
	Label:    "restriction type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "formalized_data_id", Type: models.INTEGER.String(), ForeignTable: "formalized_data", Required: true, Index: 1, Label: "related formalized data"},
	},
}

var resFamDatas = []string{"report", "internal technical note", "laboratory workbook", "career guide", "manuscript article", "thesis manuscript", "innovation filing sheet"}

var ResultFamily = models.SchemaModel{
	Name:     "result_family",
	Label:    "result family",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var resTypDatas = []string{"process", "method", "algorithm", "software", "specification", "conception", "test results", "sample/test vehicle", "demonstrator/proof of concept", "database", "standard", "innovation idea", "irt test bench"}

var ResultType = models.SchemaModel{
	Name:     "result_type",
	Label:    "result type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var supTypDatas = []string{"IRT server", "external data center", "hard disk", "USB", "paper storager", "physical conservation"}

var SupportType = models.SchemaModel{
	Name:     "support_type",
	Label:    "support type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var supDatas = []string{"paper", "digital"}

var Support = models.SchemaModel{
	Name:     "support",
	Label:    "support",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var Valuation = models.SchemaModel{ // TODO
	Name:     "valuation",
	Label:    "valuation",
	Category: "global data",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "formalized_data_id", Type: models.INTEGER.String(), ForeignTable: "formalized_data", Required: true, Index: 1, Label: "related formalized data"},
		{Name: RootID(ValuationType.Name), Type: models.INTEGER.String(), ForeignTable: ValuationType.Name, Required: true, Index: 2, Label: "related valuation type"},
		{Name: RootID(ValuationFormat.Name), Type: models.INTEGER.String(), ForeignTable: ValuationFormat.Name, Required: true, Index: 3, Label: "related valuation format"},
	},
}

var valFormDatas = []string{"scientific journal article", "conference presentation", "published thesis dissertation", "training", "BIP contribution to a new project", "license (patent/software)", "service provision"}

var ValuationFormat = models.SchemaModel{
	Name:     "valuation_format",
	Label:    "valuation format",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var valTyp = []string{"scientific", "economic"}

var ValuationType = models.SchemaModel{
	Name:     "valuation_type",
	Label:    "valuation type",
	Category: "",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
	},
}

var FormalizedData = models.SchemaModel{
	Name:     "formalized_data",
	Label:    "formalized data",
	CanOwned: true,
	Category: "global data",
	Fields: []models.FieldModel{
		{Name: "name", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: true, Index: 0},
		{Name: "ref", Type: models.VARCHAR.String(), Required: false, Label: "referencing", Readonly: true, Index: 1},
		{Name: "first_evaluation", Type: models.DECIMAL.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "first_evaluation_date", Type: models.TIMESTAMP.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 3},
		{Name: "actualized_evaluation", Type: models.DECIMAL.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 4},
		{Name: "capitalization_date", Type: models.TIMESTAMP.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 5},
		{Name: "committee_date", Type: models.TIMESTAMP.String(), Required: true, Level: models.LEVELRESPONSIBLE, Index: 6, Label: "committee reunion date"},
		{Name: "result_type_id", Type: models.INTEGER.String(), Required: false, ForeignTable: ResultType.Name, Index: 7, Label: "result type"},
		{Name: "result_family_id", Type: models.INTEGER.String(), Required: false, ForeignTable: ResultFamily.Name, Index: 8, Label: "result family"},
		{Name: "support_id", Type: models.INTEGER.String(), Required: false, ForeignTable: Support.Name, Index: 9, Label: "support"},
		{Name: "confidentiality_level_id", Type: models.INTEGER.String(), Required: false, Label: "confidentiality level", Level: models.LEVELRESPONSIBLE, ForeignTable: ConfidentialityLevel.Name, Index: 10},
		{Name: "contractual", Type: models.BOOLEAN.String(), Required: false, Level: models.LEVELRESPONSIBLE, Index: 11},
		{Name: "storage_area", Type: models.VARCHAR.String(), Required: false, Readonly: true, Index: 12},
		{Name: "storage_types", Type: models.MANYTOMANY.String(), Required: false, ForeignTable: SupportType.Name, Index: 13},
		{Name: "projects", Type: models.MANYTOMANY.String(), Required: false, ForeignTable: Project.Name, Index: 14},
		{Name: "valuations", Type: models.ONETOMANY.String(), Required: false, Level: models.LEVELRESPONSIBLE, ForeignTable: Valuation.Name, Index: 15},
		{Name: "protections", Type: models.ONETOMANY.String(), Required: false, Level: models.LEVELRESPONSIBLE, ForeignTable: Protection.Name, Index: 16},
		{Name: "restriction_types", Type: models.ONETOMANY.String(), Required: false, Level: models.LEVELRESPONSIBLE, ForeignTable: RestrictionType.Name, Index: 17},
	},
}
var DEMODATASENUM = map[string][]string{
	"confidentiality_level": confDatas,
	"protection_area":       protArDatas,
	"protection_type":       protTypDatas,
	"result_family":         resFamDatas,
	"result_type":           resTypDatas,
	"support_type":          supTypDatas,
	"support":               supDatas,
	"valuation_format":      valFormDatas,
	"valuation_type":        valTyp,
}
var DEMOROOTTABLES = []models.SchemaModel{FormalizedData, Valuation, ValuationType, ValuationFormat, Support, SupportType, ResultType, ResultFamily, RestrictionType, ProtectionType, ProtectionArea, Protection, Project, FormalizedDataStorageType, FormalizedDataProject, ConfidentialityLevel}

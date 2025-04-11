package database

import "sqldb-ws/domain/schema/models"

var CoCFR = models.SchemaModel{
	Name:     "competence center",
	Label:    "centre de compétence",
	Category: "domain",
	Fields: []models.FieldModel{
		{Name: "name", Label: "nom", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
	},
}

var ProjectFR = models.SchemaModel{
	Name:     "project",
	Label:    "projet",
	Category: "domain",
	Fields: []models.FieldModel{
		{Name: "name", Label: "nom", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
	},
}

var Axis = models.SchemaModel{
	Name:     "axis",
	Label:    "axe",
	Category: "domain",
	Fields: []models.FieldModel{
		{Name: "name", Label: "nom", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
	},
}

// should set up as json better than a go file...

var PublicationFR = models.SchemaModel{
	Name:     "publication",
	Label:    "publication",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "title", Label: "intitulé de la publication",
			Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
		{Name: "manager_" + RootID(DBUser.Name), Type: models.INTEGER.String(), Required: true,
			ForeignTable: DBUser.Name, Index: 1, Label: "responsable de la publication"},
		{Name: "project_accronym", Type: models.INTEGER.String(), Required: true,
			Index: 2, Label: "acronyme PROJET", ForeignTable: Project.Name},
		{Name: "axis", Type: models.INTEGER.String(), Required: true,
			Index: 3, Label: "axe", ForeignTable: Axis.Name},
		{Name: "competence_center", Type: models.INTEGER.String(), Required: true,
			Index: 4, Label: "centre de compétence", ForeignTable: Axis.Name},
		{Name: "authors", Type: models.ONETOMANY.String(), Required: true,
			Index: 5, Label: "auteurs", ForeignTable: DBUser.Name},
		{Name: "affiliation", Type: models.VARCHAR.String(), Required: true,
			Index: 6, Label: "affiliation"},
		{Name: "publication_type", Type: models.VARCHAR.String(), Required: true,
			Index: 6, Label: "type de publication"},
	},
}

var ArticleFR = models.SchemaModel{
	Name:     "article",
	Label:    "article journal/ chapitre d'ouvrage",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 0},
		{Name: "media_name", Label: "nom du journal", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 1},
		{Name: "DOI", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 2, Translatable: false},
		{Name: "publishing_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 3},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 4},
	},
}

var ConferenceFR = models.SchemaModel{
	Name:     "conference_presentation",
	Label:    "présentation avec acte de congrés",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 0},
		{Name: "acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 1},
		{Name: "name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 2},
		{Name: "start_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 3},
		{Name: "end_date", Label: "commentaires", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 4},
		{Name: "city", Label: "ville de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 5},
		{Name: "country", Label: "pays de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 6},
		{Name: "link", Label: "lien de la conférence", Type: models.LINK.String(), Required: false, Readonly: false, Index: 7},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 8},
	},
}

var PresentationFR = models.SchemaModel{
	Name:     "presentation",
	Label:    "présentation sans relecture (workshop, CST, GDR)",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 0},
		{Name: "conference_acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 0},
		{Name: "conference_name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 1},
		{Name: "meeting_name", Label: "nom du meeting", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 2},
		{Name: "meeting_date", Label: "date du meeting", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 3},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 4},
	},
}

var PosterFR = models.SchemaModel{
	Name:     "poster",
	Label:    "poster",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 0},
		{Name: "conference_acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 1},
		{Name: "conference_name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 2},
		{Name: "conference_start_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 3},
		{Name: "conference_end_date", Label: "commentaires", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 4},
		{Name: "conference_city", Label: "ville de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 5},
		{Name: "conference_country", Label: "pays de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 6},
		{Name: "conference_link", Label: "lien de la conférence", Type: models.LINK.String(), Required: false, Readonly: false, Index: 7},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 8},
	},
}

var HDRFR = models.SchemaModel{
	Name:     "HDR",
	Label:    "habilitation à diriger des recherches (HDR)",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 0},
		{Name: "defense_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 1},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 2},
	},
}

var ThesisFR = models.SchemaModel{
	Name:     "thesis",
	Label:    "thèse",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 0},
		{Name: "defense_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 1},
		{Name: "director_" + RootID(DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: DBUser.Name, Index: 3, Label: "directeur de thèse"},
		{Name: "co_supervisor_" + RootID(DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: DBUser.Name, Index: 4, Label: "co-encadrant de thèse"},
		{Name: "start_date", Label: "date de début de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 5},
		{Name: "end_date", Label: "date de end de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 6},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 7},
	},
}

var InternshipFR = models.SchemaModel{
	Name:     "internship",
	Label:    "stage",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 0},
		{Name: "IRT_manager" + RootID(DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: DBUser.Name, Index: 1, Label: "responsable IRT du stage"},
		{Name: "start_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 2},
		{Name: "end_date", Label: "date de end de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 3},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 4},
	},
}

var DemoFR = models.SchemaModel{
	Name:     "demo",
	Label:    "demo",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 0},
		{Name: "meeting_name", Label: "nom du meeting", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 1},
		{Name: "meeting_date", Label: "date du meeting", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 2},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 3},
	},
}

var OtherPublicationFR = models.SchemaModel{
	Name:     "other_publication",
	Label:    "autre publication",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 0},
		{Name: "publishing_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 1},
		{Name: "comments", Label: "commentaires", Type: models.BIGVARCHAR.String(), Required: false, Readonly: false, Index: 2},
	},
}

var PublicationTypeFR = models.SchemaModel{
	Name:     "publication_type",
	Label:    "type de publication",
	Category: "publications",
	Fields: []models.FieldModel{
		{Name: models.NAMEKEY, Constraint: "unique", Type: models.VARCHAR.String(), Required: true, Readonly: true, Index: 0},
		{Name: RootID(DBSchema.Name), Type: models.INTEGER.String(), ForeignTable: DBSchema.Name, Required: true, Readonly: true, Label: "template entry", Index: 3},
	},
}

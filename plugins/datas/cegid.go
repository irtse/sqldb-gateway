package datas

import (
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/schema/models"
)

var CoCFR = models.SchemaModel{
	Name:     "competence_center",
	Label:    "competence centers",
	Category: "domain",
	CanOwned: true,
	IsEnum:   true,
	Fields: []models.FieldModel{
		{Name: "name", Label: "nom", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
		{Name: ds.RootID(ds.DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: ds.DBEntity.Name, Required: true, Index: 1, Label: "entité en relation"},
	},
}

var Axis = models.SchemaModel{
	Name:     "axis",
	Label:    "IRT professional axis",
	CanOwned: true,
	Category: "domain",
	Fields: []models.FieldModel{
		{Name: "code", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Translatable: false, Readonly: false, Index: 0},
		{Name: "name", Type: models.VARCHAR.String(), Required: false, Translatable: false, Readonly: true, Index: 1},
		{Name: "domain_code", Label: "code domaine", Type: models.VARCHAR.String(), Translatable: false, Required: false, Readonly: true, Index: 2},
		{Name: ds.RootID(ds.DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: ds.DBEntity.Name, Required: true, Index: 3, Label: "entité en relation"},
	},
}

var ProjectFR = models.SchemaModel{ // todo
	Name:     "project",
	Label:    "projects",
	CanOwned: true,
	Category: "global data",
	Fields: []models.FieldModel{
		{Name: "code", Label: "code", Type: models.VARCHAR.String(), Constraint: "unique", Translatable: false, Required: true, Readonly: true, Index: 0},
		{Name: "name", Type: models.VARCHAR.String(), Required: false, Translatable: false, Readonly: true, Index: 1},
		{Name: "state", Type: models.VARCHAR.String(), Required: false, Default: models.STATEPENDING, Level: models.LEVELRESPONSIBLE, Index: 2},
		{Name: "project_task", Label: "lot projet", Type: models.VARCHAR.String(), Required: false, Readonly: true, Index: 3},
		{Name: "start_date", Label: "date de début de projet", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 4},
		{Name: "end_date", Label: "date de fin de projet", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 5},
		{Name: ds.RootID(Axis.Name), Label: "axe", Type: models.INTEGER.String(), ForeignTable: Axis.Name, Required: false, Index: 6},
		{Name: ds.RootID(ds.DBUser.Name), Label: "chef de projet", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 7},
		{Name: ds.RootID(ds.DBEntity.Name), Type: models.INTEGER.String(), ForeignTable: ds.DBEntity.Name, Required: true, Index: 8, Label: "entité en relation"},
	},
}

// should set up as json better than a go file...

var PublicationStatusFR = models.SchemaModel{
	Name:     "publication_status",
	Label:    "publication status",
	Category: "domain",
	IsEnum:   true,
	Fields: []models.FieldModel{
		{Name: "name", Label: "nom", Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
	},
}

var PublicationActConferenceFR = models.SchemaModel{
	Name:     "publication_act_conference",
	CanOwned: true,
	Label:    "publication act conference",
	Category: "domain",
	Fields: []models.FieldModel{
		{Name: "finalized_publication", Type: models.UPLOAD_STR.String(), Required: true,
			Index: 0, Label: "téléchargement de la publication finalisée"},
		{Name: "publishing_date", Label: "date effective de publication", Type: models.TIMESTAMP.String(), Required: true, Readonly: false, Index: 2},
		{Name: "major_conference", Label: "la conférence visée est-elle incontournable dans ton domaine scientifique ?", Type: models.ENUMBOOLEAN.String(), Required: false, Default: false, Readonly: false, Index: 3},
	},
}

var PublicationActArticleFR = models.SchemaModel{
	Name:     "publication_act_article",
	Label:    "publication act article",
	Category: "domain",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: "finalized_publication", Type: models.UPLOAD_STR.String(), Required: true,
			Index: 0, Label: "téléchargement de la publication finalisée"},
		{Name: "publishing_date", Label: "date effective de publication", Type: models.TIMESTAMP.String(), Required: true, Readonly: false, Index: 2},
		{Name: "major_conference", Label: "la publication est elle publiée dans un journal du premier quartile de ta discipline scientifique", Type: models.ENUMBOOLEAN.String(), Required: false, Default: false, Readonly: false, Index: 3},
	},
}

var PublicationActFR = models.SchemaModel{
	Name:     "publication_act",
	Label:    "publication act",
	Category: "domain",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: "finalized_publication", Type: models.UPLOAD_STR.String(), Required: true,
			Index: 0, Label: "téléchargement de la publication finalisée"},
		{Name: "publishing_date", Label: "date effective de publication", Type: models.TIMESTAMP.String(), Required: true, Readonly: false, Index: 2},
	},
}

var publicationFields = []models.FieldModel{
	{Name: "name", Label: "intitulé de la publication",
		Type: models.VARCHAR.String(), Constraint: "unique", Required: true, Readonly: false, Index: 0},
	{Name: ds.RootID(PublicationStatusFR.Name), Default: 1, Type: models.INTEGER.String(), ForeignTable: PublicationStatusFR.Name, Required: false, Readonly: true, Label: "statut de publication", Index: 1},
	{Name: "manager_" + ds.RootID(ds.DBUser.Name), Type: models.INTEGER.String(), Required: true,
		ForeignTable: ds.DBUser.Name, Index: 1, Label: "responsable de la publication"},
	{Name: "project_accronym", Type: models.INTEGER.String(), Required: true,
		Index: 2, Label: "acronyme PROJET", ForeignTable: Project.Name},
	{Name: "axis", Type: models.INTEGER.String(), Required: true,
		Index: 3, Label: "axe", ForeignTable: Axis.Name},
	{Name: "competence_center", Type: models.INTEGER.String(), Required: true,
		Index: 4, Label: "centre de compétence", ForeignTable: CoCFR.Name},
	{Name: "affiliation", Type: models.VARCHAR.String(), Required: true,
		Index: 5, Label: "affiliation (société ou laboratoire de ratachement du 1ier auteur)"},
	{Name: "publication", Type: models.UPLOAD_STR.String(), Required: true,
		Index: 6, Label: "téléchargement de la publication"},
	{Name: "is_awarded", Type: models.ENUMBOOLEAN.String(), Required: false, Default: "no",
		Index: 6, Label: "la production a-t-elle fait l'objet d'un award ?"},
	{Name: "awarded_by", Type: models.VARCHAR.String(), Required: false,
		Index: 6, Label: "la production a été primée par"},
}

var ArticleFR = models.SchemaModel{
	Name:     "article",
	Label:    "newspaper articles/book chapters",
	Category: "publications",
	CanOwned: true,
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: ArticleAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 9},
		{Name: "media_name", Label: "nom du journal", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 10},
		{Name: "DOI", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 11, Translatable: false},
		{Name: "publishing_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 12},
	}...),
}

var ArticleAuthorsFR = models.SchemaModel{
	Name:     "article_authors",
	Label:    "article authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("article"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "article", Required: true, Index: 2},
	},
}

var ConferenceFR = models.SchemaModel{
	Name:     "conference_presentation",
	Label:    "presentations with congress proceedings",
	CanOwned: true,
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: ConferenceAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 9},
		{Name: "acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 10},
		{Name: "name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 11},
		{Name: "start_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 12},
		{Name: "end_date", Label: "commentaires", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 13},
		{Name: "city", Label: "ville de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 14},
		{Name: "country", Label: "pays de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 15},
		{Name: "link", Label: "lien de la conférence", Type: models.URL.String(), Required: false, Readonly: false, Index: 16},
	}...),
}

var ConferenceAuthorsFR = models.SchemaModel{
	Name:     "conference_authors",
	Label:    "conference authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("conference_presentation"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "conference_presentation", Required: true, Index: 2},
	},
}

var PresentationFR = models.SchemaModel{
	Name:     "presentation",
	Label:    "presentations without proofreading (workshop, CST, GDR)",
	Category: "publications",
	CanOwned: true,
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: PresentationAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 9},
		{Name: "conference_acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 10},
		{Name: "conference_name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 12},
		{Name: "meeting_name", Label: "nom du meeting", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 12},
		{Name: "meeting_date", Label: "date du meeting", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 13},
	}...),
}

var PresentationAuthorsFR = models.SchemaModel{
	Name:     "presentation_authors",
	Label:    "presentation authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("presentation"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "presentation", Required: true, Index: 2},
	},
}

var PosterFR = models.SchemaModel{
	Name:     "poster",
	Label:    "posters",
	CanOwned: true,
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: PosterAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 9},
		{Name: "conference_acronym", Label: "nom de la conférence (acronyme)", Type: models.VARCHAR.String(), Required: true, Readonly: false, Index: 10},
		{Name: "conference_name", Label: "nom de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 11},
		{Name: "conference_start_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 12},
		{Name: "conference_end_date", Label: "date objective de fin de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 13},
		{Name: "conference_city", Label: "ville de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 14},
		{Name: "conference_country", Label: "pays de la conférence", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 15},
		{Name: "conference_link", Label: "lien de la conférence", Type: models.URL.String(), Required: false, Readonly: false, Index: 16},
	}...),
}

var PosterAuthorsFR = models.SchemaModel{
	Name:     "poster_authors",
	Label:    "poster authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("poster"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "poster", Required: true, Index: 2},
	},
}

var HDRFR = models.SchemaModel{
	Name:     "research_authorization",
	CanOwned: true,
	Label:    "authorizations to direct research",
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: HDRAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 9},
		{Name: "defense_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 10},
	}...),
}

var HDRAuthorsFR = models.SchemaModel{
	Name:     "research_authorization_authors",
	Label:    "research authorization authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("research_authorization"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "research_authorization", Required: true, Index: 2},
	},
}

var ThesisFR = models.SchemaModel{
	Name:     "thesis",
	Label:    "theses",
	CanOwned: true,
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: ThesisAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "yes", Required: false, Readonly: false, Index: 9},
		{Name: "defense_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 10},
		{Name: "director_" + ds.RootID(ds.DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: ds.DBUser.Name, Index: 11, Label: "directeur de thèse"},
		{Name: "co_supervisor_" + ds.RootID(ds.DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: ds.DBUser.Name, Index: 12, Label: "co-encadrant de thèse"},
		{Name: "start_date", Label: "date de début de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 13},
		{Name: "end_date", Label: "date de fin de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 14},
	}...),
}

var ThesisAuthorsFR = models.SchemaModel{
	Name:     "thesis_authors",
	Label:    "thesis authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("thesis"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "thesis", Required: true, Index: 2},
	},
}

var InternshipFR = models.SchemaModel{
	Name:     "internship",
	Label:    "internships",
	CanOwned: true,
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: InternshipAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 9},
		{Name: "IRT_manager" + ds.RootID(ds.DBUser.Name), Type: models.INTEGER.String(), Required: true, ForeignTable: ds.DBUser.Name, Index: 10, Label: "responsable IRT du stage"},
		{Name: "start_date", Label: "date de soutenance de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 11},
		{Name: "end_date", Label: "date de end de thèse", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 12},
	}...),
}

var InternshipAuthorsFR = models.SchemaModel{
	Name:     "internship_authors",
	Label:    "internship authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("internship"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "internship", Required: true, Index: 2},
	},
}

var DemoFR = models.SchemaModel{
	Name:     "demo",
	Label:    "demos",
	CanOwned: true,
	Category: "publications",
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: DemoAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 9},
		{Name: "meeting_name", Label: "nom du meeting", Type: models.VARCHAR.String(), Required: false, Readonly: false, Index: 10},
		{Name: "meeting_date", Label: "date du meeting", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 11},
	}...),
}

var DemoAuthorsFR = models.SchemaModel{
	Name:     "demo_authors",
	Label:    "demo authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("demo"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "demo", Required: true, Index: 2},
	},
}

var OtherPublicationFR = models.SchemaModel{
	Name:     "other_publication",
	Label:    "other publications",
	Category: "publications",
	CanOwned: true,
	Fields: append(publicationFields, []models.FieldModel{
		{Name: "authors", Type: models.MANYTOMANY.String(), Required: true, Index: 8, Label: "auteurs", ForeignTable: OtherPublicationAuthorsFR.Name},
		{Name: "reread", Label: "documents scientifiques relus par des pairs externes IRT et sélectionnés selon un process structuré", Type: models.ENUMBOOLEAN.String(), Default: "no", Required: false, Readonly: false, Index: 9},
		{Name: "publishing_date", Label: "date objective de publication", Type: models.TIMESTAMP.String(), Required: false, Readonly: false, Index: 10},
	}...),
}

var OtherPublicationAuthorsFR = models.SchemaModel{
	Name:     "other_publication_authors",
	Label:    "other publication authors",
	Category: "publications",
	CanOwned: true,
	Fields: []models.FieldModel{
		{Name: ds.RootID(ds.DBUser.Name), Label: "related user", Type: models.INTEGER.String(), ForeignTable: ds.DBUser.Name, Required: true, Index: 1},
		{Name: ds.RootID("other_publication"), Label: "related publication", Type: models.INTEGER.String(), ForeignTable: "other_publication", Required: true, Index: 2},
	},
}

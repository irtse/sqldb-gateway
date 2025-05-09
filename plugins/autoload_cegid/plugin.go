// compile with: go build -buildmode=plugin -o plugin.so plugin.go

// plugin.go
package main

import (
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	models "sqldb-ws/plugins/datas"
)

func Autoload() []sm.SchemaModel {
	ds.PERMISSIONEXCEPTION = append(ds.PERMISSIONEXCEPTION, []string{
		models.CoCFR.Name, models.ProjectFR.Name, models.Axis.Name,
		models.PublicationStatusFR.Name, models.ArticleFR.Name,
		models.OtherPublicationFR.Name,
		models.OtherPublicationAuthorsFR.Name,
		models.ArticleAuthorsFR.Name,
		models.ConferenceAuthorsFR.Name,
		models.DemoAuthorsFR.Name,
		models.HDRAuthorsFR.Name,
		models.InternshipAuthorsFR.Name,
		models.PosterAuthorsFR.Name,
		models.PresentationAuthorsFR.Name,
		models.ThesisAuthorsFR.Name,
		models.DemoFR.Name, models.InternshipFR.Name, models.ThesisFR.Name, models.HDRFR.Name,
		models.PosterFR.Name, models.PresentationFR.Name, models.ConferenceFR.Name,
	}...)
	ds.POSTPERMISSIONEXCEPTION = append(ds.POSTPERMISSIONEXCEPTION, []string{
		models.OtherPublicationFR.Name,
		models.ArticleFR.Name,
		models.DemoFR.Name,
		models.InternshipFR.Name,
		models.ThesisFR.Name,
		models.HDRFR.Name,
		models.PosterFR.Name,
		models.PresentationFR.Name,
		models.ConferenceFR.Name,

		models.OtherPublicationAuthorsFR.Name,
		models.ArticleAuthorsFR.Name,
		models.ConferenceAuthorsFR.Name,
		models.DemoAuthorsFR.Name,
		models.HDRAuthorsFR.Name,
		models.InternshipAuthorsFR.Name,
		models.PosterAuthorsFR.Name,
		models.PresentationAuthorsFR.Name,
		models.ThesisAuthorsFR.Name,
	}...)
	return []sm.SchemaModel{models.CoCFR, models.ProjectFR, models.Axis,
		models.OtherPublicationFR, models.DemoFR, models.InternshipFR, models.ThesisFR, models.HDRFR,
		models.PosterFR, models.PresentationFR, models.ConferenceFR,
		models.PublicationStatusFR, models.ArticleFR,
		models.OtherPublicationAuthorsFR,
		models.ArticleAuthorsFR,
		models.ConferenceAuthorsFR,
		models.DemoAuthorsFR,
		models.HDRAuthorsFR,
		models.InternshipAuthorsFR,
		models.PosterAuthorsFR,
		models.PresentationAuthorsFR,
		models.ThesisAuthorsFR,
	}
}

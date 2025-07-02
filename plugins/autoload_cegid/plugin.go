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
		models.ProofreadingStatus.Name,
		models.PublicationStatusFR.Name,
		models.OtherPublicationAuthorsFR.Name,
		models.OtherPublicationAffiliationAuthorsFR.Name,
		models.ArticleAuthorsFR.Name,
		models.ArticleAffiliationAuthorsFR.Name,
		models.ConferenceAuthorsFR.Name,
		models.ConferenceAffiliationAuthorsFR.Name,
		models.DemoAuthorsFR.Name,
		models.DemoAffiliationAuthorsFR.Name,
		models.HDRAuthorsFR.Name,
		models.HDRAffiliationAuthorsFR.Name,
		models.InternshipAuthorsFR.Name,
		models.InternshipAffiliationAuthorsFR.Name,
		models.PosterAuthorsFR.Name,
		models.PosterAffiliationAuthorsFR.Name,
		models.PresentationAuthorsFR.Name,
		models.PresentationAffiliationAuthorsFR.Name,
		models.ThesisAuthorsFR.Name,
		models.ThesisAffiliationAuthorsFR.Name,
		models.PublicationAwardFR.Name, models.PublicationActFR.Name,
		models.ArticleFR.Name, models.OtherPublicationFR.Name,
		models.DemoFR.Name, models.InternshipFR.Name, models.ThesisFR.Name, models.HDRFR.Name,
		models.PosterFR.Name, models.PresentationFR.Name, models.ConferenceFR.Name,
	}...)
	ds.POSTPERMISSIONEXCEPTION = append(ds.POSTPERMISSIONEXCEPTION, []string{
		models.OtherPublicationAuthorsFR.Name,
		models.OtherPublicationAffiliationAuthorsFR.Name,
		models.ArticleAuthorsFR.Name,
		models.ArticleAffiliationAuthorsFR.Name,
		models.ConferenceAuthorsFR.Name,
		models.ConferenceAffiliationAuthorsFR.Name,
		models.DemoAuthorsFR.Name,
		models.DemoAffiliationAuthorsFR.Name,
		models.HDRAuthorsFR.Name,
		models.HDRAffiliationAuthorsFR.Name,
		models.InternshipAuthorsFR.Name,
		models.InternshipAffiliationAuthorsFR.Name,
		models.PosterAuthorsFR.Name,
		models.PosterAffiliationAuthorsFR.Name,
		models.PresentationAuthorsFR.Name,
		models.PresentationAffiliationAuthorsFR.Name,
		models.ThesisAuthorsFR.Name,
		models.ThesisAffiliationAuthorsFR.Name,

		models.ArticleFR.Name,
		models.OtherPublicationFR.Name,
		models.PublicationAwardFR.Name, models.PublicationActFR.Name,
		models.DemoFR.Name, models.InternshipFR.Name, models.ThesisFR.Name, models.HDRFR.Name,
		models.PosterFR.Name, models.PresentationFR.Name, models.ConferenceFR.Name,
	}...)
	return []sm.SchemaModel{models.CoCFR, models.ProjectFR, models.Axis,
		models.OtherPublicationFR, models.DemoFR, models.InternshipFR, models.ThesisFR, models.HDRFR,
		models.PosterFR, models.PresentationFR, models.ConferenceFR,
		models.PublicationStatusFR, models.ArticleFR,
		models.OtherPublicationAuthorsFR,
		models.OtherPublicationAffiliationAuthorsFR,
		models.ArticleAuthorsFR,
		models.ArticleAffiliationAuthorsFR,
		models.ConferenceAuthorsFR,
		models.ConferenceAffiliationAuthorsFR,
		models.DemoAuthorsFR,
		models.ProofreadingStatus,
		models.DemoAffiliationAuthorsFR,
		models.HDRAuthorsFR,
		models.HDRAffiliationAuthorsFR,
		models.InternshipAuthorsFR,
		models.InternshipAffiliationAuthorsFR,
		models.PosterAuthorsFR,
		models.PosterAffiliationAuthorsFR,
		models.PresentationAuthorsFR,
		models.PresentationAffiliationAuthorsFR,
		models.ThesisAuthorsFR,
		models.ThesisAffiliationAuthorsFR,
		models.PublicationActFR,
		models.PublicationAwardFR,
	}
}

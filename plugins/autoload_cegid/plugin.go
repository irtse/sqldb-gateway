// compile with: go build -buildmode=plugin -o plugin.so plugin.go

// plugin.go
package main

import (
	sm "sqldb-ws/domain/schema/models"
	models "sqldb-ws/plugins/datas"
)

func Autoload() []sm.SchemaModel {
	return []sm.SchemaModel{models.CoCFR, models.ProjectFR, models.Axis,
		models.PublicationStatusFR, models.PublicationFR, models.ArticleFR, models.PublicationTypeFR, models.OtherPublicationFR,
		models.DemoFR, models.InternshipFR, models.ThesisFR, models.HDRFR, models.PosterFR, models.PresentationFR, models.ConferenceFR}
}

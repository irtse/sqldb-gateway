package controllers

import (
	"net/http"
	"os"

	"forge.redroom.link/yves/sqldb"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/rs/zerolog/log"
)

// Operations about schema
type HelperController struct {
	beego.Controller
}

// @Title ParseHeader
// @Description Post raw header
// @Param	body		body 	form data		"body of jsonform data"
// @Success 200 {string} success !
// @Failure 500 query error
// @router /header [post]
func (s *HelperController) ParseHeader() {
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.ListTables()
	if err != nil {
		log.Error().Msg(err.Error())
		s.Data["json"] = map[string]string{"error": err.Error()}
		s.Ctx.Output.SetStatus(http.StatusInternalServerError)
	}
	s.Data["json"] = data
	s.ServeJSON()

	db.Close()
}

// @Title CreateTable
// @Description Post raw header
// @Param	body		body 	form data		"body of jsonform data"
// @Success 200 {string} success !
// @Failure 500 query error
// @router /create [post]
func (s *HelperController) CreateTable() {
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.ListTables()
	if err != nil {
		log.Error().Msg(err.Error())
		s.Data["json"] = map[string]string{"error": err.Error()}
		s.Ctx.Output.SetStatus(http.StatusInternalServerError)
	}
	s.Data["json"] = data
	s.ServeJSON()

	db.Close()
}

// @Title Import
// @Description Post raw header
// @Param	body		body 	form data		"body of jsonform data"
// @Success 200 {string} success !
// @Failure 500 query error
// @router /import/:table [post]
func (s *HelperController) Import() {
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.ListTables()
	if err != nil {
		log.Error().Msg(err.Error())
		s.Data["json"] = map[string]string{"error": err.Error()}
		s.Ctx.Output.SetStatus(http.StatusInternalServerError)
	}
	s.Data["json"] = data
	s.ServeJSON()

	db.Close()
}

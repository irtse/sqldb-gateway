package controllers

import (
	"os"

	"forge.redroom.link/yves/sqldb"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/rs/zerolog/log"
)

// Operations about schema
type SchemaController struct {
	beego.Controller
}

// @Title GetTable
// @Description get list table
// @Success 200 {string} success !
// @Failure 403 no table
// @router / [get]
func (s *SchemaController) GetTable() {
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.ListTables()
	if err != nil {
		log.Error().Msg(err.Error())
		s.Data["json"] = map[string]string{"error": err.Error()}
	}
	s.Data["json"] = data
	s.ServeJSON()

	db.Close()
}

// @Title GetSchema
// @Description get table schema
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 success !
// @Failure 403 no table
// @router /:table [get]
func (s *SchemaController) GetSchema() {
	table := s.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.Table(table).GetSchema()
	if err != nil {
		log.Error().Msg(err.Error())
		s.Data["json"] = map[string]string{"error": err.Error()}
	}
	s.Data["json"] = data
	s.ServeJSON()

	db.Close()
}

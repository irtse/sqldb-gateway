package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sqldb-ws/security"
	"strings"

	"forge.redroom.link/yves/sqldb"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/rs/zerolog/log"
)

// Operations about table
type TableController struct {
	beego.Controller
}

type TableQuery struct {
	Table       string `json:"table,omitempty"`
	Columns     string `json:"columns,omitempty"`
	Restriction string `json:"restriction,omitempty"`
	Sortkeys    string `json:"sortkeys,omitempty"`
	Direction   string `json:"direction,omitempty"`
}

// @Title Put data in table
// @Description put data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	data		body 	json	true		"body for data content (Json format)"
// @Success 200 {string} success
// @Failure 403 :table put issue
// @router /:table [put]
func (t *TableController) Put() {
	// var FilterUserPost = func(ctx *context.Context) {

	// if strings.HasPrefix(ctx, "/") {
	// 	return
	// }

	// _, ok := ctx.Input.Session("user_id").(int)
	// if !ok {
	// 	ctx.Redirect(302, "/l")
	// }
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	defer db.Close()

	var data sqldb.AssRow
	json.Unmarshal(t.Ctx.Input.RequestBody, &data)
	println(fmt.Sprintf("%v", data))

	uid, err := db.Table(table).UpdateOrInsert(data)
	if err != nil {
		log.Error().Msg(err.Error())
		t.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		t.Ctx.Output.SetStatus(http.StatusOK)
	}
	t.Data["json"] = map[string]int64{"uid": uid}
	t.ServeJSON()
}

// web.InsertFilter("/*", web.BeforeRouter, FilterUserPost)
// }

// @Title Delete
// @Description delete the data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	body		body 			true		"body for data content (Json format)"
// @Success 200 {string} delete success!
// @Failure 403 delete issue
// @router /:table [delete]
func (t *TableController) Delete() {
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))

	var data sqldb.AssRow
	json.Unmarshal(t.Ctx.Input.RequestBody, &data)
	println(fmt.Sprintf("%v", data))

	db.Table(table).Delete(data)
	t.Data["json"] = "delete success!"
	t.Ctx.Output.SetStatus(http.StatusOK)
	t.ServeJSON()
	db.Close()
}

// @Title GetAllTable
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table [get]
func (t *TableController) GetAllTable() {
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	columns := []string{"*"}
	restriction := ""
	sortkeys := []string{}
	dir := ""
	data, err := db.Table(table).GetAssociativeArray(columns, restriction, sortkeys, dir)
	if err != nil {
		log.Error().Msg(err.Error())
		t.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		t.Data["json"] = data
	}
	str, err := json.Marshal(data)
	if err != nil {
		log.Error().Msg(err.Error())
		t.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	strToByte := []byte(strings.ReplaceAll(string(str), "\"\\u003cnil\\u003e\"", "null"))
	t.Ctx.Output.Header("Content-Type", "application/json")
	t.Ctx.Output.Body(strToByte)
	t.Ctx.Output.SetStatus(http.StatusOK)
	db.Close()
	t.Ctx.Output.Body(strToByte)
}

// @Title GetAllTableColumn
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Param	columns			path 	string	true		"Name of the columns (separate with a comma)"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/:columns [get]
func (t *TableController) GetAllTableColumn() {
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	columns := strings.Split(t.GetString(":columns"), ",")
	restriction := ""
	sortkeys := []string{}
	dir := ""
	data, err := db.Table(table).GetAssociativeArray(columns, restriction, sortkeys, dir)
	if err != nil {
		log.Error().Msg(err.Error())
		t.Data["json"] = map[string]string{"error": err.Error()}
	} else {
		t.Data["json"] = data
	}
	t.ServeJSON()
	db.Close()
}

// @Title GetAllTableColumnRestriction
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Param	columns			path 	string	true		"Name of the columns (separate with a comma)"
// @Param	restriction		path 	string	true		"SQL restriction"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/:columns/:restriction [get]
func (t *TableController) GetAllTableColumnRestriction() {
	table := t.GetString(":table")

	columns := fmt.Sprintf("%v", strings.Split(t.GetString(":columns"), ","))
	cols := strings.Split(t.GetString(":columns"), ",")
	restriction := t.GetString(":restriction")
	sortkeys := []string{}
	dir := ""
	dbuser_id := fmt.Sprintf("%v", 1)

	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	security.CheckSelect(dbuser_id, &table, &columns, &restriction)
	data, err := db.Table(table).GetAssociativeArray(cols, restriction, sortkeys, dir)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	data2 := fmt.Sprintf("%v", data)
	fmt.Println(data2)
	t.Data["json"] = data

	t.ServeJSON()

	db.Close()

}

// @Title GetAllTableColumnRestrictionSortkeys
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Param	columns			path 	string	true		"Name of the columns (separate with a comma)"
// @Param	restriction		path 	string	true		"SQL restriction"
// @Param	sortkeys		path	string	true		"Order by: columns names (separate with a comma)"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/:columns/:restriction/:sortkeys [get]
func (t *TableController) GetAllTableColumnRestrictionSortkeys() {
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	columns := strings.Split(t.GetString(":columns"), ",")
	restriction := t.GetString(":restriction")
	sortkeys := strings.Split(t.GetString(":sortkeys"), ",")
	dir := ""
	data, err := db.Table(table).GetAssociativeArray(columns, restriction, sortkeys, dir)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	t.Data["json"] = data
	t.ServeJSON()
	db.Close()
}

// @Title GetAllTableColumnRestrictionSortkeysDir
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Param	columns			path 	string	true		"Name of the columns (separate with a comma)"
// @Param	restriction		path 	string	true		"SQL restriction"
// @Param	sortkeys		path	string	true		"Order by: columns names (separate with a comma)"
// @Param	dir				path	string	true		"asc or desc"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/:columns/:restriction/:sortkeys/:dir [get]
func (t *TableController) GetAllTableColumnRestrictionSortkeysDir() {
	table := t.GetString(":table")
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	columns := strings.Split(t.GetString(":columns"), ",")
	restriction := t.GetString(":restriction")
	sortkeys := strings.Split(t.GetString(":sortkeys"), ",")
	dir := t.GetString(":dir")
	data, err := db.Table(table).GetAssociativeArray(columns, restriction, sortkeys, dir)
	if err != nil {
		log.Error().Msg(err.Error())
	}
	t.Data["json"] = data
	t.ServeJSON()
	db.Close()
}

// @Title TablePost
// @Description get all Datas
// @Param	table			path 	string	true		"Name of the table"
// @Param	body		body 	TableQuery	true		"TableQuery"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table [post]
func (t *TableController) TablePost() {
	table := t.GetString(":table")
	var request TableQuery
	json.Unmarshal(t.Ctx.Input.RequestBody, &request)

	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	data, err := db.Table(table).GetAssociativeArray(strings.Split(request.Columns, ","), request.Restriction, strings.Split(request.Sortkeys, ","), request.Direction)
	if err != nil {
		log.Error().Msg(err.Error())
		t.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	t.Data["json"] = data
	t.Ctx.Output.SetStatus(http.StatusOK)
	t.ServeJSON()
	db.Close()
}

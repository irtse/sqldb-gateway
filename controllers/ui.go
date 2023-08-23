package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"forge.redroom.link/yves/sqldb"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/rs/zerolog/log"
)

// Operations about table
type UiController struct {
	beego.Controller
}

// @Title form
// @Description create access form
// @Param	fid		path 	string	true		"The fid of the form"
// @Param	uid		path 	string	true		"The uid you want to edit"
// @Success 200 json form
// @Failure 403 body is empty
// @router /:fid/:uid [get]
func (u *UiController) GetEditForm() {
	fid := u.Ctx.Input.Query(":fid")
	uid := u.Ctx.Input.Query(":uid")

	u.buildForm(fid, uid)
}

// @Title form
// @Description create access form
// @Param	fid		path 	string	true		"The fid of the form"
// @Success 200 json form
// @Failure 403 body is empty
// @router /:fid [get]
func (u *UiController) GetEmptyForm() {
	fid := u.Ctx.Input.Query(":fid")
	u.buildForm(fid, "")
}

func (u *UiController) buildForm(fid string, uid string) {

	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	// Get form data
	formdesc, err := getFormDesc(db, fid)
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	form := make(map[string]interface{})
	form["title"] = formdesc[0]["title"].(string)
	form["description"] = formdesc[0]["header"].(string)

	formfields, err := getFormFields(db, fid)
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	var fields []interface{}
	columnnames := []string{"id"}
	for _, f := range formfields {
		columnnames = append(columnnames, f["columnname"].(string))
	}
	// Get table schema
	schema, err := db.Table(formdesc[0]["tablename"].(string)).GetSchema()
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	var data []sqldb.AssRow
	//Get edited Item data if item id is provided
	if uid != "" {
		data, err = getData(db, formdesc[0]["tablename"].(string), columnnames, uid)
	}
	for _, f := range formfields {
		if f["columnname"].(string) != "id" {
			field := make(map[string]interface{})
			field["key"] = f["columnname"].(string)
			if uid != "" {
				field["value"] = fmt.Sprintf("%v", data[0][f["columnname"].(string)])
			}
			field["type"] = f["fieldtype"].(string)
			field["label"] = f["label"].(string)
			dbType := schema.Columns[f["columnname"].(string)][:strings.Index(schema.Columns[f["columnname"].(string)], "|")]
			// foreign keys
			if f["columnname"].(string)[len(f["columnname"].(string))-3:] == "_id" {
				fklist := []map[string]interface{}{}
				// Query FK
				columns := strings.Split(f["linkcolumns"].(string), ",")
				sortkeys := strings.Split(f["linkorder"].(string), ",")
				restriction := f["linkrestriction"].(string)
				dir := ""
				fk, err := db.Table(f["columnname"].(string)[:len(f["columnname"].(string))-3]).GetAssociativeArray(columns, restriction, sortkeys, dir)
				if err != nil {
					log.Error().Msg(err.Error())
					u.Ctx.Output.SetStatus(http.StatusBadRequest)
				}
				for _, v := range fk {
					item := make(map[string]interface{})
					item["value"] = fmt.Sprintf("%v", v["id"])
					item["label"] = v["label"]
					fklist = append(fklist, item)
				}
				field["items"] = fklist
			}
			// other
			switch dbType {
			case "integer":

			case "float", "double":

			case "varchar":

			}
			if uid != "" {
				// Force data values to the right type if required
				switch field["type"] {
				case "Radio":
					field["value"] = data[0][f["columnname"].(string)]
					// force string for all text fields, whatever data type
				default:
					field["value"] = fmt.Sprintf("%v", data[0][f["columnname"].(string)])
				}
			}

			fields = append(fields, field)
		}
	}
	form["fields"] = fields
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		u.Data["json"] = form
	}
	u.ServeJSON()
}

// @Title Access form data post
// @Description insert access
// @Param	uid		path 	string	true		"The uid you want to edit"
// @Param	body		body 	form data		"body of jsonform data"
// @Success 200 json
// @Failure 403 body is empty
// @router /:uid [post]
func (u *UiController) PostAccessForm() {
	var err error
	uid := u.Ctx.Input.Query(":uid")
	var formdata map[string]interface{}
	json.Unmarshal(u.Ctx.Input.RequestBody, &formdata)
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	accesslist := formdata["fields"].([]interface{})
	switcheslist := accesslist[:len(accesslist)-3]
	families := []string{}
	for _, accessif := range switcheslist {
		access := accessif.(map[string]interface{})
		basefamily := fmt.Sprint(access["key"].(float64))
		if access["value"].(bool) {
			_, err = db.QueryAssociativeArray("insert ignore into user_basefamily(user_id, family_id) values(" + uid + "," + basefamily + ")")
			if err != nil {
				log.Error().Msg(err.Error())
				u.Ctx.Output.SetStatus(http.StatusBadRequest)
			}
			families = append(families, basefamily)
		} else {
			// remove off
			_, err = db.QueryAssociativeArray("delete from user_basefamily where user_id=" + uid + " and family_id=" + basefamily)
			if err != nil {
				log.Error().Msg(err.Error())
				u.Ctx.Output.SetStatus(http.StatusBadRequest)
			}
		}

	}

	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		u.Data["json"] = "ok"
	}
	u.ServeJSON()
}

func getFormDesc(db *sqldb.Db, fid string) ([]sqldb.AssRow, error) {
	columns := []string{"*"}
	restriction := "id=" + fid
	sortkeys := []string{}
	dir := ""
	return db.Table("dbform").GetAssociativeArray(columns, restriction, sortkeys, dir)
}

func getFormFields(db *sqldb.Db, fid string) ([]sqldb.AssRow, error) {
	columns := []string{"*"}
	restriction := "form_id=" + fid
	sortkeys := []string{"columnorder"}
	dir := ""
	return db.Table("dbformfields").GetAssociativeArray(columns, restriction, sortkeys, dir)
}

func getData(db *sqldb.Db, table string, columns []string, uid string) ([]sqldb.AssRow, error) {
	restriction := "id=" + uid
	sortkeys := []string{}
	dir := ""
	return db.Table(table).GetAssociativeArray(columns, restriction, sortkeys, dir)
}

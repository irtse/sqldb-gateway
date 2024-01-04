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
// @router /form/:fid/:uid [get]
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
// @router /form/:fid [get]
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
// @Param	fid		path 	string	true		"The fid of the form"
// @Param	uid		path 	string	true		"The uid you want to edit"
// @Param	body		body 	form data		"body of jsonform data"
// @Success 200 json
// @Failure 403 body is empty
// @router /form/:fid/:uid [post]
func (u *UiController) PostAccessForm() {
	var err error
	fid := u.Ctx.Input.Query(":uid")
	uid := u.Ctx.Input.Query(":uid")
	var formdata map[string]interface{}
	json.Unmarshal(u.Ctx.Input.RequestBody, &formdata)
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	formdesc, err := getFormDesc(db, fid)
	table := formdesc[0]["tablename"]
	print(table, uid)
	updateJson := map[string]interface{}{}
	fields := formdata["fields"].([]map[string]interface{})
	for _, f := range fields {
		// todo manage types
		updateJson[f["key"].(string)] = f["value"]
	}
	// todo send update
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		u.Data["json"] = "ok"
	}
	u.ServeJSON()
}

// @Title Tableview
// @Description Get table view
// @Param	tvid		path 	string	true		"The id of the tableview"
// @Success 200 json form
// @Failure 403 body is empty
// @router /tableview/:tvid [get]
func (u *UiController) GetListView() {
	lvid := u.Ctx.Input.Query(":tvid")
	tvdata := map[string]interface{}{}
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	lv, err := getTableView(db, lvid)
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	tvdata["title"] = lv[0]["title"].(string)
	tvdata["header"] = lv[0]["header"].(string)
	tvdata["form_id"] = lv[0]["form_id"].(int64)
	if lv[0]["tablerestriction"] == nil {
		tvdata["tablerestriction"] = ""
	} else {
		tvdata["tablerestriction"] = lv[0]["tablerestriction"].(string)
	}
	if lv[0]["tableorder"] == nil {
		tvdata["tableorder"] = ""
	} else {
		tvdata["tableorder"] = lv[0]["tableorder"].(string)
	}
	if lv[0]["tableorderdir"] == nil {
		tvdata["tableorderdir"] = ""
	} else {
		tvdata["tableorderdir"] = lv[0]["tableorderdir"].(string)
	}
	schema, err := db.Table(lv[0]["tablename"].(string)).GetSchema()
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	}
	tvdata["columns"] = schema.Columns
	data, err := db.Table(lv[0]["tablename"].(string)).GetAssociativeArray(strings.Split("id,"+lv[0]["tablecolumns"].(string), ","), tvdata["tablerestriction"].(string), strings.Split(tvdata["tableorder"].(string), ","), tvdata["tableorderdir"].(string))
	tvdata["items"] = data
	if err != nil {
		log.Error().Msg(err.Error())
		u.Ctx.Output.SetStatus(http.StatusBadRequest)
	} else {
		u.Data["json"] = tvdata
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

func getTableView(db *sqldb.Db, lvid string) ([]sqldb.AssRow, error) {
	columns := []string{"*"}
	restriction := "id=" + lvid
	sortkeys := []string{}
	dir := ""
	return db.Table("dbtableview").GetAssociativeArray(columns, restriction, sortkeys, dir)
}

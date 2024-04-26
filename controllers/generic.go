package controllers

import (
	"fmt"
	"sqldb-ws/lib/domain/utils"
	"sqldb-ws/lib/domain/schema"
)

type MainController struct { AbstractController }
// Operations about table
type GenericController struct { AbstractController }

// @Title /
// @Description Main call
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @router / [get]
func (l *MainController) Main() {
	// Main is the default root of the API, it gives back all your allowed shallowed view
	l.paramsOverload = map[string]string{ utils.RootTableParam : schema.DBView.Name, utils.RootRowsParam : utils.ReservedParam,
										  utils.RootShallow : "enable", "indexable" : fmt.Sprintf("%v", true) }
	l.SafeCall(utils.SELECT)
}
// @Title Post data in table
// @Description post data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	data		body 	json	true		"body for data content (Json format)"
// @Success 200 {string} success
// @Failure 403 :table post issue
// @router /:table [post]
func (t *GenericController) Post() { t.SafeCall(utils.CREATE) }
// @Title Put data in table
// @Description put data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	data		body 	json	true		"body for data content (Json format)"
// @Success 200 {string} success
// @Failure 403 :table put issue
// @router /:table [put]
func (t *GenericController) Put() { t.SafeCall(utils.UPDATE) }
// web.InsertFilter("/*", web.BeforeRouter, FilterUserPost)
// }

// @Title Delete
// @Description delete the data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	body		body 			true		"body for data content (Json format)"
// @Success 200 {string} delete success!
// @Failure 403 delete issue
// @router /:table [delete]
func (t *GenericController) Delete() { t.SafeCall(utils.DELETE) }

// @Title Get
// @Description get Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table [get]
func (t *GenericController) Get() { t.SafeCall(utils.SELECT) }
// @Title Count
// @Description count Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table [get]
func (t *GenericController) Count() { t.SafeCall(utils.COUNT) }

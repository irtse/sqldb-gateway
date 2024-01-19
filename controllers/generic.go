package controllers

import (
	tool "sqldb-ws/lib"
)

// Operations about table
type GenericController struct { AbstractController }
// @Title Post data in table
// @Description post data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	data		body 	json	true		"body for data content (Json format)"
// @Success 200 {string} success
// @Failure 403 :table post issue
// @router /:table [post]
func (t *GenericController) Post() { t.SafeCall(tool.CREATE, "CreateOrUpdate") }
// @Title Put data in table
// @Description put data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	data		body 	json	true		"body for data content (Json format)"
// @Success 200 {string} success
// @Failure 403 :table put issue
// @router /:table [put]
func (t *GenericController) Put() { t.SafeCall(tool.UPDATE, "CreateOrUpdate") }
// web.InsertFilter("/*", web.BeforeRouter, FilterUserPost)
// }

// @Title Delete
// @Description delete the data in table
// @Param	table		path 	string	true		"Name of the table"
// @Param	body		body 			true		"body for data content (Json format)"
// @Success 200 {string} delete success!
// @Failure 403 delete issue
// @router /:table [delete]
func (t *GenericController) Delete() { t.SafeCall(tool.DELETE, "Delete") }

// @Title Get
// @Description get Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table [get]
func (t *GenericController) Get() { t.SafeCall(tool.SELECT, "Get") }
// @Title Import
// @Description post Import
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/import/:filename [post]
func (t *GenericController) Importated() { t.SafeCall(tool.CREATE, "Import", t.GetString(":filename")) }
// @Title Delete by Import
// @Description delete Import
// @Param	table path 	string true "Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:table/import/:filename [delete]
func (t *GenericController) NotImportated() { t.SafeCall(tool.DELETE, "Import", t.GetString(":filename")) }
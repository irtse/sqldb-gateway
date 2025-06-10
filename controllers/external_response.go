package controllers

import (
	"errors"
	"os"
	"sqldb-ws/controllers/controller"
	"sqldb-ws/domain"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

// Operations about login
type ExternalResponseController struct{ controller.AbstractController }

// LLDAP HERE
// func (l *AuthController) LoginLDAP() { }

// @Title get
// @Description Post get external response
// @Param	body		body 	Response	true		"Response"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @router /:code [get]
func (e *ExternalResponseController) Post() {
	code := e.Ctx.Input.Params()[":code"]
	p, _ := e.Params()
	body := map[string]interface{}{}
	for k, v := range p {
		if v == "true" {
			body[k] = true
		} else if v == "false" {
			body[k] = false
		} else {
			body[k] = v
		}
	}
	d := domain.Domain(true, "", nil)
	if res, err := d.GetDb().ClearQueryFilter().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		"code": connector.Quote(code),
	}, false); err == nil && len(res) > 0 {
		emailRelated := res[0]
		body[ds.EmailSendedDBField] = emailRelated[utils.SpecialIDParam]
		if _, err := d.CreateSuperCall(utils.AllParams(ds.DBEmailResponse.Name).Enrich(map[string]interface{}{
			"code": code,
		}), body); err != nil {
			e.Response(utils.Results{}, err, "", "")
			return
		}
		e.Ctx.Output.ContentType("text/html") // Optional, Beego usually handles it
		target := os.Getenv("LANG")
		if target == "" {
			target = "fr"
		}
		f, err := os.ReadFile("/opt/html/index_" + target + ".html")
		if err != nil {
			e.Response(utils.Results{}, err, "", "")
		}
		content := string(f)
		e.Ctx.WriteString(content)
	} else {
		e.Response(utils.Results{}, errors.New("not a valid code response"), "", "")
	}
}

package controllers

import (
	"errors"
	"fmt"
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
	fmt.Println("EXTERNAL")
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
	if res, err := d.GetDb().SelectQueryWithRestriction(ds.DBEmailSended.Name, map[string]interface{}{
		"code": connector.Quote(code),
	}, false); err == nil && len(res) > 0 {
		emailRelated := res[0]
		body[ds.EmailSendedDBField] = emailRelated[utils.SpecialIDParam]
		if _, err := d.CreateSuperCall(utils.AllParams(ds.DBEmailResponse.Name), body); err != nil {
			e.Response(utils.Results{}, err, "", "")
		}
		e.Ctx.Output.ContentType("html") // Optional, Beego usually handles it
		e.Ctx.WriteString(fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<title>%s</title>
			<style>
				body {
					background: #f0f4f8;
					font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
					display: flex;
					justify-content: center;
					align-items: center;
					height: 100vh;
				}
				.card {
					background: white;
					padding: 40px 60px;
					border-radius: 12px;
					box-shadow: 0 8px 16px rgba(0,0,0,0.15);
					text-align: center;
				}
				h1 {
					color: #4CAF50;
					margin-bottom: 20px;
				}
				p {
					color: #333;
					font-size: 1.2em;
				}
				.button {
					margin-top: 20px;
					padding: 10px 20px;
					background-color: #4CAF50;
					color: white;
					text-decoration: none;
					border-radius: 6px;
					transition: background-color 0.3s ease;
				}
				.button:hover {
					background-color: #45a049;
				}
			</style>
		</head>
		<body>
			<div class="card">
				<h1>%s</h1>
				<p>%s</p>
			</div>
		</body>
		</html>
		`, utils.Translate("Thank you"),
			utils.Translate("Thanks for your answer!"),
			utils.Translate("We appreciate your feedback and your time."),
		))
		return
	}
	e.Response(utils.Results{}, errors.New("not a valid code response"), "", "")
}

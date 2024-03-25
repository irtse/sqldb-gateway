package controllers

import (
	"fmt"
	"errors"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/entities"
	domain "sqldb-ws/lib/domain"
	"github.com/matthewhartstonge/argon2"
)
// Operations about login
type AuthController struct { AbstractController }
// LLDAP HERE
// func (l *AuthController) LoginLDAP() { }

// @Title Login
// @Description User login
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @router /login [post]
func (l *AuthController) Login() { 
	// login function will overide generic procedure foundable in controllers.go
	
	body := l.body(false) // extracting body
	if log, ok := body["login"]; ok { // search for login in body 
		params := l.paramsOver(map[string]string{ tool.RootTableParam : entities.DBUser.Name, 
												  tool.RootRowsParam : tool.ReservedParam, })
		d := domain.Domain(false, log.(string), false) // create a new domain with current permissions of user
		d.Specialization = false // when launching call disable every auth check up (don't forget you are not logged)
		response, err := d.SuperCall(params, tool.Record{}, tool.SELECT, "Get", "name='" + log.(string) + "' OR email='" + log.(string) + "'")
		if err != nil {  l.response(response, err); return }
		if len(response) == 0 {  l.response(response, errors.New("AUTH : username/email invalid")); return }
		// if no problem check if logger is authorized to work on API and properly registered
		pass, ok := body["password"] // then compare password founded in base and ... whatever... you know what's about
		pwd, ok1 := response[0]["password"].(string)
		if ok && ok1 {
			if ok, err := argon2.VerifyEncoded([]byte(pass.(string)), []byte(pwd)); ok && err == nil{
				// when password matching
				token := l.session(log.(string), response[0]["super_admin"].(bool), false) // update session variables
				response[0]["token"]=token
				d := domain.Domain(false, log.(string), false) 
				params := tool.Params{ tool.RootTableParam : entities.DBNotification.Name, 
									   tool.RootRowsParam : tool.ReservedParam, 
									   tool.RootRawView : "enable",}
				notifs, err := d.PermsSuperCall(params, tool.Record{}, tool.SELECT, "Get")
				n := tool.Results{}
				for _, notif := range notifs {
					params := tool.Params{ tool.RootTableParam : entities.DBView.Name, 
										   tool.RootRowsParam : tool.ReservedParam,
										   tool.RootShallow : "enable",
										   entities.RootID(entities.DBSchema.Name) : notif.GetString(entities.DBSchema.Name),
										   entities.RootID("dest_table") : notif.GetString(entities.RootID("dest_table")),
										}
					views, err := d.Call(params, tool.Record{}, tool.SELECT, "Get")
					if err == nil || len(views) > 0 {
						id := "-1"
						for _, view := range views {
							if view["max"] != nil && view["max"].(int64) > 0 {
								id = fmt.Sprintf("%v", view["id"])
								break
							}
						}
						n = append(n, tool.Record{
							entities.NAMEATTR : notif.GetString(entities.NAMEATTR),
							"description" : notif.GetString("description"),
							"data_ref" : "#" + id + ":" + notif.GetString(entities.RootID("dest_table")),
						})
					}
					
				}
				if err == nil { response[0]["notifications"]=n
				} else { response[0]["notifications"]=[]interface{}{} }
				l.response(response, nil)
				return
			}
		}	
		l.response(tool.Results{}, errors.New("AUTH : password invalid")) // API response
		return 
	}
	l.response(tool.Results{}, errors.New("AUTH : username/email invalid")) 
}

// @Title Logout
// @Description User logout
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @Failure 402 user already connected
// @router /logout [get]
func (l *AuthController) Logout() {
	login, superAdmin, err := l.authorized() // check if already connected
	if err != nil { l.response(nil, err) }
	l.session(login, superAdmin, true) // update session variables
	l.response(tool.Results{ tool.Record { "name" : login }}, nil)
}
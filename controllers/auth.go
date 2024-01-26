package controllers

import (
	"errors"
	tool "sqldb-ws/lib"
	domain "sqldb-ws/lib/domain"
	"github.com/matthewhartstonge/argon2"
)
// Operations about login
type AuthController struct { AbstractController }
// @Title Login
// @Description User login
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @router /login [post]
func (l *AuthController) Login() {
	body := l.body(false)
	if log, ok := body["login"]; ok {
		params := l.paramsOver(map[string]string{ tool.RootTableParam : "dbuser", 
												  tool.RootRowsParam : "all", 
												  "login" : log.(string) })
		d := domain.Domain(false, log.(string), false)
		d.Specialization = false
		response, err := d.Call(params, tool.Record{}, tool.SELECT, false, "Get")
		if err != nil {  l.response(response, err); return }
		if len(response) == 0 {  l.response(response, errors.New("AUTH : username/email invalid")); return }
		user_id, _, err := l.authorized()
		if err == nil && user_id == log.(string) { // token verify
			l.response(response, errors.New("already log in")); return
		}
		pass, ok := body["password"]
		pwd, ok1 := response[0]["password"].(string)
		if ok && ok1 {
			if ok, err := argon2.VerifyEncoded([]byte(pass.(string)), []byte(pwd)); ok && err == nil{
				l.session(log.(string), response[0]["super_admin"].(bool), false)
				l.response(response, nil)
				return
			}
		}	
		l.response(response, errors.New("AUTH : password invalid"))
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
	login, superAdmin, err := l.authorized()
	if err != nil { l.response(nil, err) }
	l.session(login, superAdmin, true)
	l.response(tool.Results{ tool.Record { "login" : login }}, nil)
}
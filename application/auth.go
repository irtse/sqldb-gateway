package application

import (
	"errors"
	"fmt"
	"sqldb-ws/application/controller"
	"sqldb-ws/domain"
	"sqldb-ws/domain/utils"

	"github.com/matthewhartstonge/argon2"
)

// Operations about login
type AuthController struct{ controller.AbstractController }

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
	body := l.Body(false)             // extracting body
	if log, ok := body["login"]; ok { // search for login in body
		response, err := domain.IsLogged(false, fmt.Sprintf("%v", log), "")
		if err != nil {
			l.Response(response, err)
			return
		}
		if len(response) == 0 {
			l.Response(response, errors.New("AUTH : username/email invalid"))
			return
		}
		// if no problem check if logger is authorized to work on API and properly registered
		pass, ok := body["password"] // then compare password founded in base and ... whatever... you know what's about
		pwd, ok1 := response[0]["password"].(string)
		if ok && ok1 {
			if ok, err := argon2.VerifyEncoded([]byte(pass.(string)), []byte(pwd)); ok && err == nil {
				// when password matching
				token := l.MySession(log.(string), response[0]["super_admin"].(bool), false) // update session variables
				response[0]["token"] = token
				l.Response(response, nil)
				return
			}
		}
		l.Response(utils.Results{}, errors.New("AUTH : password invalid")) // API response
		return
	}
	l.Response(utils.Results{}, errors.New("AUTH : can't find login data"))
}

// @Title Logout
// @Description User logout
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @Failure 402 user already connected
// @router /logout [get]
func (l *AuthController) Logout() {
	login, superAdmin, err := l.IsAuthorized() // check if already connected
	if err != nil {
		l.Response(nil, err)
	}
	l.MySession(login, superAdmin, true) // update session variables
	l.Response(utils.Results{utils.Record{"name": login}}, nil)
}

// @Title Refresh
// @Description User logout
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @Failure 402 user already connected
// @router /logout [get]
func (l *AuthController) Refresh() {
	login, superAdmin, err := l.IsAuthorized() // check if already connected
	if err != nil {
		l.Response(nil, err)
	}
	token := l.MySession(login, superAdmin, false) // update session variables
	response, err := domain.IsLogged(true, login, token)
	l.Response(response, err)
}

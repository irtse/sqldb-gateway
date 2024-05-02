package controllers

import (
	"fmt"
	"time"
	"errors"
	"sqldb-ws/lib/domain/utils"
	"sqldb-ws/lib/domain/schema"
	domain "sqldb-ws/lib/domain"
	"github.com/golang-jwt/jwt/v5"
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
		params := l.paramsOver(utils.AllParams(schema.DBUser.Name))
		d := domain.Domain(false, log.(string), false) // create a new domain with current permissions of user
		// d.Specialization = false // when launching call disable every auth check up (don't forget you are not logged)
		response, err := d.SuperCall(params, utils.Record{}, utils.SELECT, "name='" + log.(string) + "' OR email='" + log.(string) + "'")
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
				params := utils.Params{ utils.RootTableParam : schema.DBNotification.Name, 
									   utils.RootRowsParam : utils.ReservedParam, 
									   utils.RootRawView : "enable", }
				notifs, err := d.PermsSuperCall(params, utils.Record{}, utils.SELECT)
				n := utils.Results{}
				for _, notif := range notifs {
					sch, err := schema.GetSchemaByID(int64(notif["link_id"].(float64)))
					if err != nil { continue }
					n = append(n, utils.Record{
						utils.SpecialIDParam : notif.GetString(utils.SpecialIDParam),
						schema.NAMEKEY : notif.GetString(schema.NAMEKEY),
						"description" : notif.GetString("description"),
						"link_path" : "/" + utils.MAIN_PREFIX + "/" + schema.DBNotification.Name + "?" + utils.RootRowsParam + "=" + notif.GetString("id"),
						"data_ref" : "/" + utils.MAIN_PREFIX + "/" + sch.Name + "?" + utils.RootRowsParam + "=" + notif.GetString(schema.RootID("dest_table")),
					})
				}
				if err == nil { response[0]["notifications"]=n
				} else { response[0]["notifications"]=[]interface{}{} }
				l.response(response, nil)
				return
			}
		}	
		l.response(utils.Results{}, errors.New("AUTH : password invalid")) // API response
		return 
	}
	l.response(utils.Results{}, errors.New("AUTH : username/email invalid")) 
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
	l.response(utils.Results{ utils.Record { "name" : login }}, nil)
}

// @Title Refresh
// @Description User logout
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @Failure 402 user already connected
// @router /logout [get]
func (l *AuthController) Refresh() {
	login, superAdmin, err := l.authorized() // check if already connected
	if err != nil { l.response(nil, err) }
	token := l.session(login, superAdmin, false) // update session variables
	d := domain.Domain(superAdmin, login, false) 
	params := utils.Params{ utils.RootTableParam : schema.DBNotification.Name, 
						   utils.RootRowsParam : utils.ReservedParam, utils.RootRawView : "enable",}
	notifs, err := d.PermsSuperCall(params, utils.Record{}, utils.SELECT)
	resp := utils.Record{}
	n := utils.Results{}
	for _, notif := range notifs {
		sch, err := schema.GetSchemaByID(int64(notif["link_id"].(int64)))
		if err != nil { continue }
		n = append(n, utils.Record{
			utils.SpecialIDParam : notif.GetString(utils.SpecialIDParam),
			schema.NAMEKEY : notif.GetString(schema.NAMEKEY),
			"description" : notif.GetString("description"),
			"link_path" : "/" + utils.MAIN_PREFIX + "/" + schema.DBNotification.Name + "?" + utils.RootRowsParam + "=" + notif.GetString("id"),
			"data_ref" : "/" + utils.MAIN_PREFIX + "/" + sch.Name + "?" + utils.RootRowsParam + "=" + notif.GetString(schema.RootID("dest_table")),
		})
	}
	if err == nil { resp["notifications"]=n
	} else { resp["notifications"]=[]interface{}{} }
	resp["token"]=token
	l.response(utils.Results{resp}, nil)
}

var SESSIONS_KEY="user_id"
var ADMIN_KEY="super_admin"
var AUTHMODE = []string{"session", "token"}

var secret = []byte("weakest-secret")

type Token struct {}

func (t *Token) Create(user_id string, superAdmin bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		SESSIONS_KEY : user_id,
		ADMIN_KEY : superAdmin,
		"exp" : time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenStr, err := token.SignedString(secret)
	if err != nil { return "", err }
	return tokenStr, nil
}

func (t *Token) Verify(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil { return nil, err }
	if !token.Valid { return nil, fmt.Errorf("invalid token")}
	tokenStr, err = token.SignedString(secret)
	return token, err
}
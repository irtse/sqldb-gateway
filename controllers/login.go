package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sqldb-ws/models"

	"forge.redroom.link/yves/sqldb"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/lib/pq"
	"github.com/matthewhartstonge/argon2"
	"github.com/rs/zerolog/log"
)

// Operations about login
type LoginController struct {
	beego.Controller
}

type Credential struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

// @Title AddUser
// @Description Add user
// @Param	username		query	string	true		"The username for register format"
// @Param	password		query 	string	true		"The password for register"
// @Success 200
// @Failure 403 user already exist
// @router /adduser [post]
func (l *LoginController) AddUser() {

	argon := argon2.DefaultConfig()

	username := l.GetString("username")
	pass := l.GetString("password")

	hash, err := argon.HashEncoded([]byte(pass))
	if err != nil {
		log.Error().Msg(err.Error())
	}

	record := make(sqldb.AssRow)
	record["login"] = username
	record["password"] = string(hash)

	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	existing, err := db.QueryAssociativeArray("SELECT * FROM dbuser WHERE login =" + pq.QuoteLiteral(username) + ";")
	if err != nil {
		log.Error().Msg(err.Error())
	}
	if existing != nil {
		l.Ctx.Output.SetStatus(403)
	} else {
		_, err := db.Table("dbuser").Insert(record)
		if err != nil {
			log.Error().Msg(err.Error())
		}
	}

	db.Close()
}

// @Title Login
// @Description User login
// @Param	body		body 	Credential	true		"Credentials"
// @Success 200 {string} success !
// @Failure 403 user does not exist
// @Failure 402 user already connected
// @router /login [post]
func (l *LoginController) Login() {
	var creds Credential
	json.Unmarshal(l.Ctx.Input.RequestBody, &creds)

	if l.GetSession("user_id") != creds.Login {

		if creds.Login == "" || creds.Password == "" {
			l.Ctx.Output.SetStatus(403)
		}

		db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
		user, err := db.Table("dbuser").GetAssociativeArray([]string{"password"}, "login="+sqldb.Quote(creds.Login), []string{}, "")
		pass := user[0].GetString("password")
		if err != nil {
			log.Error().Msg(err.Error())
		}
		ok, err := argon2.VerifyEncoded([]byte(creds.Password), []byte(pass))
		if err != nil {
			log.Error().Msg(err.Error())
		}
		matches := "no 🔒"
		if ok {
			matches = "yes 🔓"
			username := l.GetString("username")
			l.SetSession("user_id", username)
			models.GetLogin(username)
			l.Ctx.Output.SetStatus(http.StatusOK)
			l.Data["json"] = map[string]string{"login": "ok"}

		}
		fmt.Printf("Password Matches: %s\n", matches)
		//security.Test()

	} else {
		l.Ctx.Output.SetStatus(403)
		l.Data["json"] = map[string]string{"login": "fail"}
	}

	l.ServeJSON()
}

// @Title Logout
// @Description Logs user
// @Success 200
// @Failure 403 user not exist
// @router /logout [post]
func (l *LoginController) Logout() {

	user := l.GetSession("user_id")

	if user != nil {
		l.DelSession("user_id")
	}

}

package models

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"forge.redroom.link/yves/sqldb"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

var (
	UsersRights map[string]*UserRight
	username    string
)

func GetLogin(user string) {
	username = user
}

func Init() {
	UsersRights = make(map[string]*UserRight)
	db := sqldb.Open(os.Getenv("driverdb"), os.Getenv("paramsdb"))
	query, err := db.QueryAssociativeArray("SELECT dbuser.*, dbuserrole.entity_id AS entity_id, dbuserrole.dbrole_id AS role_id, dbtableaccess.userrolerestrictions AS restrictions FROM dbuser INNER JOIN dbuserrole AS dbuserrole ON dbuser.id = dbuserrole.dbuser_id INNER JOIN dbrole ON dbuserrole.dbrole_id = dbrole.id INNER JOIN dbtableaccess ON dbtableaccess.dbrole_id = dbrole.id WHERE login =" + pq.QuoteLiteral(username) + ";")
	db.Close()
	if err != nil {
		log.Error().Msg(err.Error())
	}
	var UserId, EntityId, RoleId int
	var Restrictions, Password, Login string

	for _, element := range query.(Results) {
		for key, element := range element {
			if key == "login" {
				Login = fmt.Sprintf("%v", element)

			}
			if key == "id" {
				UserId, _ = strconv.Atoi(fmt.Sprintf("%v", element))

			}
			if key == "password" {
				Password = fmt.Sprintf("%v", element)

			}
			if key == "entity_id" {
				EntityId, _ = strconv.Atoi(fmt.Sprintf("%v", element))

			}
			if key == "role_id" {
				RoleId, _ = strconv.Atoi(fmt.Sprintf("%v", element))

			}
			if key == "restrictions" {
				Restrictions = Restrictions + fmt.Sprintf("%v", element) + "|"

			}
		}
	}
	UsersRights = make(map[string]*UserRight)
	u := UserRight{Login, UserId, RoleId, EntityId, Password, Restrictions}
	UsersRights[Login] = &u
}

type UserRight struct {
	Login        string
	UserId       int
	EntityId     int
	RoleId       int
	Password     string
	Restrictions string
}

func GetUser(uid string) (u *UserRight, err error) {
	if u, ok := UsersRights[uid]; ok {
		return u, nil
	}
	return nil, errors.New("user not exists")
}

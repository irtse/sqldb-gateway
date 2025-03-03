package domain

import (
	"fmt"
	"sqldb-ws/domain/schema"
	ds "sqldb-ws/domain/schema/database_resources"
	sm "sqldb-ws/domain/schema/models"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

func SetToken(superAdmin bool, user string, token interface{}) (utils.Results, error) {
	return Domain(superAdmin, user, false, nil).Call( // replace token by a nil
		utils.AllParams(ds.DBUser.Name), utils.Record{"token": token}, utils.UPDATE, getQueryFilter(user))
}

func IsLogged(superAdmin bool, user string, token string) (utils.Results, error) {
	domain := Domain(superAdmin, user, false, nil)
	params := utils.Params{utils.RootTableParam: ds.DBNotification.Name,
		utils.RootRowsParam: utils.ReservedParam, utils.RootRawView: "enable"}
	notifs, err := domain.SuperCall(params.RootRaw(), utils.Record{}, utils.SELECT, false)
	n := utils.Results{}
	for _, notif := range notifs {
		sch, err := schema.GetSchemaByID(int64(notif["link_id"].(float64)))
		if err != nil {
			continue
		}
		nn := utils.Record{
			utils.SpecialIDParam: notif.GetString(utils.SpecialIDParam),
			sm.NAMEKEY:           notif.GetString(sm.NAMEKEY),
			"description":        notif.GetString("description"),
			"link_path":          "/" + utils.MAIN_PREFIX + "/" + ds.DBNotification.Name + "?" + utils.RootRowsParam + "=" + notif.GetString("id"),
			"data_ref":           "@" + fmt.Sprintf("%v", sch.ID) + ":" + fmt.Sprintf("%v", notif[utils.RootDestTableIDParam]),
		}
		n = append(n, nn)
	}
	response, err := domain.Call(utils.AllParams(ds.DBUser.Name), utils.Record{}, utils.SELECT, getQueryFilter(user))
	if err != nil || len(response) == 0 {
		return nil, err
	}
	resp := response[0]
	resp["notifications"] = n
	resp["token"] = token
	return utils.Results{resp}, nil
}

func getQueryFilter(user string) string {
	return connector.FormatSQLRestrictionWhereByMap("", map[string]interface{}{
		"name":  user,
		"email": user,
	}, true)
}

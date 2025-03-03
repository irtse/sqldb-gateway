// @APIVersion 1.0.0
// @Title SqlDB WS API
// @Description Generic database access API
// @Contact yves.cerezal@irt-saintexupery.com
// @TermsOfServiceUrl https://www.irt-saintexupery.com/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"sqldb-ws/application"
	"sqldb-ws/domain/utils"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

var namespaceV1 = map[string]beego.ControllerInterface{
	"main":            &application.MainController{},
	"auth":            &application.AuthController{},
	utils.MAIN_PREFIX: &application.GenericController{},
}

func init() {
	var FilterUser = func(ctx *context.Context) {}
	v1 := []beego.LinkNamespace{}
	for key, val := range namespaceV1 {
		v1 = append(v1, beego.NSNamespace("/"+key,
			beego.NSBefore(FilterUser),
			beego.NSInclude(val),
		))
	}
	ns := beego.NewNamespace("/v1", v1...)
	beego.AddNamespace(ns)
}

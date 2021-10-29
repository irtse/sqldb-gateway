// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"net/http"
	"sqldb-ws/controllers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
)

func init() {

	var FilterUser = func(ctx *context.Context) {
		//session := ctx.Input.Session("user_id")
		_, ok := ctx.Input.Session("user_id").(string)
		if !ok {
			ctx.Output.SetStatus(http.StatusUnauthorized)
			ctx.Redirect(302, "/v1/l")
		}
	}

	ns := beego.NewNamespace("/v1",
		beego.NSNamespace("/t",
			beego.NSBefore(FilterUser),
			beego.NSInclude(
				&controllers.TableController{},
			),
		),
		beego.NSNamespace("/s",
			beego.NSInclude(
				&controllers.SchemaController{},
			),
		),
		beego.NSNamespace("/l",
			beego.NSInclude(
				&controllers.LoginController{},
			),
		),
	)
	beego.AddNamespace(ns)

	//beego.InsertFilter("/v1/t", beego.BeforeRouter, FilterUser)
}

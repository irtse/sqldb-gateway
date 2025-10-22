package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"],
        beego.ControllerComments{
            Method: "GetOK",
            Router: `/:code`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-gateway/controllers:GenericController"],
        beego.ControllerComments{
            Method: "GetMessage",
            Router: `/:code/message`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}

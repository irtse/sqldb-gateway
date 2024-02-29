package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {
    beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"],
        beego.ControllerComments{
            Method: "LogCheck",
            Router: `/logcheck`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"],
        beego.ControllerComments{
            Method: "Login",
            Router: `/login`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:AuthController"],
        beego.ControllerComments{
            Method: "Logout",
            Router: `/logout`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:MainController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:MainController"],
        beego.ControllerComments{
            Method: "Main",
            Router: `/`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Post",
            Router: `/:table`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Put",
            Router: `/:table`,
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: `/:table`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Count",
            Router: `/:table/count`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Get",
            Router: `/:table`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Importated",
            Router: `/:table/:filename/import`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "NotImportated",
            Router: `/:table/:filename/import`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    var docs Docs
    docs.GenerateDocs()
}

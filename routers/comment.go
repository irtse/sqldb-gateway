package routers

import (
    "fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {
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
            Method: "Get",
            Router: `/:table`,
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "Importated",
            Router: `/:table/import/:filename`,
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"],
        beego.ControllerComments{
            Method: "NotImportated",
            Router: `/:table/import/:filename`,
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})
    fmt.Printf("%v\n", beego.GlobalControllerRouter["sqldb-ws/controllers:GenericController"])
    var docs Docs
    docs.GenerateDocs()
}

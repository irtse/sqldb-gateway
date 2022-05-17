package routers

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context/param"
)

func init() {

    beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"],
        beego.ControllerComments{
            Method: "AddUser",
            Router: "/adduser",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"],
        beego.ControllerComments{
            Method: "Login",
            Router: "/login",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:LoginController"],
        beego.ControllerComments{
            Method: "Logout",
            Router: "/logout",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:SchemaController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:SchemaController"],
        beego.ControllerComments{
            Method: "GetTable",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:SchemaController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:SchemaController"],
        beego.ControllerComments{
            Method: "GetSchema",
            Router: "/:table",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "Put",
            Router: "/:table",
            AllowHTTPMethods: []string{"put"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "Delete",
            Router: "/:table",
            AllowHTTPMethods: []string{"delete"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "GetAllTable",
            Router: "/:table",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "TablePost",
            Router: "/:table",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "GetAllTableColumn",
            Router: "/:table/:columns",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "GetAllTableColumnRestriction",
            Router: "/:table/:columns/:restriction",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "GetAllTableColumnRestrictionSortkeys",
            Router: "/:table/:columns/:restriction/:sortkeys",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"] = append(beego.GlobalControllerRouter["sqldb-ws/controllers:TableController"],
        beego.ControllerComments{
            Method: "GetAllTableColumnRestrictionSortkeysDir",
            Router: "/:table/:columns/:restriction/:sortkeys/:dir",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}

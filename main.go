package main

import (
	"fmt"
	"os"
	_ "sqldb-gateway/routers"

	beego "github.com/beego/beego/v2/server/web"
)

func main() {
	title := "  _____  ____  _      _____  ____     __          _______ \n"
	title += " / ____|/ __ \\| |    |  __ \\|  _ \\    \\ \\        / / ____|\n"
	title += "| (___ | |  | | |    | |  | | |_) |____\\ \\  /\\  / / (___  \n"
	title += " \\___ \\| |  | | |    | |  | |  _ <______\\ \\/  \\/ / \\___ \\ \n"
	title += " ____) | |__| | |____| |__| | |_) |      \\  /\\  /  ____) |\n"
	title += "|_____/ \\___\\_\\______|_____/|____/        \\/  \\/  |_____/ \n"
	title += "														 "
	fmt.Printf("%s\n", title)
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	beego.SetStaticPath("/", "web")
	fmt.Printf("%s\n", "Service in "+os.Getenv("AUTH_MODE")+" mode")
	fmt.Printf("%s\n", "Running server...")
	beego.Run()
}

// irt-aese.local

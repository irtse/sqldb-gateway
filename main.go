package main

import (
	"fmt"
	"os"
	"plugin"
	domain "sqldb-ws/domain"
	"sqldb-ws/domain/domain_service/task"
	"sqldb-ws/domain/schema"
	_ "sqldb-ws/routers"
	"strings"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/matthewhartstonge/argon2"
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
	for key, value := range DEFAULTCONF {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	if os.Getenv("SUPERADMIN_PASSWORD") != "" {
		argon := argon2.DefaultConfig()
		hash, _ := argon.HashEncoded([]byte(os.Getenv("SUPERADMIN_PASSWORD")))
		os.Setenv("SUPERADMIN_PASSWORD", string(hash))
	}

	fmt.Printf("%s\n", "Service in "+os.Getenv("AUTH_MODE")+" mode")
	schema.Load(domain.Domain(true, os.Getenv("SUPERADMIN_NAME"), nil))
	task.Load(domain.Domain(true, os.Getenv("SUPERADMIN_NAME"), nil))
	fmt.Printf("%s\n", "Running server...")
	if os.Getenv("PLUGINS") != "" {
		for _, plug := range strings.Split(os.Getenv("PLUGINS"), ",") {
			if p, err := plugin.Open("./plugins/" + plug + "/plugin.so"); err == nil {
				if sym, err := p.Lookup("Run"); err == nil {
					launchFunc := sym.(func())
					go launchFunc()
				}
			} else {
				fmt.Println(err)
			}
		}
	}
	beego.Run()
}

var DEFAULTCONF = map[string]string{
	"SUPERADMIN_NAME":     "root",
	"SUPERADMIN_PASSWORD": "admin",
	"SUPERADMIN_EMAIL":    "morgane.roques@irt-saintexupery.com",
	"AUTH_MODE":           "token",
	"DBDRIVER":            "postgres",
	"DBHOST":              "127.0.0.1",
	"DBPORT":              "5432",
	"DBUSER":              "test",
	"DBPWD":               "test",
	"DBNAME":              "test",
	"DBSSL":               "disable",
	"log":                 "disable",
}

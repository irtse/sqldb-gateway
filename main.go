package main

import (
	"os"
	"fmt"
	_ "sqldb-ws/routers"
	"github.com/spf13/viper"
	auth "sqldb-ws/controllers"
	domain "sqldb-ws/lib/domain"
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
	viper.SetConfigName("config")          // name of config file (without extension)
	viper.AddConfigPath("/etc/sqldb-ws/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.sqldb-ws") // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig()            // Find and read the config file
	if err != nil {                        // Handle errors reading the config file
		//panic(fmt.Errorf("Fatal error config file: %w \n", err))
		viper.SetDefault("authmode", auth.AUTHMODE[0])
		viper.SetDefault("driverdb", "postgres")
		viper.SetDefault("dbhost", "127.0.0.1")
		viper.SetDefault("dbport", "5432")
		viper.SetDefault("dbuser", "test")
		viper.SetDefault("dbpwd", "test")
		viper.SetDefault("dbname", "test")
		viper.SetDefault("dbssl", "disable")
	}
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	if os.Getenv("authmode") == "" { os.Setenv("authmode", auth.AUTHMODE[0]) }
	if os.Getenv("driverdb") == "" { os.Setenv("driverdb", "postgres") }
	if os.Getenv("dbhost") == "" { os.Setenv("dbhost", "127.0.0.1") }
	if os.Getenv("dbport") == "" { os.Setenv("dbport", "5432") }
	if os.Getenv("dbuser") == "" { os.Setenv("dbuser", "test") }
	if os.Getenv("dbpwd") == "" { os.Setenv("dbpwd", "test") }
	if os.Getenv("dbname") == "" { os.Setenv("dbname", "test") }
	if os.Getenv("dbssl") == "" { os.Setenv("dbssl", "disable") }
	if os.Getenv("log") == "" { os.Setenv("log", "disable") }
	fmt.Printf("%s\n", "Checking for root DBBases... \nWait for server to launch... (may take a while on first start)")
	domain.Load()
	os.Setenv("log", "enable")
	beego.Run()
}
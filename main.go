package main

import (
	"os"
	_ "sqldb-ws/routers"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")          // name of config file (without extension)
	viper.AddConfigPath("/etc/sqldb-ws/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.sqldb-ws") // call multiple times to add many search paths
	viper.AddConfigPath(".")               // optionally look for config in the working directory
	err := viper.ReadInConfig()            // Find and read the config file
	if err != nil {                        // Handle errors reading the config file
		//panic(fmt.Errorf("Fatal error config file: %w \n", err))
		viper.SetDefault("driverdb", "postgres")
		viper.SetDefault("paramsdb", "host=127.0.0.1 port=5432 user=test password=test dbname=test sslmode=disable")
	}

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	os.Setenv("driverdb", "postgres")
	os.Setenv("paramsdb", "host=127.0.0.1 port=5432 user=test password=test dbname=test sslmode=disable")
	beego.Run()

}

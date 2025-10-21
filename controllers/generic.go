package controllers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/redis/go-redis/v9"
)

// Operations about table
type GenericController struct{ beego.Controller }

// @Title Get
// @Description get Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router /:code [get]
func (t *GenericController) GetOK() {
	code := t.Ctx.Input.Params()[":code"]
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost:6379"
	}
	fmt.Println("racac", host, code)
	var s = "false"
	path := strings.Split(t.Ctx.Input.URI(), "?")
	fmt.Println("racac1", path)
	if len(path) >= 2 {
		uri := strings.Split(path[1], "&")
		for _, val := range uri {
			kv := strings.Split(val, "=")
			if len(kv) > 1 && kv[0] == "got_response" {
				s = kv[1]
				break
			}
		}
	}
	fmt.Println("racac4", path)
	rdb := redis.NewClient(&redis.Options{
		Addr:     host, // Redis server address
		Password: "",   // no password set
		DB:       0,    // use default DB
	})
	// Save data to Redis
	if err := rdb.Set(context.Background(), code, s, 24*time.Hour).Err(); err != nil {
		fmt.Println("Could not set key: %v", err)
	}
	t.Ctx.Output.ContentType("text/html") // Optional, Beego usually handles it
	target := os.Getenv("LANG")
	fmt.Println("racac2", target, s, code)

	if target == "" {
		target = "fr"
	}
	f, err := os.ReadFile("/opt/html/index_" + target + ".html")
	if err != nil {
		t.Data["error"] = err
	}
	content := string(f)
	t.Ctx.WriteString(content)
}

// @Title Get
// @Description get Datas
// @Param	table			path 	string	true		"Name of the table"
// @Success 200 {string} success !
// @Failure 403 no table
// @router / [get]
func (t *GenericController) Get() {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     host, // Redis server address
		Password: "",   // no password set
		DB:       0,    // use default DB
	})
	var cursor uint64
	var keys []string
	for {
		var scannedKeys []string
		var err error
		scannedKeys, cursor, err = rdb.Scan(context.Background(), cursor, "*", 10).Result()
		if err != nil {
			t.Data["data"] = map[string]interface{}{
				"status": "NOT OK",
				"error":  err,
			}
			t.ServeJSON()
			return
		}
		keys = append(keys, scannedKeys...)
		if cursor == 0 {
			break
		}
	}
	data := map[string]string{}
	for _, key := range keys {
		val, err := rdb.Get(context.Background(), key).Result()
		if err != redis.Nil && err == nil {
			data[key] = val
		}
	}
	t.Data["data"] = map[string]interface{}{
		"status": "OK",
		"data":   data,
		"error":  nil,
	}
	t.ServeJSON()
}

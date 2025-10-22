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
// @router /:code/message [get]
func (t *GenericController) GetMessage() {
	code := t.Ctx.Input.Params()[":code"]
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis-server:6379"
	}
	fmt.Println("racac", host, code)
	var s = "false"
	path := strings.Split(t.Ctx.Input.URI(), "?")
	fmt.Println("racac1", path)
	if len(path) >= 2 {
		uri := strings.Split(path[1], "&")
		for _, val := range uri {
			kv := strings.Split(val, "=")
			if len(kv) > 1 && kv[0] == "message" {
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
	if err := rdb.Set(context.Background(), code+"_str", s, 24*time.Hour).Err(); err != nil {
		fmt.Println("Could not set key: %v", err)
		t.Data["json"] = map[string]interface{}{
			"status": "NOT OK",
			"error":  err,
		}
		t.ServeJSON()
		return
	}
	t.Ctx.Output.ContentType("text/html") // Optional, Beego usually handles it
	target := os.Getenv("LANG")

	if target == "" {
		target = "fr"
	}
	f, err := os.ReadFile("/opt/html/index_" + target + "_message.html")
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
// @router /:code [get]
func (t *GenericController) GetOK() {
	code := t.Ctx.Input.Params()[":code"]
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "redis-server:6379"
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
	if err := rdb.Set(context.Background(), code, s, 25*time.Hour).Err(); err != nil {
		fmt.Println("Could not set key: %v", err)
		t.Data["json"] = map[string]interface{}{
			"status": "NOT OK",
			"error":  err,
		}
		t.ServeJSON()
		return
	}
	t.Ctx.Output.ContentType("text/html") // Optional, Beego usually handles it
	target := os.Getenv("LANG")

	if target == "" {
		target = "fr"
	}
	f, err := os.ReadFile("/opt/html/index_" + target + ".html")
	if err != nil {
		t.Data["error"] = err
	}
	content := string(f)
	content = strings.ReplaceAll(content, "<code>", code)

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
		host = "redis-server:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     host, // Redis server address
		Password: "",   // no password set
		DB:       0,    // use default DB
	})
	var keys []string
	keys, err := rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		t.Data["json"] = map[string]interface{}{
			"status": "NOT OK",
			"error":  err,
		}
		t.ServeJSON()
		return
	}
	fmt.Println("KEYZ", keys)
	data := map[string]string{}
	for _, key := range keys {
		val, err := rdb.Get(context.Background(), key).Result()
		fmt.Println(key, val, err)
		if err == nil {
			data[key] = val
			err := rdb.Del(context.Background(), key).Err()
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	t.Data["json"] = map[string]interface{}{
		"status": "OK",
		"data":   data,
		"error":  nil,
	}
	t.ServeJSON()
}

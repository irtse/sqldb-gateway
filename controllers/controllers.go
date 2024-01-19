
package controllers
import (
	"os"
	"time"
	"errors"
	"strings"
	"net/http"
	"encoding/json"
	tool "sqldb-ws/lib"
	"sqldb-ws/lib/domain"
	"sqldb-ws/lib/domain/auth"
	"github.com/rs/zerolog/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matthewhartstonge/argon2"
	beego "github.com/beego/beego/v2/server/web"
	entities "sqldb-ws/lib/infrastructure/entities"
)

var JSON = "json"
var DATA = "data"
var ERROR = "error"
// Operations about table
type AbstractController struct { beego.Controller }

func (t *AbstractController) SafeCall(method tool.Method, funcName string, args... interface{}) {
	t.Call(true, method, funcName, args...)
}

func (t *AbstractController) UnSafeCall(method tool.Method, funcName string, args... interface{}) {
	t.Call(false, method, funcName, args...)
}

func (t *AbstractController) Call(auth bool, method tool.Method, funcName string, args... interface{}) {
	superAdmin := false
	var user string
	var err error
	if auth {
		user, superAdmin, err = t.authorized()
		if err != nil { t.response(tool.Results{}, err); return }
	}
	response, err := domain.Domain(false).SafeCall(superAdmin, user, t.params(), t.body(true), method, funcName, args...)
	t.response(response, err)
}

func (t *AbstractController) authorized() (string, bool, error) {
	found := false
	for _, mode := range auth.AUTHMODE {
		if mode == os.Getenv("authmode") { found = true }
	}
	if !found { return "", false, errors.New("authmode not allowed <" + os.Getenv("authmode") + ">") }
	if os.Getenv("authmode") == auth.AUTHMODE[0] {
		if t.GetSession(auth.SESSIONS_KEY) != nil { 
			return t.GetSession(auth.SESSIONS_KEY).(string), t.GetSession(auth.ADMIN_KEY).(bool), nil 
		}
		return "", false, errors.New("user not found in session")
	}
	header := t.Ctx.Request.Header
	a, ok := header["Authorization"]
	if !ok { return "", false, errors.New("No authorization in header") }
	tokenService := &auth.Token{}
	token, err := tokenService.Verify(a[0])
	if err != nil { return "", false, err }
	claims := token.Claims.(jwt.MapClaims)
	if user_id, ok := claims[auth.SESSIONS_KEY]; ok { 
		return user_id.(string), claims[auth.SESSIONS_KEY].(bool), nil 
	}
	return "", false, errors.New("user not found in token")
}

func (t *AbstractController) paramsOver(override map[string]string) map[string]string {
	params := t.params()
	for k, v := range override { params[k]=v }
	return params
}
func (t *AbstractController) params() map[string]string {
	params := map[string]string{} 
	rawParams := t.Ctx.Input.Params()
	for key, val := range rawParams {
		if strings.Contains(key, ":") && strings.Contains(key, "splat") == false {
			params[key[1:]] = val
		}
	}
	queries := []string{}
    queries = append(queries, tool.RootParams...)
	for _, val := range queries {
		name := t.Ctx.Input.Query(val)
		if name != "" { params[val] = name }
	}
	if pass, ok := params["password"]; ok { 
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(pass))
		if err != nil { log.Error().Msg(err.Error()) }
		params["password"] = string(hash)
	}
	return params
}

func (t *AbstractController) body(hashed bool) tool.Record {
	var res tool.Record 
	json.Unmarshal(t.Ctx.Input.RequestBody, &res)
	if pass, ok := res["password"]; ok { 
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(pass.(string)))
		if err != nil { log.Error().Msg(err.Error()) }
		if hashed { res["password"] = string(hash) }
	}
	return res
}

func (t *AbstractController) response(resp tool.Results, err error) { 
	t.Ctx.Output.SetStatus(http.StatusOK)
	if err != nil {
		if strings.Contains(err.Error(), "AUTH") { t.Ctx.Output.SetStatus(http.StatusUnauthorized) }
		log.Error().Msg(err.Error())
		t.Data[JSON]=map[string]interface{}{ DATA : tool.Results{}, ERROR : err.Error() }
	} else { 
		if len(resp) == 0 { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : "Datas not found" } 			
		} else { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : nil } }
		for _, json := range t.Data[JSON].(map[string]interface{})[DATA].(tool.Results) {
			if _, ok := json["password"]; ok { delete(json, "password") }
		}
	}
	t.ServeJSON()
}

func (t *AbstractController) session(userId string, superAdmin bool, delete bool) {
	delFunc := func() { 
		if os.Getenv("authmode") == auth.AUTHMODE[0] {
			if t.GetSession(auth.SESSIONS_KEY) != nil { 
				t.DelSession(auth.SESSIONS_KEY) 
				t.DelSession(auth.ADMIN_KEY) 
			} 
		} else {
			params := t.paramsOver(map[string]string{ tool.RootTableParam : entities.DBUser.Name, 
				                                      tool.RootRowsParam : tool.ReservedParam, 
													  "login" : userId })
			domain.Domain(false).UnSafeCall(
				userId,
				params, 
				tool.Record{ "token" : nil }, 
				tool.UPDATE, 
				"CreateOrUpdate",
			)
		}
	}
	if delete { delFunc(); return }
	if os.Getenv("authmode") == auth.AUTHMODE[0] { 
		t.SetSession(auth.SESSIONS_KEY, userId) 
		t.SetSession(auth.ADMIN_KEY, superAdmin)
	} else { 
		tokenService := &auth.Token{}
		token, err := tokenService.Create(userId, superAdmin)
		if err != nil { t.response(tool.Results{}, err); return }
		params := t.paramsOver(map[string]string{ tool.RootTableParam : entities.DBUser.Name,
			                                      tool.RootRowsParam : tool.ReservedParam, 
												  "login" : userId })
		domain.Domain(false).UnSafeCall(
			userId,
			params, 
			tool.Record{ "token" : token }, 
			tool.UPDATE, 
			"CreateOrUpdate",
		)
	}
	timer := time.AfterFunc(time.Hour * 24, delFunc)
	defer timer.Stop()
}

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
	entities "sqldb-ws/lib/entities"
)
/*
	AbstractController defines main procedure that a generic Handler would get. 
*/
var JSON = "json"
var DATA = "data"
var ERROR = "error"
// Operations about table
type AbstractController struct { 
	paramsOverload map[string]string
	beego.Controller 
}
// SafeCaller will ask for a authenticated procedure
func (t *AbstractController) SafeCall(method tool.Method, funcName string, args... interface{}) {
	t.Call(true, method, funcName, args...)
}
// SafeCaller will ask for a free of authentication procedure
func (t *AbstractController) UnSafeCall(method tool.Method, funcName string, args... interface{}) {
	t.Call(false, method, funcName, args...)
}
// Call function invoke Domain service and ask for the proper function by function name & method
func (t *AbstractController) Call(auth bool, method tool.Method, funcName string, args... interface{}) {
	superAdmin := false
	var user string
	var err error
	if auth { // we will verify authentication and status if auth is expected
		user, superAdmin, err = t.authorized() // will check up if allowed (or authenticated)
		if err != nil { t.response(tool.Results{}, err); return }
	} // then proceed to exec by calling domain
	response, err := domain.Domain(superAdmin, user, false).Call(t.params(), t.body(true), method, true, funcName, args...)
	t.response(response, err) // send back response
}
// authorized is authentication check up func of the HANDLER
func (t *AbstractController) authorized() (string, bool, error) {
	found := false
	for _, mode := range auth.AUTHMODE { // above all check for kind of auth (token in authorization header, or session API)
		if mode == os.Getenv("authmode") { found = true }
	} // if none found give an error
	if !found { return "", false, errors.New("authmode not allowed <" + os.Getenv("authmode") + ">") }
	// session auth will look in session variables in API ONLY
	if os.Getenv("authmode") == auth.AUTHMODE[0] {
		if t.GetSession(auth.SESSIONS_KEY) != nil { 
			return t.GetSession(auth.SESSIONS_KEY).(string), t.GetSession(auth.ADMIN_KEY).(bool), nil 
		}
		return "", false, errors.New("user not found in session")
	} // TOKEN verification is a little bit verbose by extractin token in Authorization Header and look after its properties
	header := t.Ctx.Request.Header
	a, ok := header["Authorization"] // extract token in HEADER
	if !ok { return "", false, errors.New("No authorization in header") }
	tokenService := &auth.Token{}
	token, err := tokenService.Verify(a[0]) // Verify if token is valid
	if err != nil { return "", false, err }
	claims := token.Claims.(jwt.MapClaims)
	if user_id, ok := claims[auth.SESSIONS_KEY]; ok { // if all in claims send back super mode and user as confirmation
		return user_id.(string), claims[auth.ADMIN_KEY].(bool), nil 
	}
	return "", false, errors.New("user not found in token")
}
// paramsOver is an overide of params applying a manual addition of parameters into Params struct
func (t *AbstractController) paramsOver(override map[string]string) map[string]string {
	params := t.params() // get initial params
	for k, v := range override { params[k]=v } // add custom params
	return params
}
// params will produce a Params struct compose of url & query parameters
func (t *AbstractController) params() map[string]string {
	if t.paramsOverload != nil { return t.paramsOverload }
	params := map[string]string{} 
	rawParams := t.Ctx.Input.Params() // extract all params from url and fill params
	for key, val := range rawParams {
		if strings.Contains(key, ":") && strings.Contains(key, "splat") == false {
			params[key[1:]] = val
		}
	}
	queries := []string{} // then we will extract query parameters
    queries = append(queries, tool.RootParams...) // firstival we will try to found pertinent query params
	queries = append(queries, tool.HiddenParams...)
	if tablename, ok := params[tool.RootTableParam]; ok { // retrieve schema
		params := tool.Params{ tool.RootTableParam : tablename, }
		d := domain.Domain(true, "", false) // create a new domain with current permissions of user
		d.Specialization = false // when launching call disable every auth check up (don't forget you are not logged)
		response, err := d.Call(params, tool.Record{}, tool.SELECT, false, "Get")
		if cols, ok2 := response[0]["columns"]; ok2 && err == nil {
			for colName, _ := range cols.(map[string]entities.TableColumnEntity) {
				queries = append(queries, colName)
			}
		}
	}

	for _, val := range queries {
		name := t.Ctx.Input.Query(val)
		if name != "" { params[val] = name }
	} // GET SCHEMA PARAMETERS
	if pass, ok := params["password"]; ok { // if any password founded hash it
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(pass))
		if err != nil { log.Error().Msg(err.Error()) }
		params["password"] = string(hash)
	}
	return params
}
// body is the main body extracter from the controller
func (t *AbstractController) body(hashed bool) tool.Record {
	var res tool.Record 
	json.Unmarshal(t.Ctx.Input.RequestBody, &res)
	if pass, ok := res["password"]; ok { // if any password founded hash it
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(pass.(string)))
		if err != nil { log.Error().Msg(err.Error()) }
		if hashed { res["password"] = string(hash) }
	}
	return res
}
// response rules every http response 
func (t *AbstractController) response(resp tool.Results, err error) { 
	t.Ctx.Output.SetStatus(http.StatusOK) // defaulting on absolute success
	if err != nil { // Check nature of error if there is one
		if strings.Contains(err.Error(), "AUTH") { t.Ctx.Output.SetStatus(http.StatusUnauthorized) }
		if strings.Contains(err.Error(), "partial") { 
			t.Ctx.Output.SetStatus(http.StatusPartialContent) 
			t.Data[JSON]=map[string]interface{}{ DATA : resp, ERROR : err.Error() }
		} else {
			log.Error().Msg(err.Error())
			t.Data[JSON]=map[string]interface{}{ DATA : tool.Results{}, ERROR : err.Error() }
		}
	} else { // if success precise an error if no datas is founded
		if len(resp) == 0 { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : "datas not found" } 			
		} else { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : nil } }
		for _, json := range t.Data[JSON].(map[string]interface{})[DATA].(tool.Results) {
			if _, ok := json["password"]; ok { delete(json, "password") } // never send back a password in any manner
		}
	}
	t.ServeJSON() // then serve response by beego
}
// session is the main manager from Handler
func (t *AbstractController) session(userId string, superAdmin bool, delete bool) {
	delFunc := func() { // set up a lambda call back function to delete in session and token in base if needed
		if t.GetSession(auth.SESSIONS_KEY) != nil { 
			t.DelSession(auth.SESSIONS_KEY) // user_id key
			t.DelSession(auth.ADMIN_KEY) // super_admin key
		} 
		if os.Getenv("authmode") != auth.AUTHMODE[0] { // in case of token way of authenticate
			params := t.paramsOver(map[string]string{ tool.RootTableParam : entities.DBUser.Name, 
				                                      tool.RootRowsParam : tool.ReservedParam, 
													  "name" : userId })
			domain.Domain(false, userId, false).Call( // replace token by a nil
				params, 
				tool.Record{ "token" : nil }, 
				tool.UPDATE, 
				false,
				"CreateOrUpdate",
			)
		}
	}
	if delete { delFunc(); return } // if only deletion quit after launching lambda
	t.SetSession(auth.SESSIONS_KEY, userId) // load superadmin and user id in session in any case
	t.SetSession(auth.ADMIN_KEY, superAdmin)
	if os.Getenv("authmode") != auth.AUTHMODE[0] { // if token way of authentication
		tokenService := &auth.Token{} // generate a new token with all needed claims
		token, err := tokenService.Create(userId, superAdmin); 
		if err != nil { t.response(tool.Results{}, err); return } // then update user with its brand new token.
		params := t.paramsOver(map[string]string{ tool.RootTableParam : entities.DBUser.Name,
			                                      tool.RootRowsParam : tool.ReservedParam, 
												  "name" : userId })
		domain.Domain(false, userId, false).Call(
			params, 
			tool.Record{ "token" : token }, 
			tool.UPDATE, 
			false,
			"CreateOrUpdate",
		)
	} // launch a 24h session timer after this session will be killed.
	timer := time.AfterFunc(time.Hour * 24, delFunc)
	defer timer.Stop()
}
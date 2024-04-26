
package controllers
import (
	"os"
	"fmt"
	"time"
	"errors"
	"strings"
	"net/http"
	"encoding/csv"
	"encoding/json"
	"sqldb-ws/lib/domain"
	"sqldb-ws/lib/domain/utils"
	"github.com/rs/zerolog/log"
	"sqldb-ws/lib/domain/schema"
	"github.com/golang-jwt/jwt/v5"
	"github.com/thedatashed/xlsxreader"
	"github.com/matthewhartstonge/argon2"
	beego "github.com/beego/beego/v2/server/web"
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
func (t *AbstractController) SafeCall(method utils.Method, args... interface{}) {
	t.Call(true, method, args...)
}
// SafeCaller will ask for a free of authentication procedure
func (t *AbstractController) UnSafeCall(method utils.Method,  args... interface{}) {
	t.Call(false, method, args...)
}
// Call function invoke Domain service and ask for the proper function by function name & method
func (t *AbstractController) Call(auth bool, method utils.Method, args... interface{}) {
	superAdmin := false
	var user string
	var err error
	if auth { // we will verify authentication and status if auth is expected
		user, superAdmin, err = t.authorized() // will check up if allowed (or authenticated)
		if err != nil { t.response(utils.Results{}, err); return }
	} // then proceed to exec by calling domain
	d := domain.Domain(superAdmin, user, false)
	d.SetExternalSuperAdmin(superAdmin)
	p, asLabel := t.params()
	files, err := t.formFile(asLabel)
	if err == nil && len(files) > 0 {
		resp := utils.Results{}; var error error
		for _, file := range files {
			response, err := d.Call(p, file, method, args...)
			resp = append(resp, response...)
			if err != nil { 
				if error == nil { error = err } else { error = errors.New(error.Error() + " | " + err.Error()) }
			}
		}
		t.response(resp, error) // send back response
	} else {
		response, err := d.Call(p, t.body(true), method, args...)
		if format, ok := p[utils.RootExport]; ok { t.download(format, p[utils.RootFilename], asLabel, response, err); return }
		t.response(response, err) // send back response
	}	
}
// authorized is authentication check up func of the HANDLER
func (t *AbstractController) authorized() (string, bool, error) {
	found := false
	for _, mode := range AUTHMODE { // above all check for kind of auth (token in authorization header, or session API)
		if mode == os.Getenv("authmode") { found = true }
	} // if none found give an error
	if !found { return "", false, errors.New("authmode not allowed <" + os.Getenv("authmode") + ">") }
	// session auth will look in session variables in API ONLY
	if os.Getenv("authmode") == AUTHMODE[0] {
		if t.GetSession(SESSIONS_KEY) != nil { 
			return t.GetSession(SESSIONS_KEY).(string), t.GetSession(ADMIN_KEY).(bool), nil 
		}
		return "", false, errors.New("user not found in session")
	} // TOKEN verification is a little bit verbose by extractin token in Authorization Header and look after its properties
	header := t.Ctx.Request.Header
	a, ok := header["Authorization"] // extract token in HEADER
	if !ok { return "", false, errors.New("No authorization in header") }
	tokenService := &Token{}
	token, err := tokenService.Verify(a[0]) // Verify if token is valid
	if err != nil { return "", false, err }
	claims := token.Claims.(jwt.MapClaims)
	if user_id, ok := claims[SESSIONS_KEY]; ok { // if all in claims send back super mode and user as confirmation
		return user_id.(string), claims[ADMIN_KEY].(bool), nil 
	}
	return "", false, errors.New("user not found in token")
}
// paramsOver is an overide of params applying a manual addition of parameters into Params struct
func (t *AbstractController) paramsOver(override map[string]string) map[string]string {
	params, _ := t.params() // get initial params
	for k, v := range override { params[k]=v } // add custom params
	return params
}
func (t *AbstractController) formFile(asLabel map[string]string) (utils.Results, error) {
	file, header, err := t.Ctx.Request.FormFile("file")
	if err == nil {
		defer file.Close()
		cols := []string{}
		results := utils.Results{}
		if strings.Contains(header.Filename, ".csv") {
			reader := csv.NewReader(file) // check if file is a CSV
			records, err := reader.ReadAll() 
			if err != nil || len(records) == 1 { return nil, err }
			cols = records[0]
			for _, col := range cols {
				for k, label := range asLabel {
					if col == label { col = strings.Replace(k, "_aslabel", "", -1) }
				}
			}
			for _, rec := range records[1:] {
				newRecord := utils.Record{}
				for i, col := range cols { newRecord[col] = rec[i] }
				results = append(results, newRecord)
			}
			return results, nil
		} else if strings.Contains(header.Filename, ".xlsx") {
			b := []byte{}
			_, err := file.Read(b)
			if err != nil { return nil, err }
			xl, err := xlsxreader.NewReader(b)
			if err != nil { return nil, err }
			first := true
			for row := range xl.ReadRows(xl.Sheets[0]){
				if first {
					first = false
					for i, cell := range row.Cells {
						for k, label := range asLabel {
							if cell.Value == label { cols = append(cols, strings.Replace(k, "_aslabel", "", -1)) }
						}
						if len(cols) < i + 1 { cols = append(cols, cell.Value) }
					}
				} else {
					newRecord := utils.Record{}
					for i, cell := range row.Cells { newRecord[cols[i]] = cell.Value }
					results = append(results, newRecord)
				}
			}
			return results, nil
		}
	}
	return nil, err
}
// params will produce a Params struct compose of url & query parameters
func (t *AbstractController) params() (map[string]string, map[string]string) {
	paramsAsLabel := map[string]string{}
	if t.paramsOverload != nil { return t.paramsOverload, paramsAsLabel }
	params := map[string]string{} 
	rawParams := t.Ctx.Input.Params() // extract all params from url and fill params
	for key, val := range rawParams {
		if strings.Contains(key, ":") && strings.Contains(key, "splat") == false {
			params[key[1:]] = val
		}
	}
	path := strings.Split(t.Ctx.Input.URI(), "?")
	if len(path) >= 2 {
		uri := strings.Split(path[1], "&")
		for _, val := range uri {
			kv := strings.Split(val, "=")
			if strings.Contains(kv[0], "_aslabel") { paramsAsLabel[kv[0]]=kv[1]
			} else { params[kv[0]]=kv[1] }
		}
	}
	if pass, ok := params["password"]; ok { // if any password founded hash it
		argon := argon2.DefaultConfig()
		hash, err := argon.HashEncoded([]byte(pass))
		if err != nil { log.Error().Msg(err.Error()) }
		params["password"] = string(hash)
	}
	if t, ok := params[utils.RootTableParam]; !ok || t == schema.DBView.Name { delete(params, utils.RootExport) }
	if _, ok := params[utils.RootExport]; ok { 
		params[utils.RootRawView] = "" 
		if _, ok := params[utils.RootFilename]; !ok { params[utils.RootFilename] = params[utils.RootTableParam] }
	}
	return params, paramsAsLabel
}
// body is the main body extracter from the controller
func (t *AbstractController) body(hashed bool) utils.Record {
	var res utils.Record 
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
func (t *AbstractController) response(resp utils.Results, err error) { 
	t.Ctx.Output.SetStatus(http.StatusOK) // defaulting on absolute success
	if err != nil { // Check nature of error if there is one
		//if strings.Contains(err.Error(), "AUTH") { t.Ctx.Output.SetStatus(http.StatusUnauthorized) }
		if strings.Contains(err.Error(), "partial") { 
			t.Ctx.Output.SetStatus(http.StatusPartialContent) 
			t.Data[JSON]=map[string]interface{}{ DATA : resp, ERROR : err.Error() }
		} else {
			log.Error().Msg(err.Error())
			t.Data[JSON]=map[string]interface{}{ DATA : utils.Results{}, ERROR : err.Error() }
		}
	} else { // if success precise an error if no datas is founded
		if len(resp) == 0 { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : "datas not found" } 			
		} else { t.Data[JSON] = map[string]interface{}{ DATA : resp, ERROR : nil } }
		for _, json := range t.Data[JSON].(map[string]interface{})[DATA].(utils.Results) {
			if _, ok := json["password"]; ok { delete(json, "password") } // never send back a password in any manner
		}
	}
	t.ServeJSON() // then serve response by beego
}

func (t *AbstractController) download(format string, name string, mapping map[string]string, resp utils.Results, error error) {
	cols, results := t.mapping(mapping, resp) // mapping
	t.Ctx.ResponseWriter.Header().Set("Content-Type", "text/" + format)
	t.Ctx.ResponseWriter.Header().Set("Content-Disposition", "attachment; filename=" + name + "_" + strings.Replace(time.Now().Format(time.RFC3339), " ", "_", -1) + "." + format)
	data := t.csv(cols, results) // rationalize to CSV
	if format == "csv" { 
		csv.NewWriter(t.Ctx.ResponseWriter).WriteAll(data) 
	} else { 
		t.response(results, error)
	}
}
func (t *AbstractController) mapping(mapping map[string]string, resp utils.Results) ([]string, utils.Results) {
	cols := []string{}; results := utils.Results{}
	if len(resp) == 0 { return cols, results }
	r := resp[0]
	order := r["order"].([]interface{})
	schema := r["schema"].(map[string]interface{})
	for _, o := range order {
		key := o.(string)
		if scheme, ok := schema[o.(string)]; !ok && strings.Contains(scheme.(map[string]interface{})["type"].(string), "many") { continue }
		label := strings.Replace(schema[key].(map[string]interface{})["label"].(string), "_", " ", -1)
		if mapKey, ok := mapping[key]; ok && mapKey != "" { label = mapKey }	
		cols = append(cols, label)
	}
	for _, item := range r["items"].([]interface{}) {
		record := utils.Record{}
		for _, o := range order {
			key := o.(string)
			it := item.(map[string]interface{})
			if scheme, ok := schema[key]; !ok && strings.Contains(scheme.(map[string]interface{})["type"].(string), "many") { continue }
			label := strings.Replace(schema[key].(map[string]interface{})["label"].(string), "_", " ", -1)
			if mapKey, ok := mapping[key]; ok && mapKey != "" { label = mapKey }	
			label = label
			if v, ok := it["values_shallow"].(map[string]interface{})[key]; ok { record[label] = v.(map[string]interface{})["name"].(string)
			} else if v, ok := it["values"].(map[string]interface{})[key]; ok && v != nil { record[label] = fmt.Sprintf("%v", v) 
			} else { record[label] = "" }
		}
		results = append(results, record)
	}
	return cols, results
}
func (t *AbstractController) csv(cols []string, results utils.Results) [][]string {
	var data [][]string
	data = append(data, cols)
	for _, r := range results {
		var row []string
		for _, c := range cols { 
			if v, ok := r[c]; !ok || v == nil  { row = append(row, ""); continue }
			row = append(row, fmt.Sprintf("%v", r[c])) 
		}
		data = append(data, row)
	}
	return data
}
// session is the main manager from Handler
func (t *AbstractController) session(userId string, superAdmin bool, delete bool) string {
	var err error
	token := ""
	delFunc := func() { // set up a lambda call back function to delete in session and token in base if needed
		if t.GetSession(SESSIONS_KEY) != nil { 
			t.DelSession(SESSIONS_KEY) // user_id key
			t.DelSession(ADMIN_KEY) // super_admin key
		} 
		if os.Getenv("authmode") != AUTHMODE[0] { // in case of token way of authenticate
			params := t.paramsOver(map[string]string{ utils.RootTableParam : schema.DBUser.Name, 
				                                      utils.RootRowsParam : utils.ReservedParam, 
													})
			sqlFilter := "name='" + userId + "' OR email='" + userId + "'"
			domain.Domain(false, userId, false).Call( // replace token by a nil
				params, utils.Record{ "token" : nil }, utils.UPDATE, sqlFilter)
		}
	}
	if delete { delFunc(); return token } // if only deletion quit after launching lambda
	t.SetSession(SESSIONS_KEY, userId) // load superadmin and user id in session in any case
	t.SetSession(ADMIN_KEY, superAdmin)
	if os.Getenv("authmode") != AUTHMODE[0] { // if token way of authentication
		tokenService := &Token{} // generate a new token with all needed claims
		token, err = tokenService.Create(userId, superAdmin); 
		if err != nil { t.response(utils.Results{}, err); return token } // then update user with its brand new token.
		params := t.paramsOver(map[string]string{ utils.RootTableParam : schema.DBUser.Name,
			                                      utils.RootRowsParam : utils.ReservedParam, })
		sqlFilter := "name='" + userId + "' OR email='" + userId + "'"
		_, err = domain.Domain(false, userId, false).Call(
			params, 
			utils.Record{ "token" : token }, 
			utils.UPDATE, 
			sqlFilter,
		)
	} // launch a 24h session timer after this session will be killed.
	timer := time.AfterFunc(time.Hour * 24, delFunc)
	defer timer.Stop()
	return token
}
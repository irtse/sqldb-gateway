package controller

import (
	"encoding/csv"
	"net/http"
	"net/url"
	"sqldb-ws/domain/utils"
	"strings"
	"time"
)

func (t *AbstractController) Respond(params map[string]string, asLabel map[string]string, method utils.Method, domain utils.DomainITF, args ...interface{}) {
	if _, ok := params[utils.RootExport]; ok {
		params[utils.RootRawView] = "disable"
	}
	response, err := domain.Call(params, t.Body(true), method, args...)
	if format, ok := params[utils.RootExport]; ok {
		var cols, cmd, cmdCols string = "", "", ""
		if pp, ok := params[utils.RootColumnsParam]; ok {
			cols = pp
		}
		if pp, ok := params[utils.RootCommandRow]; ok {
			cmd = pp
		}
		if pp, ok := params[utils.RootCommandCols]; ok {
			cmdCols = pp
		}
		t.download(domain, cols, cmdCols, cmd, format, params[utils.RootFilename], asLabel, response, err)
		return
	}
	t.Response(response, err) // send back response
}

// response rules every http response
func (t *AbstractController) Response(resp utils.Results, err error) {
	t.Ctx.Output.SetStatus(http.StatusOK) // defaulting on absolute success
	if err != nil {                       // Check nature of error if there is one
		//if strings.Contains(err.Error(), "AUTH") { t.Ctx.Output.SetStatus(http.StatusUnauthorized) }
		if strings.Contains(err.Error(), "partial") {
			t.Ctx.Output.SetStatus(http.StatusPartialContent)
			t.Data[JSON] = map[string]interface{}{DATA: resp, ERROR: err.Error()}
		} else {
			t.Data[JSON] = map[string]interface{}{DATA: utils.Results{}, ERROR: err.Error()}
		}
	} else { // if success precise an error if no datas is founded
		t.Data[JSON] = map[string]interface{}{DATA: resp, ERROR: nil}
		for _, json := range utils.ToMap(t.Data[JSON])[DATA].(utils.Results) {
			delete(json, "password") // never send back a password in any manner
		}
	}
	t.ServeJSON() // then serve response by beego
}

func (t *AbstractController) download(d utils.DomainITF, col string, colsCmd string, cmd string, format string, name string, mapping map[string]string, resp utils.Results, error error) {
	cols, lastLineMap, results := t.mapping(col, colsCmd, cmd, mapping, resp) // mapping
	t.Ctx.ResponseWriter.Header().Set("Content-Type", "text/"+format)
	t.Ctx.ResponseWriter.Header().Set("Content-Disposition", "attachment; filename="+name+"_"+strings.Replace(time.Now().Format(time.RFC3339), " ", "_", -1)+"."+format)
	data := t.csv(d, lastLineMap, cols, results) // rationalize to CSV
	if format == "csv" {
		csv.NewWriter(t.Ctx.ResponseWriter).WriteAll(data)
	} else {
		t.Response(results, error)
	}
}

func (t *AbstractController) csv(d utils.DomainITF, colsFunc map[string]string, cols []string, results utils.Results) [][]string {
	var data [][]string
	data = append(data, cols)
	lastLine := []string{}
	for _, c := range cols {
		if v, ok := colsFunc[c]; ok && v != "" {
			r, err := d.GetDb().QueryAssociativeArray("SELECT " + v + " as result FROM " + d.GetTable() + " WHERE " + d.GetDb().SQLRestriction)
			if err == nil && len(r) > 0 {
				splitted := strings.Split(v, "(")
				lastLine = append(lastLine, splitted[0]+": "+utils.GetString(r[0], "result"))
			}
		} else {
			lastLine = append(lastLine, "")
		}
	}
	for _, r := range results {
		var row []string
		for _, c := range cols {
			if v, ok := r[c]; !ok || v == nil {
				row = append(row, "")
				continue
			}
			row = append(row, utils.ToString(r[c]))
		}
		data = append(data, row)
	}
	data = append(data, lastLine)
	return data
}

func (t *AbstractController) mapping(col string, colsCmd string, cmd string, mapping map[string]string, resp utils.Results) ([]string, map[string]string, utils.Results) {
	cols := []string{}
	results := utils.Results{}
	colsFunc := map[string]string{}
	if len(resp) == 0 {
		return cols, colsFunc, results
	}
	r := resp[0]
	additionnalCol := ""
	order := []interface{}{"id"}
	order = append(order, utils.ToList(r["order"])...)
	if cmd != "" {
		decodedLine, _ := url.QueryUnescape(cmd)
		re := strings.Split(decodedLine, " as ")
		if len(re) > 1 {
			additionnalCol = re[len(re)-1]
			order = append(order, additionnalCol)
			colsFunc[additionnalCol] = re[0]
		}
	}

	for _, c := range strings.Split(colsCmd, ",") {
		re := strings.Split(c, ":")
		if len(re) > 1 {
			if v, ok := colsFunc[re[0]]; ok && v != "" {
				colsFunc[re[0]] = strings.ToUpper(re[1]) + "((" + v + "))"
			} else {
				colsFunc[re[0]] = re[1]
			}
		}
	}
	schema := utils.ToMap(r["schema"])
	for _, o := range order {
		key := utils.ToString(o)
		if col != "" && !strings.Contains(col, key) && !(additionnalCol == "" || strings.Contains(additionnalCol, key)) {
			continue
		}
		if scheme, ok := schema[key]; ok && strings.Contains(utils.ToString(utils.ToMap(scheme)["type"]), "many") {
			continue
		}
		label := key
		if scheme, ok := schema[key]; ok {
			label = strings.Replace(utils.ToString(utils.ToMap(scheme)["label"]), "_", " ", -1)
		}
		if mapKey, ok := mapping[key]; ok && mapKey != "" {
			label = mapKey
		}
		cols = append(cols, label)
	}
	for _, item := range utils.ToList(r["items"]) {
		record := utils.Record{}
		for _, o := range order {
			key := utils.ToString(o)
			it := utils.ToMap(item)
			if scheme, ok := schema[key]; ok && key != "id" && strings.Contains(
				utils.ToString(utils.ToMap(scheme)["type"]), "many") {
				continue
			}
			label := key
			if scheme, ok := schema[key]; ok {
				label = strings.Replace(utils.ToString(utils.ToMap(scheme)["label"]), "_", " ", -1)
			}
			if mapKey, ok := mapping[key]; ok && mapKey != "" {
				label = mapKey
			}
			if v, ok := utils.ToMap(it["values_shallow"])[key]; ok {
				record[label] = utils.ToString(utils.ToMap(v)["name"])
			} else if v, ok := utils.ToMap(it["values"])[key]; ok && v != nil {
				record[label] = utils.ToString(v)
			} else {
				record[label] = ""
			}
			colsFunc[label] = colsFunc[key]
		}
		results = append(results, record)
	}
	return cols, colsFunc, results
}

package routers

import (
	"os"
	"fmt"
	"slices"
	"strings"
	"reflect"
	"encoding/json"
	tool "sqldb-ws/lib"

	"github.com/ghodss/yaml"
	beego "github.com/beego/beego/v2/server/web"
)

type Docs struct {}

func (d *Docs) GenerateDocs() {
	data, err := json.MarshalIndent(d.generateDocsOnRun(), "", " ")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	file, err := os.Create("swagger/swagger.json")
    if err != nil { return }
    defer file.Close()
	file.Write(data)

	yml, err := yaml.JSONToYAML(data)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}
	file, err = os.Create("swagger/swagger.yml")
    if err != nil { return }
    defer file.Close()
	file.Write(yml)
}

func (d *Docs) generateDocsOnRun() map[string]interface{} {
	documents := make(map[string]interface{})
	for key, controller := range beego.GlobalControllerRouter {
        for _, entry := range controller {
            controller := strings.Split(key, ":")
			parameters := []string{}
            queries := []string{}
			if strings.Contains(strings.ToLower(key), "generic") { 
				queries = append(queries, tool.RootParams...) 
			}
            paths := strings.Split(entry.Router, "/")
			tag := ""
			for route, obj := range namespaceV1 {
				if strings.Contains(key, string(getType(obj)[1:])) { tag = route }	
			}
			path := ""
			if tag != "" { path = "/" + tag }
            for _, p := range paths {
				if p != ""  {
					if strings.Contains(p, ":") && slices.Contains(parameters, p[1:]) == false { 
						path += "/{" + p[1:] + "}"
						parameters = append(parameters, p[1:])
					} else { 
						if slices.Contains(parameters, p) == false {
							path += "/" + p 
						}
					}
				}
            }
			if _, ok := documents[path]; ok == false {
				documents[path] = make(map[string]interface{})
			}
			pars := []map[string]interface{}{}
            for _, param := range parameters {
                pars= append(pars, map[string]interface{}{
                    "in": "path",
                    "name": param,
                    "description": "Value of the " + param,
                    "required": true,
                    "type": "string",
                })
            }
			alreadySet := []string{}
			for _, query := range queries {
				if slices.Contains(alreadySet, query) == false {
					alreadySet = append(alreadySet, query)
					sets := map[string]interface{}{
						"in": "query",
						"name": query,
						"description": "(Optionnal) Value of " + query,
						"type": "string",
					}
					if desc, ok := tool.RootParamsDesc[query]; ok {
						sets["description"]=desc
					}
					pars= append(pars, sets)
				}  
            }
			pars=append(pars, map[string]interface{}{
				"in": "header",
				"name": "Authorization",
				"description": "Authorization Token HEADER ",
				"type": "string",
			})
			if entry.AllowHTTPMethods[0] == "put" || entry.AllowHTTPMethods[0] == "post" {
				pars=append(pars, map[string]interface{}{
                    "in": "body",
                    "name": "data",
                    "description": "Request body ",
                    "required": true,
                    "schema": map[string]string{
						"$ref": "#/definitions/json",
					},
                })
			}
			d := map[string]interface{}{}
			if d2, ok := documents[path]; ok { d = d2.(map[string]interface{})
			} else { d = documents[path].(map[string]interface{}) }
			d[strings.ToLower(entry.AllowHTTPMethods[0])] = map[string]interface{}{
                "description": entry.Method + " Datas\n\u003cbr\u003e",
				"tags": []string{tag},
                "operationId": controller[len(controller) - 1] + "." + entry.Method,
                "parameters": pars,
                "responses": map[string]interface{}{
                    "200": map[string]string{
                        "description": "{string} success !",
                    },
                    "403": map[string]string{
                        "description": "no table",
                    },
                },
            }
        }
    }
	docs := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{} {
			"title": "SqlDB WS API",
			"description": "Generic database access API\n",
			"version": "1.0.0",
			"termsOfService": "https://www.irt-saintexupery.com/",
			"contact": map[string]interface{}{
				"email": "yves.cerezal@irt-saintexupery.com",
			},
			"license": map[string]interface{}{
				"name": "Apache 2.0",
				"url": "http://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		"basePath": "/v1",
		"definitions": map[string]interface{}{
			"json": map[string]interface{}{
				"title": "json",
				"type": "object",
			},
		},
		"paths": documents,
	}
	return docs
}

func getType(myvar interface{}) string {
	if t := reflect.TypeOf(myvar); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}
// compile with: go build -buildmode=plugin -o plugin.so plugin.go

// plugin.go
package main

import (
	"encoding/csv"
	"slices"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/infrastructure/connector"
	"strings"

	"fmt"
	"os"
	"sqldb-ws/domain"
	"sqldb-ws/domain/utils"
	"time"

	models "sqldb-ws/plugins/datas"
)

func Run() {
	for {
		ImportUserHierachy()
		ImportProjectAxis()
		time.Sleep(24 * time.Hour)
	}
}

func ImportProjectAxis() {
	mapped := map[string]string{
		"Code Axe":     "code",
		"Libellé Axe":  "name",
		"Code Domaine": "domain_code",
	}
	d := domain.Domain(true, os.Getenv("SUPERADMIN_NAME"), nil)
	filepath := os.Getenv("PROJECT_FILE_PATH")
	if filepath == "" {
		filepath = "./project_test.csv"
	}
	headers, datas := importFile(filepath)
	inside := []string{}
	for _, data := range datas {
		record := map[string]interface{}{}
		for i, header := range headers {
			if realLabel, ok := mapped[header]; ok && realLabel != "" && data[i] != "" {
				record[realLabel] = data[i]
			}
		}
		if len(record) == 3 && !slices.Contains(inside, utils.GetString(record, "name")) {
			inside = append(inside, utils.GetString(record, "name"))
			if res, err := d.GetDb().SelectQueryWithRestriction(models.Axis.Name, map[string]interface{}{
				"code": connector.Quote(utils.GetString(record, "code")),
			}, false); err == nil && len(res) > 0 {
				record[utils.SpecialIDParam] = res[0][utils.SpecialIDParam]
				d.UpdateSuperCall(utils.GetRowTargetParameters(models.Axis.Name, res[0][utils.SpecialIDParam]), record)
				continue
			}
			res, err := d.CreateSuperCall(utils.AllParams(ds.DBEntity.Name), map[string]interface{}{
				"name": record["name"],
			})
			if err == nil && len(res) > 0 {
				record[ds.EntityDBField] = res[0][utils.SpecialIDParam]
				d.CreateSuperCall(utils.AllParams(models.Axis.Name), record)
			}
		}
	}
	mapped = map[string]string{
		"Projet":            "code",
		"Tâche Projet":      "project_task",
		"Abrégé Projet":     "name",
		"Etat Ligne Projet": "state",
	}
	inside = []string{}
	for _, data := range datas {
		record := map[string]interface{}{}
		for i, header := range headers {
			if realLabel, ok := mapped[header]; ok && realLabel != "" && data[i] != "" {
				if strings.ToLower(data[i]) == "non" {
					record[realLabel] = false
				} else if strings.ToLower(data[i]) == "oui" {
					record[realLabel] = true
				} else {
					record[realLabel] = data[i]
				}
			}
			if strings.ToLower(header) == "date fin de projet" && data[i] != "" {
				s := strings.Split(data[i], "/")
				for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
					s[i], s[j] = s[j], s[i]
				}
				record["start_date"] = strings.Join(s, "-")
			}
			if strings.ToLower(header) == "date fin de projet" && data[i] != "" {
				s := strings.Split(data[i], "/")
				for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
					s[i], s[j] = s[j], s[i]
				}
				record["end_date"] = strings.Join(s, "-")
			}
			if strings.ToLower(header) == "code axe" && data[i] != "" {
				if res, err := d.GetDb().SelectQueryWithRestriction(models.Axis.Name, map[string]interface{}{
					"code": connector.Quote(data[i]),
				}, false); err == nil && len(res) > 0 {
					record[ds.RootID(models.Axis.Name)] = res[0][utils.SpecialIDParam]
				}
			}
			if strings.ToLower(header) == "email chef de projet" && data[i] != "" {
				if res, err := d.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
					"email": connector.Quote(data[i]),
				}, false); err == nil && len(res) > 0 {
					record[ds.UserDBField] = res[0][utils.SpecialIDParam]
				}
			}
		}
		if len(record) > 0 && !slices.Contains(inside, utils.GetString(record, "name")) {
			inside = append(inside, utils.GetString(record, "name"))
			if res, err := d.GetDb().SelectQueryWithRestriction(models.Project.Name, map[string]interface{}{
				"code": connector.Quote(utils.GetString(record, "code")),
			}, false); err == nil && len(res) > 0 {
				record[utils.SpecialIDParam] = res[0][utils.SpecialIDParam]
				d.UpdateSuperCall(utils.GetRowTargetParameters(models.Project.Name, res[0][utils.SpecialIDParam]), record)
				continue
			}
			res, err := d.CreateSuperCall(utils.AllParams(ds.DBEntity.Name),
				map[string]interface{}{
					"name": record["name"],
				})
			if err == nil && len(res) > 0 {
				record[ds.EntityDBField] = res[0][utils.SpecialIDParam]
				d.CreateSuperCall(utils.AllParams(models.Project.Name), record)
			}
		}
	}
}

func ImportUserHierachy() {
	mapped := map[string]string{
		"Salarié Présent ?": "active",
		"Login Utilisateur": "name",
		"Email Utilisateur": "email",
		"Matricule Salarié": "code",
	}
	d := domain.Domain(true, os.Getenv("SUPERADMIN_NAME"), nil)
	filepath := os.Getenv("USER_FILE_PATH")
	if filepath == "" {
		filepath = "./user_test.csv"
	}
	headers, datas := importFile(filepath)
	inside := []string{}
	for _, data := range datas {
		record := map[string]interface{}{}
		for i, header := range headers {
			if realLabel, ok := mapped[header]; ok && realLabel != "" && data[i] != "" {
				if strings.ToLower(data[i]) == "non" {
					record[realLabel] = false
				} else if strings.ToLower(data[i]) == "oui" {
					record[realLabel] = true
				} else {
					record[realLabel] = data[i]
				}

			}
		}
		if len(record) > 0 && !slices.Contains(inside, utils.GetString(record, "name")) {
			inside = append(inside, utils.GetString(record, "name"))
			if utils.GetBool(record, "active") {
				if res, err := d.GetDb().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
					"email": record["email"],
				}, false); err == nil && len(res) > 0 {
					record[utils.SpecialIDParam] = res[0][utils.SpecialIDParam]
					d.UpdateSuperCall(utils.GetRowTargetParameters(ds.DBUser.Name, res[0][utils.SpecialIDParam]), record)
					return
				}
			}
			d.CreateSuperCall(utils.AllParams(ds.DBUser.Name), record)
		}
	}
	for _, data := range datas {
		userID := ""
		hierarchyID := ""
		for i, header := range headers {
			if strings.ToLower(header) == "email utilisateur" && data[i] != "" {
				if res, err := d.Db.ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
					"email": connector.Quote(data[i]),
				}, false); err == nil && len(res) > 0 {
					userID = utils.GetString(res[0], utils.SpecialIDParam)
				}
			}
			if strings.ToLower(header) == "matricule responsable" && data[i] != "" {
				if res, err := d.Db.ClearQueryFilter().SelectQueryWithRestriction(ds.DBUser.Name, map[string]interface{}{
					"code": connector.Quote(data[i]),
				}, false); err == nil && len(res) > 0 {
					hierarchyID = utils.GetString(res[0], utils.SpecialIDParam)
				}
			}
		}
		if userID != "" && hierarchyID != "" {
			d.DeleteSuperCall(utils.AllParams(ds.DBHierarchy.Name), map[string]interface{}{
				ds.UserDBField: userID,
			})
			d.CreateSuperCall(utils.AllParams(ds.DBHierarchy.Name), map[string]interface{}{
				"parent_" + ds.UserDBField: hierarchyID,
				ds.UserDBField:             userID,
			})
		}
	}
}

func importFile(filePath string) ([]string, [][]string) {
	// Open CSV file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return []string{}, [][]string{}
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Failed to read CSV:", err)
		return []string{}, [][]string{}
	}

	if len(records) < 2 {
		fmt.Println("Not enough rows to sort")
		return []string{}, [][]string{}
	}

	headers := records[0]
	datas := records[1:]
	return headers, datas
}

package task

import (
	"slices"
	ds "sqldb-ws/domain/schema/database_resources"
	"sqldb-ws/domain/utils"
	"sqldb-ws/infrastructure/connector"
)

var endedRequestCache = map[string]map[string]map[string][]string{}

func GetEndedRequest(schemaID string, destID string, requestID string) *[]string {
	if s, ok := endedRequestCache[schemaID]; ok {
		if ss, ok := s[destID]; ok {
			if sss, ok := ss[requestID]; ok {
				return &sss
			}
		}
	}
	return nil
}

func IsEndedRequestByTask(schemaID string, destID string, requestID string, taskID string) ([]string, bool) {
	reqs := GetEndedRequest(schemaID, destID, requestID)
	if reqs == nil || len(*reqs) == 0 {
		return []string{}, false
	}
	t := *reqs
	if slices.Contains(t, taskID) {
		return []string{}, false
	}
	return t, slices.Contains(t, taskID)
}

func IsEndedRequest(schemaID string, destID string, requestID string) ([]string, bool) {
	reqs := GetEndedRequest(schemaID, destID, requestID)
	if reqs == nil || len(*reqs) == 0 {
		return []string{}, false
	}
	t := *reqs
	return t, true
}

func SetEndedRequest(schemaID string, destID string, requestID string, db connector.DB) {
	if _, ok := endedRequestCache[schemaID]; !ok {
		endedRequestCache[schemaID] = map[string]map[string][]string{}
	}
	if _, ok := endedRequestCache[schemaID][destID]; !ok {
		endedRequestCache[schemaID][destID] = map[string][]string{}
	}
	if res, err := db.SelectQueryWithRestriction(ds.DBTask.Name, map[string]interface{}{
		ds.RequestDBField: requestID,
	}, false); err == nil && len(res) > 0 {
		if _, ok := endedRequestCache[schemaID][destID][requestID]; !ok {
			endedRequestCache[schemaID][destID][requestID] = []string{}
		}
		for _, r := range res {
			endedRequestCache[schemaID][destID][requestID] = append(
				endedRequestCache[schemaID][destID][requestID], utils.GetString(r, utils.SpecialIDParam))
		}
	}

}

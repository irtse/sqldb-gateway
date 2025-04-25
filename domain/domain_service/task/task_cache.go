package task

type Task struct {
	RequestID string
	TaskID    string
}

var taskCache = map[string]map[string][]Task{}

func GetTasks(schemaID string, destID string) *[]Task {
	if s, ok := taskCache[schemaID]; ok {
		if ss, ok := s[destID]; ok {
			return &ss
		}
	}
	return nil
}

func IsTask(schemaID string, destID string, taskID string) ([]Task, bool) {
	tasks := GetTasks(schemaID, destID)
	if tasks == nil || len(*tasks) == 0 {
		return []Task{}, false
	}
	t := *tasks
	for _, tt := range t {
		if tt.TaskID == taskID {
			return t, true
		}
	}
	return nil, false
}

func IsTaskFromRequest(schemaID string, destID string, requestID string) ([]Task, bool) {
	tasks := GetTasks(schemaID, destID)
	if tasks == nil || len(*tasks) == 0 {
		return []Task{}, false
	}
	t := *tasks
	return t, t[0].RequestID == requestID
}

func SetTasks(schemaID string, destID string, requestID string, taskID string) {
	if _, ok := taskCache[schemaID]; !ok {
		taskCache[schemaID] = map[string][]Task{}
	}
	if _, ok := taskCache[schemaID][destID]; !ok {
		taskCache[schemaID][destID] = []Task{}
	}
	taskCache[schemaID][destID] = append(taskCache[schemaID][destID], Task{
		RequestID: requestID,
		TaskID:    taskID,
	})
}

func DeleteTasks(schemaID string, destID string, taskID string) {
	if _, ok := taskCache[schemaID]; ok {
		if _, ok := taskCache[schemaID][destID]; ok {
			for _, task := range taskCache[schemaID][destID] {
				tasks := []Task{}
				if task.TaskID != taskID {
					tasks = append(tasks, task)
				}
				if len(tasks) == 0 {
					delete(taskCache[schemaID], destID)
				} else {
					taskCache[schemaID][destID] = tasks
				}
			}

		}
	}
}

// OK
var taskViewCache = map[string]map[string]map[string][]string{}

// for assignee
func GetViewTask(schemaID string, destID string, userID string) []string {
	if s, ok := taskViewCache[schemaID]; ok {
		if ss, ok := s[destID]; ok {
			if sss, ok := ss[userID]; ok {
				return sss
			}
		}
	}
	return []string{}
}

func SetViewTask(schemaID string, destID string, userID string, vals []string) {
	if _, ok := taskViewCache[schemaID]; !ok {
		taskViewCache[schemaID] = map[string]map[string][]string{}
	}
	if _, ok := taskViewCache[schemaID][destID]; !ok {
		taskViewCache[schemaID][destID] = map[string][]string{}
	}
	taskViewCache[schemaID][destID][userID] = vals
}

func DeleteViewTask(schemaID string, destID string, userID string) {
	if _, ok := taskViewCache[schemaID]; ok {
		if _, ok := taskViewCache[schemaID][destID]; ok {
			delete(taskViewCache[schemaID][destID], userID)
		}
	}
}

// everytime there is an action related
var taskReadOnlyCache = map[string]map[string][]string{}

func GetReadonlyTask(schemaID string, destID string) *[]string {
	if s, ok := taskReadOnlyCache[schemaID]; ok {
		if ss, ok := s[destID]; ok {
			return &ss
		}
	}
	return nil
}

func SetReadonlyTask(schemaID string, destID string, assignee string) {
	if _, ok := taskReadOnlyCache[schemaID]; !ok {
		taskReadOnlyCache[schemaID] = map[string][]string{}
	}
	if _, ok := taskReadOnlyCache[schemaID][destID]; !ok {
		taskReadOnlyCache[schemaID][destID] = []string{}
	}
	taskReadOnlyCache[schemaID][destID] = append(taskReadOnlyCache[schemaID][destID], assignee)
}

func DeleteReadonlyTask(schemaID string, destID string, userID string) {
	if _, ok := taskReadOnlyCache[schemaID]; ok {
		if _, ok := taskReadOnlyCache[schemaID][destID]; ok {
			for _, usr := range taskReadOnlyCache[schemaID][destID] {
				usrs := []string{}
				if userID != usr {
					usrs = append(usrs, usr)
				}
				if len(usrs) == 0 {
					delete(taskReadOnlyCache[schemaID], destID)
				} else {
					taskReadOnlyCache[schemaID][destID] = usrs
				}
			}

		}
	}
}

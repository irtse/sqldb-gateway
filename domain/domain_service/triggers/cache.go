package triggers

var triggersRedirection = map[string]string{}

func GetRedirection(key string) (string, bool) {
	k, ok := triggersRedirection[key]
	delete(triggersRedirection, key)
	return k, ok
}

func HasRedirection(key string) bool {
	_, ok := triggersRedirection[key]
	return ok
}

func SetRedirection(key string, path string) {
	triggersRedirection[key] = path
}

func DeleteRedirection(key string) {
	delete(triggersRedirection, key)
}

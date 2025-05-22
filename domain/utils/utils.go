package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func PrepareEnum(enum string) string {
	if !strings.Contains(enum, "enum") {
		return enum
	}
	return TransformType(enum)
}

func TransformType(enum string) string {
	e := strings.Replace(ToString(enum), " ", "", -1)
	e = strings.Replace(e, "'", "", -1)
	e = strings.Replace(e, "(", "__", -1)
	e = strings.Replace(e, ",", "_", -1)
	e = strings.Replace(e, ")", "", -1)
	return strings.ToLower(e)
}

func ToMap(who interface{}) map[string]interface{} {
	if reflect.TypeOf(who).Kind() == reflect.Map {
		return who.(map[string]interface{})
	}
	return map[string]interface{}{}
}

func ToListAnonymized(who []string) []interface{} {
	i := []interface{}{}
	if reflect.TypeOf(who).Kind() == reflect.Slice {
		for _, w := range who {
			i = append(i, w)
		}
		return i
	}
	return i
}

func ToList(who interface{}) []interface{} {
	if reflect.TypeOf(who).Kind() == reflect.Slice {
		return who.([]interface{})
	}
	return []interface{}{}
}

func ToFloat64(who interface{}) float64 {
	if who == nil {
		return 0
	}
	i, err := strconv.ParseFloat(fmt.Sprintf("%v", who), 64)
	if err != nil {
		return 0
	}
	return float64(i)
}

func ToInt64(who interface{}) int64 {
	if who == nil {
		return 0
	}
	i, err := strconv.Atoi(fmt.Sprintf("%v", who))
	if err != nil {
		return 0
	}
	return int64(i)
}

func ToString(who interface{}) string {
	if who == nil {
		return ""
	}
	return fmt.Sprintf("%v", who)
}

func Compare(who interface{}, what interface{}) bool {
	return who != nil && fmt.Sprintf("%v", who) == fmt.Sprintf("%v", what)
}

func Translate(str string) string {
	url := "https://libretranslate.com/translate"
	target := os.Getenv("LANG")
	if target == "" {
		target = "fr"
	}

	data := map[string]string{
		"q":      str,
		"source": "en",
		"target": target,
		"format": "text",
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if body, err := io.ReadAll(resp.Body); err == nil {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		return GetString(result, "translatedText")
	}
	return str
}

func SearchInFile(filename string, searchTerm string) bool {
	filePath := filename
	if !strings.Contains(filePath, "/mnt/files/") {
		filePath = "/mnt/files/" + filename
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, searchTerm) {
			return true
			// break // Uncomment to stop at first match
		}
	}
	return false
}

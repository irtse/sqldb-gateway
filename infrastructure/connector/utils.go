package connector

import (
	"fmt"
	"net/url"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

var COUNTREQUEST = 0

var SpecialTypes = []string{"char", "text", "date", "time", "interval", "var", "blob", "set", "enum", "year", "USER-DEFINED"}

func Quote(s string) string { return "'" + s + "'" }

func RemoveLastChar(s string) string {
	r := []rune(s)
	if len(r) > 0 {
		return string(r[:len(r)-1])
	}
	return string(r)
}

func FormatMathViewQuery(algo string, col string, naming ...string) string {
	resName := "result"
	if len(naming) > 0 {
		resName = naming[0]
	}
	return strings.ToUpper(algo) + "(" + col + ") as " + resName
}

func FormatSQLRestrictionWhereInjection(injection string, getTypeAndLink func(string) (string, string, error)) string {
	alterRestr := ""
	injection = SQLInjectionProtector(injection)
	ands := strings.Split(injection, "+")
	for _, andUndecoded := range ands {
		and, _ := url.QueryUnescape(fmt.Sprint(andUndecoded))
		if len(strings.Trim(alterRestr, " ")) > 0 {
			alterRestr += " AND "
		}
		ors := strings.Split(and, "|")
		if len(ors) == 0 {
			continue
		}
		orRestr := ""
		for _, or := range ors {
			operator := "~"
			keyVal := []string{}
			if strings.Contains(or, "<>~") {
				keyVal = strings.Split(or, "<>~")
				operator = " NOT LIKE "
			} else if strings.Contains(or, "~") {
				keyVal = strings.Split(or, "~")
				operator = " LIKE "
			} else if strings.Contains(or, "<>") {
				keyVal = strings.Split(or, "<>")
				operator = "<>"
			} else if strings.Contains(or, "<:") {
				keyVal = strings.Split(or, "<:")
				operator = "<="
			} else if strings.Contains(or, ">:") {
				keyVal = strings.Split(or, ">:")
				operator = ">="
			} else if strings.Contains(or, ":") {
				keyVal = strings.Split(or, ":")
				operator = "="
			} else if strings.Contains(or, "<") {
				keyVal = strings.Split(or, "<")
				operator = "<"
			} else if strings.Contains(or, ">") {
				keyVal = strings.Split(or, ">")
				operator = ">"
			}
			if len(keyVal) != 2 {
				continue
			}
			typ, link, err := getTypeAndLink(keyVal[0])
			if err == nil && keyVal[0] != "id" {
				continue
			}
			if len(strings.Trim(orRestr, " ")) > 0 {
				orRestr += " OR "
			}
			orRestr = MakeSqlItem(orRestr, typ, link, keyVal[0], keyVal[1], operator)
		}
		if len(orRestr) > 0 {
			alterRestr += "( " + orRestr + " )"
		}
	}
	alterRestr = strings.ReplaceAll(strings.ReplaceAll(alterRestr, " OR ()", ""), " AND ()", "")
	alterRestr = strings.ReplaceAll(alterRestr, "()", "")
	return alterRestr
}

func MakeSqlItem(alterRestr string, typ string, foreignName string, key string, or string, operator string) string {
	sql := or
	sql = FormatForSQL(typ, sql)
	if sql == "" {
		return alterRestr
	}
	if strings.Contains(sql, "NULL") {
		operator = "IS "
	}
	if foreignName != "" {
		if strings.Contains(sql, "%") {
			alterRestr += key + " IN (SELECT id FROM " + foreignName + " WHERE name::text LIKE " + sql + " OR id::text " + operator + sql + ")"
		} else {
			if strings.Contains(sql, "'") {
				if strings.Contains(sql, "NULL") {
					alterRestr += key + " IN (SELECT id FROM " + foreignName + " WHERE name IS " + sql + ")"
				} else {
					alterRestr += key + " IN (SELECT id FROM " + foreignName + " WHERE name = " + sql + ")"
				}
			} else {
				alterRestr += key + " IN (SELECT id FROM " + foreignName + " WHERE id " + operator + " " + sql + ")"
			}
		}
	} else if strings.Contains(sql, "%") {
		alterRestr += key + "::text " + operator + sql
	} else {
		alterRestr += key + " " + operator + " " + sql
	}
	return alterRestr
}

func FormatLimit(limited string, offset interface{}) string {
	if i, err := strconv.Atoi(limited); err == nil {
		limited = "LIMIT " + fmt.Sprintf("%v", i)
		if offset != nil && offset != "" {
			if i2, err := strconv.Atoi(fmt.Sprintf("%v", offset)); err == nil {
				limited += " OFFSET " + fmt.Sprintf("%v", i2)
			}
		}
	}
	return limited
}

func FormatOperatorSQLRestriction(operator interface{}, separator interface{}, name string, value interface{}, typ string) string {
	if operator == nil || separator == nil {
		return ""
	}
	filter := ""
	if len(filter) > 0 {
		filter += " " + fmt.Sprintf("%v", separator) + " "
	}
	if fmt.Sprintf("%v", operator) == "LIKE" {
		filter += name + "::text " + fmt.Sprintf("%v", operator) + " '%" + fmt.Sprintf("%v", value) + "%'"
	} else {
		filter += name + " " + fmt.Sprintf("%v", operator) + " " + FormatForSQL(typ, value)
	}
	return filter
}

func FormatSQLRestrictionByList(SQLrestriction string, restrictions []interface{}, isOr bool) string {
	for _, v := range restrictions {
		if len(SQLrestriction) > 0 {
			if isOr {
				SQLrestriction += " OR "
			} else {
				SQLrestriction += " AND "
			}
		}
		SQLrestriction += fmt.Sprintf("%v", v)
	}
	return SQLrestriction
}

func FormatSQLRestrictionWhereByMap(SQLrestriction string, restrictions map[string]interface{}, isOr bool) string {
	for k, r := range restrictions {
		if len(SQLrestriction) > 0 {
			if isOr {
				SQLrestriction += " OR "
			} else {
				SQLrestriction += " AND "
			}
		}
		if r == nil {
			SQLrestriction += k + " IS NULL"
		} else {
			divided := strings.Split(fmt.Sprintf("%v", r), " ")
			if len(divided) > 1 && slices.Contains([]string{"SELECT", "INSERT", "UPDATE", "DELETE"}, strings.ToUpper(divided[0])) {
				SQLrestriction += k + " IN (" + fmt.Sprintf("%v", r) + ")"
			} else if len(divided) > 1 && slices.Contains([]string{"!SELECT", "!INSERT", "!UPDATE", "!DELETE"}, strings.ToUpper(divided[0])) {
				SQLrestriction += k + " NOT IN (" + fmt.Sprintf("%v", r) + ")"
			} else if reflect.TypeOf(r).Kind() == reflect.Slice {
				str := ""
				for _, rr := range r.([]string) {
					str += fmt.Sprintf("%v", rr) + ","
				}
				SQLrestriction += k + " IN (" + RemoveLastChar(str) + ")"
			} else {
				SQLrestriction += k + "=" + fmt.Sprintf("%v", r)
			}
		}
	}
	return SQLrestriction
}

func FormatSQLRestrictionWhere(SQLrestriction string, restriction string, verify func() bool, additionnalRestr ...string) (string, string) {
	if restriction != "" && verify() && len(restriction) > 0 {
		if len(SQLrestriction) > 0 {
			SQLrestriction += " AND "
		}
		SQLrestriction += restriction
	}
	lateAddition := ""
	for _, restr := range additionnalRestr {
		if strings.Contains(restr, " IN ") {
			if len(lateAddition) > 0 {
				lateAddition += " AND "
			}
			lateAddition += restr
			continue
		}
		if len(SQLrestriction) > 0 && len(restr) > 0 {
			SQLrestriction = restr + " AND " + SQLrestriction
		} else {
			SQLrestriction = restr
		}
	}
	return SQLrestriction, lateAddition
}

func FormatSQLLimit(limit string, offset string) string {
	var SQLlimit string
	if i, err := strconv.Atoi(limit); err == nil {
		SQLlimit = "LIMIT " + fmt.Sprintf("%v", i)
		if i2, err := strconv.Atoi(offset); err == nil {
			SQLlimit += " OFFSET " + fmt.Sprintf("%v", i2)
		}
	}
	return SQLlimit
}

func FormatSQLOrderBy(orderBy []string, ascDesc []string, verify func(string) bool) string {
	var order string
	if len(orderBy) == 0 {
		return "id DESC"
	}
	for i, el := range orderBy {
		if verify(el) && len(ascDesc) > i {
			upperAscDesc := strings.Replace(strings.ToUpper(ascDesc[i]), " ", "", -1)
			if upperAscDesc == "ASC" || upperAscDesc == "DESC" {
				order += SQLInjectionProtector(el + " " + upperAscDesc + ",")
				continue
			}
			order += SQLInjectionProtector(el + " ASC,")
		}
	}
	return RemoveLastChar(order)
}

func FormatForSQL(datatype string, value interface{}) string {
	if value == nil {
		return ""
	}
	strval := fmt.Sprintf("%v", value)
	if len(strval) == 0 {
		return ""
	}
	if strval == "NULL" || strval == "NOT NULL" {
		return strval
	}
	for _, typ := range SpecialTypes {
		if strings.Contains(datatype, typ) {
			if value == "CURRENT_TIMESTAMP" {
				return fmt.Sprint(value)
			} else {
				decodedValue := fmt.Sprint(value)
				if strings.Contains(strings.ToUpper(datatype), "DATE") || strings.Contains(strings.ToUpper(datatype), "TIME") {
					if len(decodedValue) > 10 {
						decodedValue = decodedValue[:10]
					}
				}
				return Quote(strings.Replace(SQLInjectionProtector(decodedValue), "'", "''", -1))
			}
		}
	}
	if strings.Contains(strval, "%") {
		decodedValue := fmt.Sprint(value)
		return Quote(strings.Replace(SQLInjectionProtector(decodedValue), "'", "''", -1))
	}
	return SQLInjectionProtector(strval)
}

func SQLInjectionProtector(injection string) string {
	quoteCounter := strings.Count(injection, "'")
	quoteCounter2 := strings.Count(injection, "\"")
	if (quoteCounter%2) != 0 || (quoteCounter2%2) != 0 {
		log.Error().Msg("injection alert: strange activity of quoting founded")
		return ""
	}
	notAllowedChar := []string{"«", "#", "union", ";", ")", "%27", "%22", "%23", "%3B", "%29"}
	for _, char := range notAllowedChar {
		if strings.Contains(strings.ToLower(injection), char) {
			log.Error().Msg("injection alert: not allowed " + char + " filter")
			return ""
		}
	}
	return injection
}

func FormatEnumName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), ",", "_"), "'", ""), "(", "__"), ")", ""), " ", "")
}

func FormatReverseEnumName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ToLower(name), "__", "('"), "_", "','") + "')"
}

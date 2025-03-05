package utils

import (
	"fmt"
	"strings"
)

func PrepareEnum(enum string) string {
	e := strings.Replace(fmt.Sprintf("%v", enum), " ", "", -1)
	e = strings.Replace(e, "'", "", -1)
	e = strings.Replace(e, "(", "__", -1)
	e = strings.Replace(e, ",", "_", -1)
	e = strings.Replace(e, ")", "", -1)
	return strings.ToLower(e)
}

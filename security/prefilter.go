package security

import (
	"fmt"
	"sqldb-ws/models"
)

func CheckSelect(dbuser_id string, table *string, columns *string, restriction *string) {

	fmt.Println(dbuser_id, fmt.Sprintf("%v", table), fmt.Sprintf("%v", columns), fmt.Sprintf("%v", restriction))
}

func removeLastChar(s string) string {
	r := []rune(s)
	return string(r[:len(r)-1])
}

func Test() {
	var teste = models.UsersRights

	for _, element := range teste {
		fmt.Println(fmt.Sprintf("%v", element))
	}
}

/*

##### SCHEMA DE FONCTION #####

*/

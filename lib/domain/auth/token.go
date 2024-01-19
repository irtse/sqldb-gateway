package auth

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("weakest-secret")

type Token struct {}

func (t *Token) Create(user_id string, superAdmin bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		SESSIONS_KEY : user_id,
		ADMIN_KEY : superAdmin,
		"exp" : time.Now().Add(time.Hour * 24).Unix(),
	})
	tokenStr, err := token.SignedString(secret)
	if err != nil { return "", err }
	return tokenStr, nil
}

func (t *Token) Verify(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil { return nil, err }
	if !token.Valid { return nil, fmt.Errorf("invalid token")}
	tokenStr, err = token.SignedString(secret)
	return token, err
}
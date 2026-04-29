package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	SampleModelId int
	Name          string
	jwt.RegisteredClaims
}

func NewJwtHS256(id int, name string) (string, error) {

	newClaims := CustomClaims{
		SampleModelId: id,
		Name:          name,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   name,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 60)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	return jwt.SignedString([]byte("secret_key"))
}

func VaildJwtHS256(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invaild token method")
		}
		return []byte("secret_key"), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, jwt.ErrInvalidKeyType
	}

	return claims, nil
}

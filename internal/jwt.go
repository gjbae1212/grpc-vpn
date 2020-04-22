package internal

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	HS256 = "HS256"
)

// DecodeJWT is to decode jwt using HS256.
func DecodeJWT(data string, salt []byte) (*jwt.Token, error) {
	if data == "" || len(salt) == 0 {
		return nil, errors.Wrapf(ErrorInvalidParams, "Method: DecodeJWT")
	}

	token, err := jwt.ParseWithClaims(data, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != HS256 {
			return nil, errors.Wrapf(ErrorInvalidJWT, "Method: DecodeJWT")
		}
		return salt, nil
	})
	if err != nil {
		return nil, errors.Wrapf(ErrorInvalidJWT, "Method: DecodeJWT")
	}
	if token == nil || !token.Valid {
		return nil, errors.Wrapf(ErrorInvalidJWT, "Method: DecodeJWT")
	}
	return token, nil

}

// EncodeJWT is to encode jwt using HS256.
func EncodeJWT(claims *jwt.StandardClaims, salt []byte) (string, error) {
	if claims == nil || len(salt) == 0 {
		return "", errors.Wrapf(ErrorInvalidParams, "Method: EncodeJWT")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(salt)
}

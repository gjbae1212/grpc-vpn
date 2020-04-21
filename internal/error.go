package internal

import "errors"

var (
	ErrorInvalidParams = errors.New("[ERR] Invalid Params")
	ErrorUnauthorized  = errors.New("[ERR] Unauthorized")
	ErrorInvalidJWT    = errors.New("[ERR] Invalid JWT")
)

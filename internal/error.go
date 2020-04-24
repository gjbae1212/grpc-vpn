package internal

import "errors"

var (
	ErrorUnknown              = errors.New("[ERR] Unknown")
	ErrorInvalidParams        = errors.New("[ERR] Invalid Params")
	ErrorUnauthorized         = errors.New("[ERR] Unauthorized")
	ErrorInvalidJWT           = errors.New("[ERR] Invalid JWT")
	ErrorInvalidContext       = errors.New("[ERR] Invalid Context")
	ErrorExceedClientPool     = errors.New("[ERR] Exceed Client Pool")
	ErrorCloseConnection      = errors.New("[ERR] Close Connection")
	ErrorReceiveUnknownPacket = errors.New("[ERR] Receive Unknown Packet")
	ErrorMismatchVpnIP        = errors.New("[ERR] Mismatch Vpn IP")
	ErrorStoppingServer       = errors.New("[ERR] Stopping Server")
	ErrorAlreadyRunning       = errors.New("[ERR] Already Running")
)

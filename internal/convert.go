package internal

import "strconv"

// InterfaceToString converts a value having interface type to string
func InterfaceToString(i interface{}) string {
	if i == nil {
		return ""
	}

	switch i.(type) {
	case int:
		return strconv.FormatInt(int64(i.(int)), 10)
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case int32:
		return strconv.FormatInt(int64(i.(int32)), 10)
	case float32:
		return strconv.FormatFloat(float64(i.(float32)), 'f', -1, 64)
	case float64:
		return strconv.FormatFloat(i.(float64), 'f', -1, 64)
	case bool:
		return strconv.FormatBool(i.(bool))
	case string:
		return i.(string)
	default:
		return ""
	}
}

package internal

// IsMatchedStringFromSlice checks whether to be matched string from slice or not.
func IsMatchedStringFromSlice(s string, slice []string) bool {
	for _, e := range slice {
		if e == s {
			return true
		}
	}
	return false
}

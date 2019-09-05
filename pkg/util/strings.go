package util

import "strings"

func EqualsIgnoreCase(a string, another ...string) bool {
	lower := strings.ToLower(a)
	for _, s := range another {
		if lower != strings.ToLower(s) {
			return false
		}
	}
	return true
}

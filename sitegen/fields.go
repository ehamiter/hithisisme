package sitegen

import (
	"strconv"
	"strings"
)

// Resolve returns the value at path from obj using dotted notation.
func Resolve(obj interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	cur := obj
	for _, p := range parts {
		switch c := cur.(type) {
		case map[string]interface{}:
			cur = c[p]
		case []interface{}:
			idx, err := strconv.Atoi(p)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil
			}
			cur = c[idx]
		default:
			return nil
		}
	}
	return cur
}

// GetString gets a string value or empty string.
func GetString(obj interface{}) string {
	if s, ok := obj.(string); ok {
		return s
	}
	return ""
}

package helpers

import "strings"

func NormalizeString(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

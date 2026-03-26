package validators

import "strings"

func normalizeString(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

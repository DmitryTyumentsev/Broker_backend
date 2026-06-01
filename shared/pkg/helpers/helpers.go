package helpers

import "strings"

func NormalizeEmail(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func NormalizeText(v string) string {
	return strings.TrimSpace(v)
}

func NormalizeOptionText(v *string) *string {
	if v == nil {
		return nil
	}

	normalizeV := NormalizeText(*v)
	if normalizeV == "" {
		return nil
	}

	return &normalizeV
}

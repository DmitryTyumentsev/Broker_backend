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

func NormalizePhoneNumber(v string) string {
	trimmed := strings.TrimSpace(v)
	hasExplicitCountryCode := strings.HasPrefix(trimmed, "+")

	var digits strings.Builder
	digits.Grow(len(trimmed))

	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] >= '0' && trimmed[i] <= '9' {
			digits.WriteByte(trimmed[i])
		}
	}

	normalized := digits.String()
	switch {
	case len(normalized) == 10 && !hasExplicitCountryCode:
		return "7" + normalized
	case len(normalized) == 11 && normalized[0] == '8' && !hasExplicitCountryCode:
		return "7" + normalized[1:]
	default:
		return normalized
	}
}

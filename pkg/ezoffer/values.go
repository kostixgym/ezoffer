package ezoffer

import (
	"fmt"
	"strings"
)

// ValidationError says a filter value is not one of the values the column can
// hold. Returning it beats silently dropping the filter: a dropped filter widens
// the result set, and the client sees a full list where it asked for a subset.
type ValidationError struct {
	Field string
	Value string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("invalid %s: %q", e.Field, e.Value)
}

// The dictionaries below mirror the CHECK constraints in docs/ezoffer.sql. Keys
// are lowercased so a client may send "Lead" or "liveCoding" in any casing;
// values are the exact spelling stored in the column.
var (
	gradeValues = canonicalMap("junior", "middle", "senior", "lead")

	taskTypeValues = canonicalMap("liveCoding", "algorithms", "systemDesign")
)

func canonicalMap(values ...string) map[string]string {
	out := make(map[string]string, len(values))
	for _, v := range values {
		out[strings.ToLower(v)] = v
	}

	return out
}

func strValue(v *string) string {
	if v == nil {
		return ""
	}

	return strings.TrimSpace(*v)
}

// canonical maps one filter value onto its stored spelling.
func canonical(field, value string, known map[string]string) (string, error) {
	v, ok := known[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return "", ValidationError{Field: field, Value: value}
	}

	return v, nil
}

// canonicalAll maps a whole filter list. Blanks are skipped, anything unknown is
// an error.
func canonicalAll(field string, values []string, known map[string]string) ([]string, error) {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if strings.TrimSpace(v) == "" {
			continue
		}

		c, err := canonical(field, v, known)
		if err != nil {
			return nil, err
		}

		out = append(out, c)
	}

	return out, nil
}

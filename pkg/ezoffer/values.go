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


var (
	gradeValues = canonicalMap("junior", "middle", "senior", "lead")

	taskTypeValues = canonicalMap("liveCoding", "algorithms", "systemDesign")

	interviewTypeValues = canonicalMap("technical", "liveCoding", "algorithmic", "hrScreening", "final", "systemDesign")
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


func canonical(field, value string, known map[string]string) (string, error) {
	v, ok := known[strings.ToLower(strings.TrimSpace(value))]
	if !ok {
		return "", ValidationError{Field: field, Value: value}
	}

	return v, nil
}

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

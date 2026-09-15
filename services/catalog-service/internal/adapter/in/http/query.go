package http

import (
	"fmt"
	"strconv"
	"strings"
)

// intParam parses an optional integer query parameter.
//
// An absent value yields fallback (so the service can apply its own default);
// a present-but-unparsable value is an error, because silently ignoring bad
// input would make the caller think their filter was applied when it was not.
func intParam(raw string, fallback int) (int, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return fallback, nil
	}
	v, err := strconv.Atoi(trimmed)
	if err != nil {
		return 0, err
	}
	return v, nil
}

// floatParam parses an optional numeric query parameter. Absent → nil, which
// is how "no bound on this side of the range" is expressed.
func floatParam(raw string) (*float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// boolParam parses an optional boolean query parameter, accepting
// true/false/1/0 in any case. Absent → false.
func boolParam(raw string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "0", "false":
		return false, nil
	case "1", "true":
		return true, nil
	}
	return false, fmt.Errorf("expected true/false/1/0, got %q", raw)
}

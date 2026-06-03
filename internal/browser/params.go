package browser

import "fmt"

// paramString extracts a required string parameter from a Command's Params map.
// Returns a descriptive error when the key is missing or has a non-string type,
// so callers can surface it to the user instead of panicking.
func paramString(params map[string]interface{}, key string) (string, error) {
	if params == nil {
		return "", fmt.Errorf("missing required param: %s (no params provided)", key)
	}
	v, ok := params[key]
	if !ok {
		return "", fmt.Errorf("missing required param: %s", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("param %s must be string, got %T", key, v)
	}
	return s, nil
}

// optString extracts an optional string parameter; returns def when absent
// or when the value is not a string.
func optString(params map[string]interface{}, key, def string) string {
	if params == nil {
		return def
	}
	v, ok := params[key]
	if !ok {
		return def
	}
	if s, ok := v.(string); ok {
		return s
	}
	return def
}

// paramFloat extracts a required number parameter. JSON unmarshals all
// numbers to float64 by default, but caller-built maps (e.g. when a client
// constructs params with `map[string]interface{}{"x": 3}` instead of
// serializing JSON) may use int. We accept both so that both code paths
// work without forcing every caller through a JSON round-trip.
func paramFloat(params map[string]interface{}, key string) (float64, error) {
	if params == nil {
		return 0, fmt.Errorf("missing required param: %s (no params provided)", key)
	}
	v, ok := params[key]
	if !ok {
		return 0, fmt.Errorf("missing required param: %s", key)
	}
	switch n := v.(type) {
	case float64:
		return n, nil
	case int:
		return float64(n), nil
	case int64:
		return float64(n), nil
	default:
		return 0, fmt.Errorf("param %s must be number, got %T", key, v)
	}
}

// optFloat extracts an optional number parameter; returns def when absent
// or when the value is not a number. Accepts float64, int, and int64.
func optFloat(params map[string]interface{}, key string, def float64) float64 {
	if params == nil {
		return def
	}
	v, ok := params[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	case int64:
		return float64(n)
	}
	return def
}

// paramBool extracts a required bool parameter.
func paramBool(params map[string]interface{}, key string) (bool, error) {
	if params == nil {
		return false, fmt.Errorf("missing required param: %s (no params provided)", key)
	}
	v, ok := params[key]
	if !ok {
		return false, fmt.Errorf("missing required param: %s", key)
	}
	b, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("param %s must be bool, got %T", key, v)
	}
	return b, nil
}

// optBool extracts an optional bool parameter; returns def when absent
// or when the value is not a bool.
func optBool(params map[string]interface{}, key string, def bool) bool {
	if params == nil {
		return def
	}
	v, ok := params[key]
	if !ok {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

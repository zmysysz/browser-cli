package browser

import (
	"testing"
)

func TestParamString(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		key     string
		want    string
		wantErr bool
	}{
		{"nil map", nil, "x", "", true},
		{"missing key", map[string]interface{}{"y": "z"}, "x", "", true},
		{"valid string", map[string]interface{}{"x": "hello"}, "x", "hello", false},
		{"wrong type", map[string]interface{}{"x": 42}, "x", "", true},
		{"empty string", map[string]interface{}{"x": ""}, "x", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := paramString(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("paramString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("paramString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOptString(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		key    string
		def    string
		want   string
	}{
		{"nil map", nil, "x", "default", "default"},
		{"missing key", map[string]interface{}{"y": "z"}, "x", "default", "default"},
		{"valid", map[string]interface{}{"x": "hello"}, "x", "default", "hello"},
		{"wrong type", map[string]interface{}{"x": 42}, "x", "default", "default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := optString(tt.params, tt.key, tt.def); got != tt.want {
				t.Errorf("optString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParamFloat(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		key     string
		want    float64
		wantErr bool
	}{
		{"nil map", nil, "x", 0, true},
		{"missing key", map[string]interface{}{"y": 1.0}, "x", 0, true},
		{"float64", map[string]interface{}{"x": 3.14}, "x", 3.14, false},
		{"int (caller-built map)", map[string]interface{}{"x": 3}, "x", 3.0, false},
		{"int64", map[string]interface{}{"x": int64(7)}, "x", 7.0, false},
		{"string (rejected)", map[string]interface{}{"x": "1.5"}, "x", 0, true},
		{"bool (rejected)", map[string]interface{}{"x": true}, "x", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := paramFloat(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("paramFloat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("paramFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptFloat(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		key    string
		def    float64
		want   float64
	}{
		{"nil map", nil, "x", 9.9, 9.9},
		{"missing key", map[string]interface{}{}, "x", 9.9, 9.9},
		{"float64", map[string]interface{}{"x": 1.5}, "x", 0, 1.5},
		{"int", map[string]interface{}{"x": 2}, "x", 0, 2.0},
		{"string (rejected, use def)", map[string]interface{}{"x": "x"}, "x", 4.4, 4.4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := optFloat(tt.params, tt.key, tt.def); got != tt.want {
				t.Errorf("optFloat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParamBool(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		key     string
		want    bool
		wantErr bool
	}{
		{"nil map", nil, "x", false, true},
		{"missing key", map[string]interface{}{"y": true}, "x", false, true},
		{"true", map[string]interface{}{"x": true}, "x", true, false},
		{"false", map[string]interface{}{"x": false}, "x", false, false},
		{"string (rejected)", map[string]interface{}{"x": "true"}, "x", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := paramBool(tt.params, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("paramBool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("paramBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptBool(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		key    string
		def    bool
		want   bool
	}{
		{"nil map", nil, "x", true, true},
		{"missing key", map[string]interface{}{}, "x", false, false},
		{"true", map[string]interface{}{"x": true}, "x", false, true},
		{"false", map[string]interface{}{"x": false}, "x", true, false},
		{"string (rejected, use def)", map[string]interface{}{"x": "x"}, "x", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := optBool(tt.params, tt.key, tt.def); got != tt.want {
				t.Errorf("optBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

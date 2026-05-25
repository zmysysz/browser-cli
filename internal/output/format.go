package output

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Format represents output format
type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatJSON     Format = "json"
)

// Result represents a command result
type Result struct {
	Command string      `json:"command"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Formatter handles output formatting
type Formatter struct {
	format Format
}

// NewFormatter creates a new formatter
func NewFormatter(format Format) *Formatter {
	return &Formatter{format: format}
}

// Format outputs the result
func (f *Formatter) Format(r Result) string {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(r)
	default:
		return f.formatMarkdown(r)
	}
}

func (f *Formatter) formatJSON(r Result) string {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "marshal failed: %s"}`, err)
	}
	return string(b)
}

func (f *Formatter) formatMarkdown(r Result) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("## %s\n", strings.Title(r.Command)))
	sb.WriteString(fmt.Sprintf("- Status: %s\n", r.Status))

	if r.Error != "" {
		sb.WriteString(fmt.Sprintf("- Error: %s\n", r.Error))
		return sb.String()
	}

	if r.Data == nil {
		return sb.String()
	}

	// Format data based on type
	switch v := r.Data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", key, val))
		}
	case []interface{}:
		for i, item := range v {
			sb.WriteString(fmt.Sprintf("%d. %v\n", i+1, item))
		}
	default:
		sb.WriteString(fmt.Sprintf("- Result: %v\n", v))
	}

	return sb.String()
}
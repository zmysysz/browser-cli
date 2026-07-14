package commands

import (
	"encoding/json"
	"fmt"
	"strings"
)

// printResult renders a server response to stdout and returns an error
// if the command failed (so cobra produces a non-zero exit code).
//
// format: "json" or "markdown"
func printResult(action string, success bool, data map[string]interface{}, errMsg string, format string) error {
	switch format {
	case "markdown":
		printMarkdown(action, success, data, errMsg)
	default:
		printJSON(action, success, data, errMsg)
	}

	if !success {
		return fmt.Errorf("%s failed: %s", action, errMsg)
	}
	return nil
}


func printJSON(action string, success bool, data map[string]interface{}, errMsg string) {
	output := map[string]interface{}{
		"command": action,
		"status":  "success",
		"session": sessionID,
	}
	if data != nil {
		output["data"] = data
	}
	if errMsg != "" {
		output["status"] = "error"
		output["error"] = errMsg
	}

	dataBytes, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(dataBytes))
}

// printMarkdown renders a human-readable summary. For AI agents, JSON is
// recommended, but markdown is the default for interactive use.
func printMarkdown(action string, success bool, data map[string]interface{}, errMsg string) {
	if !success {
		fmt.Printf("## %s\n- Status: error\n- Error: %s\n", titleCase(action), errMsg)
		return
	}

	fmt.Printf("## %s\n- Status: success\n", titleCase(action))
	if data != nil {
		printMarkdownData(data, 0)
	}
}

// printMarkdownData recursively renders data fields as markdown bullets.
func printMarkdownData(data map[string]interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)
	for key, val := range data {
		switch v := val.(type) {
		case map[string]interface{}:
			fmt.Printf("%s- %s:\n", prefix, key)
			printMarkdownData(v, indent+1)
		case []interface{}:
			fmt.Printf("%s- %s: [%d items]\n", prefix, key, len(v))
			for i, item := range v {
				if m, ok := item.(map[string]interface{}); ok {
					fmt.Printf("%s  [%d]:\n", prefix, i+1)
					printMarkdownData(m, indent+2)
				} else {
					fmt.Printf("%s  [%d] %v\n", prefix, i+1, item)
				}
			}
		case string:
			// Truncate long strings for readability
			s := v
			if len(s) > 200 {
				s = s[:200] + "..."
			}
			fmt.Printf("%s- %s: %s\n", prefix, key, s)
		default:
			fmt.Printf("%s- %s: %v\n", prefix, key, val)
		}
	}
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

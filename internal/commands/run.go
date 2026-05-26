package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <actions>",
	Short: "Execute multiple browser actions in a single session",
	Long: `Execute multiple browser actions sequentially.

The browser server is auto-started if not running.

ACTION SYNTAX:
  Actions are separated by semicolons (;). Each action is a command with arguments.
  
  Example: "navigate https://example.com; click #btn; text"

SUPPORTED ACTIONS:
  navigate <url>          - Navigate to URL
  click <selector>        - Click element (CSS selector)
  fill <selector> <value> - Fill input field with value
  type <selector> <text>  - Type text character by character
  select <selector> <val> - Select option from dropdown
  screenshot [file]       - Take screenshot (default: screenshot.png)
  text                    - Extract visible page text
  elements <selector>     - Find and list elements
  eval <javascript>       - Execute JavaScript
  wait <selector>         - Wait for element to appear
  scroll <direction>      - Scroll page (up or down)
  sleep <duration>        - Sleep for manual operations (e.g., 30s, 1m)
  back / forward / reload - Navigation controls

EXAMPLES:
  browser-cli run "navigate https://example.com; text; screenshot"
  browser-cli run "navigate https://login.com; fill '#email' 'user@test.com'; click '#submit'"
  browser-cli --output json run "navigate https://example.com; elements a"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		actionsStr := args[0]
		actions := parseActions(actionsStr)

		// Convert to server format
		actionList := make([]map[string]interface{}, 0)
		for _, action := range actions {
			parsed := parseActionToMap(action)
			if parsed != nil {
				actionList = append(actionList, parsed)
			}
		}

		return sendCommand("run", map[string]interface{}{
			"actions": actionList,
		})
	},
}

func parseActions(s string) []string {
	// Smart split: don't split semicolons inside eval command
	actions := make([]string, 0)

	// Find eval commands and protect their content
	evalStart := strings.Index(s, "eval ")
	if evalStart != -1 {
		// Find the end of eval (next semicolon outside of quotes, or end of string)
		evalContent := s[evalStart:]
		evalEnd := findEvalEnd(evalContent)

		// Split before eval
		beforeEval := s[:evalStart]
		if beforeEval != "" {
			parts := strings.Split(beforeEval, ";")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					actions = append(actions, p)
				}
			}
		}

		// Add eval as single action
		evalAction := strings.TrimSpace(evalContent[:evalEnd])
		if evalAction != "" {
			actions = append(actions, evalAction)
		}

		// Split after eval
		if evalEnd < len(evalContent) {
			afterEval := evalContent[evalEnd+1:]
			parts := strings.Split(afterEval, ";")
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					actions = append(actions, p)
				}
			}
		}
	} else {
		// No eval, simple split
		parts := strings.Split(s, ";")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				actions = append(actions, p)
			}
		}
	}

	return actions
}

// findEvalEnd finds the end of eval command content
func findEvalEnd(s string) int {
	inSingleQuote := false
	inDoubleQuote := false

	for i, c := range s {
		if c == '\'' && !inDoubleQuote {
			inSingleQuote = !inSingleQuote
		} else if c == '"' && !inSingleQuote {
			inDoubleQuote = !inDoubleQuote
		} else if c == ';' && !inSingleQuote && !inDoubleQuote {
			return i
		}
	}
	return len(s)
}

func parseActionToMap(action string) map[string]interface{} {
	parts := strings.Fields(action)
	if len(parts) == 0 {
		return nil
	}

	cmdName := parts[0]
	args := parts[1:]

	// Handle quoted arguments
	args = parseQuotedArgs(args)

	result := map[string]interface{}{
		"action": cmdName,
		"params": map[string]interface{}{},
	}

	params := result["params"].(map[string]interface{})

	switch cmdName {
	case "navigate":
		if len(args) < 1 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "navigate requires URL",
			}
		}
		params["url"] = args[0]

	case "click":
		if len(args) < 1 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "click requires selector",
			}
		}
		params["selector"] = args[0]

	case "fill":
		if len(args) < 2 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "fill requires selector and value",
			}
		}
		params["selector"] = args[0]
		params["value"] = args[1]

	case "type":
		if len(args) < 2 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "type requires selector and text",
			}
		}
		params["selector"] = args[0]
		params["text"] = args[1]
		params["delay"] = 50

	case "select":
		if len(args) < 2 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "select requires selector and value",
			}
		}
		params["selector"] = args[0]
		params["value"] = args[1]

	case "screenshot":
		path := "screenshot.png"
		if len(args) > 0 {
			path = args[0]
		}
		params["path"] = path

	case "text":
		// No params needed

	case "elements":
		if len(args) < 1 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "elements requires selector",
			}
		}
		params["selector"] = args[0]

	case "eval":
		if len(args) < 1 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "eval requires JavaScript",
			}
		}
		params["script"] = strings.Join(args, " ")

	case "wait":
		if len(args) < 1 {
			return map[string]interface{}{
				"action": cmdName,
				"params": map[string]interface{}{},
				"error":  "wait requires selector",
			}
		}
		params["selector"] = args[0]

	case "scroll":
		direction := "down"
		if len(args) > 0 {
			direction = args[0]
		}
		params["direction"] = direction
		params["distance"] = 300

	case "sleep":
		// Sleep is handled client-side, not sent to server
		duration := 30 * time.Second
		if len(args) > 0 {
			if d, err := time.ParseDuration(args[0]); err == nil {
				duration = d
			}
		}
		time.Sleep(duration)
		return nil // Skip sending to server

	case "back", "forward", "reload":
		// No params needed

	default:
		return map[string]interface{}{
			"action": cmdName,
			"params": map[string]interface{}{},
			"error":  fmt.Sprintf("unknown action: %s", cmdName),
		}
	}

	return result
}

// parseQuotedArgs handles quoted arguments like "hello world"
func parseQuotedArgs(args []string) []string {
	result := make([]string, 0)
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "'") || strings.HasPrefix(arg, "\"") {
			quote := arg[0]
			combined := arg[1:]
			i++
			for i < len(args) && !strings.HasSuffix(args[i], string(quote)) {
				combined += " " + args[i]
				i++
			}
			if i < len(args) {
				combined += " " + strings.TrimSuffix(args[i], string(quote))
				i++
			} else {
				combined = strings.TrimSuffix(combined, string(quote))
			}
			result = append(result, combined)
		} else {
			result = append(result, arg)
			i++
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(runCmd)
}
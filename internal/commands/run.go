package commands

import (
	"fmt"
	"os"
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
  upload <selector> <file> - Upload file to input
  pdf [file]              - Save page as PDF (Chromium only)
  keyboard <key>          - Press keyboard key/combo (e.g., Ctrl+A)
  right-click <selector>  - Right-click element
  dblclick <selector>     - Double-click element
  sleep <duration>        - Sleep for the given duration (e.g. 30s, 1m)
  back / forward / reload - Navigation controls

EXAMPLES:
  browser-cli run "navigate https://example.com; text; screenshot"
  browser-cli run "navigate https://login.com; fill '#email' 'user@test.com'; click '#submit'"
  browser-cli --output json run "navigate https://example.com; elements a"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		actionsStr := args[0]
		actions := parseActions(actionsStr)

		// Convert to server format. Parse errors are reported on stderr but do
		// not abort the whole pipeline: the run is a batch and one bad entry
		// shouldn't take down the rest.
		actionList := make([]map[string]interface{}, 0)
		for _, action := range actions {
			parsed, err := parseActionToMap(action)
			if err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "run: skipping action %q: %v\n", action, err)
				continue
			}
			actionList = append(actionList, parsed)
		}

		if len(actionList) == 0 {
			return fmt.Errorf("no valid actions to execute")
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

// parseActionToMap converts a single action string into a Command-shaped map.
// Returns an error (instead of embedding an "error" field in the map, which
// callers used to ignore) so parse problems are reported to the user.
func parseActionToMap(action string) (map[string]interface{}, error) {
	parts := strings.Fields(action)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty action")
	}

	cmdName := parts[0]
	args := parts[1:]

	// Handle quoted arguments. parseQuotedArgs only fails on unterminated
	// quotes, which is a programmer error in the action string — fail the
	// whole action with a useful message rather than silently truncating.
	args, err := parseQuotedArgs(args)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", cmdName, err)
	}

	result := map[string]interface{}{
		"action": cmdName,
		"params": map[string]interface{}{},
	}

	params := result["params"].(map[string]interface{})

	switch cmdName {
	case "navigate":
		if len(args) < 1 {
			return nil, fmt.Errorf("navigate requires URL")
		}
		params["url"] = args[0]

	case "click":
		if len(args) < 1 {
			return nil, fmt.Errorf("click requires selector")
		}
		params["selector"] = args[0]

	case "fill":
		if len(args) < 2 {
			return nil, fmt.Errorf("fill requires selector and value")
		}
		params["selector"] = args[0]
		params["value"] = args[1]

	case "type":
		if len(args) < 2 {
			return nil, fmt.Errorf("type requires selector and text")
		}
		params["selector"] = args[0]
		params["text"] = args[1]
		params["delay"] = 50

	case "select":
		if len(args) < 2 {
			return nil, fmt.Errorf("select requires selector and value")
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
			return nil, fmt.Errorf("elements requires selector")
		}
		params["selector"] = args[0]

	case "eval":
		if len(args) < 1 {
			return nil, fmt.Errorf("eval requires JavaScript")
		}
		params["script"] = strings.Join(args, " ")

	case "eval-file":
		if len(args) < 1 {
			return nil, fmt.Errorf("eval-file requires a path")
		}
		data, err := os.ReadFile(args[0])
		if err != nil {
			return nil, fmt.Errorf("eval-file: read %q: %w", args[0], err)
		}
		// eval-file is a client-side convenience: the server only knows
		// "eval". We swap the action name and pass the file contents in
		// the same "script" field, so the server path is identical.
		result["action"] = "eval"
		params["script"] = string(data)

	case "wait":
		if len(args) < 1 {
			return nil, fmt.Errorf("wait requires selector")
		}
		params["selector"] = args[0]

	case "scroll":
		direction := "down"
		if len(args) > 0 {
			direction = args[0]
		}
		params["direction"] = direction
		params["distance"] = 300

	case "upload":
		if len(args) < 2 {
			return nil, fmt.Errorf("upload requires selector and file path")
		}
		params["selector"] = args[0]
		params["path"] = args[1]

	case "pdf":
		path := "output.pdf"
		if len(args) > 0 {
			path = args[0]
		}
		params["path"] = path
		params["landscape"] = false
		params["format"] = "A4"

	case "keyboard":
		if len(args) < 1 {
			return nil, fmt.Errorf("keyboard requires key")
		}
		params["key"] = args[0]

	case "right-click":
		if len(args) < 1 {
			return nil, fmt.Errorf("right-click requires selector")
		}
		params["selector"] = args[0]

	case "dblclick":
		if len(args) < 1 {
			return nil, fmt.Errorf("dblclick requires selector")
		}
		params["selector"] = args[0]

	case "sleep":
		// Sleep runs server-side so that the wait happens inside the same
		// session's action queue, not on the client (where it would race
		// with the next request already in flight to the server).
		duration := 30 * time.Second
		if len(args) > 0 {
			d, err := time.ParseDuration(args[0])
			if err != nil {
				return nil, fmt.Errorf("sleep: invalid duration %q: %w", args[0], err)
			}
			duration = d
		}
		params["duration_ms"] = duration.Milliseconds()

	case "back", "forward", "reload":
		// No params needed

	default:
		return nil, fmt.Errorf("unknown action: %s", cmdName)
	}

	return result, nil
}

// parseQuotedArgs handles quoted arguments like "hello world". An unterminated
// quoted string now returns an error so that values aren't silently truncated.
func parseQuotedArgs(args []string) ([]string, error) {
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
				return nil, fmt.Errorf("unterminated quoted argument starting with %q", arg)
			}
			result = append(result, combined)
		} else {
			result = append(result, arg)
			i++
		}
	}
	return result, nil
}

func init() {
	rootCmd.AddCommand(runCmd)
}

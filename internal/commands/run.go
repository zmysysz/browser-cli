package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/browser-cli/internal/browser"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <actions>",
	Short: "Execute multiple browser actions in a single session",
	Long: `Execute multiple browser actions sequentially in a single browser session.

This is the RECOMMENDED way for AI agents to interact with the browser, because:
  1. The browser stays alive across all operations
  2. You can navigate, interact, and extract data in one call
  3. Results are returned as a structured array

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

SELECTOR SYNTAX:
  Use standard CSS selectors:
  • "#id"          - Element by ID
  • ".class"       - Elements by class
  • "tag"          - Elements by tag name
  • "[attr=val]"   - Elements by attribute
  • "a[href]"      - All links
  • "input[type=text]" - Text inputs

QUOTED VALUES:
  For values with spaces, use single or double quotes:
  fill '#search' 'hello world'
  fill "#search" "hello world"

EXAMPLES:
  # Basic navigation and extraction
  browser-cli run "navigate https://example.com; text; screenshot"
  
  # Form interaction
  browser-cli run "navigate https://login.example.com; fill '#email' 'user@test.com'; fill '#password' 'secret'; click '#submit'"
  
  # Click link and extract new page
  browser-cli run "navigate https://example.com; click 'a[href]'; wait 'body'; text"
  
  # Get structured JSON output (recommended for AI)
  browser-cli --output json run "navigate https://example.com; elements a; eval document.title"

OUTPUT:
  Returns an array of results, one for each action:
  • Each result has: step, action, status (success/error)
  • Successful results include action-specific data
  • Failed results include error message
  • Use --output json for structured parsing`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		actionsStr := args[0]
		actions := parseActions(actionsStr)

		b, err := getBrowser()
		if err != nil {
			printError("run", err)
			return
		}

		results := make([]map[string]interface{}, 0)

		for i, action := range actions {
			result := executeAction(b, action)
			result["step"] = i + 1
			results = append(results, result)
		}

		printSuccess("run", map[string]interface{}{
			"total_steps": len(results),
			"results":     results,
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
		// Find where eval ends - look for unquoted semicolon or end
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
	// eval content ends at next unquoted semicolon or end of string
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

func executeAction(b *browser.Browser, action string) map[string]interface{} {
	parts := strings.Fields(action)
	if len(parts) == 0 {
		return map[string]interface{}{
			"action": "",
			"status": "error",
			"error":  "empty action",
		}
	}

	cmdName := parts[0]
	args := parts[1:]

	// Handle quoted arguments
	args = parseQuotedArgs(args)

	result := map[string]interface{}{
		"action": action,
		"status": "success",
	}

	switch cmdName {
	case "navigate":
		if len(args) < 1 {
			result["status"] = "error"
			result["error"] = "navigate requires URL"
			return result
		}
		r, err := b.Navigate(args[0], timeout)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["url"] = r.URL
			result["title"] = r.Title
		}

	case "click":
		if len(args) < 1 {
			result["status"] = "error"
			result["error"] = "click requires selector"
			return result
		}
		if err := b.Click(args[0], timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "fill":
		if len(args) < 2 {
			result["status"] = "error"
			result["error"] = "fill requires selector and value"
			return result
		}
		if err := b.Fill(args[0], args[1], timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "type":
		if len(args) < 2 {
			result["status"] = "error"
			result["error"] = "type requires selector and text"
			return result
		}
		if err := b.Type(args[0], args[1], 50, timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "select":
		if len(args) < 2 {
			result["status"] = "error"
			result["error"] = "select requires selector and value"
			return result
		}
		if err := b.Select(args[0], args[1], timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "screenshot":
		path := "screenshot.png"
		if len(args) > 0 {
			path = args[0]
		}
		r, err := b.Screenshot(path)
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["path"] = r.Path
		}

	case "text":
		r, err := b.Text()
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["content"] = r.Text
		}

	case "elements":
		if len(args) < 1 {
			result["status"] = "error"
			result["error"] = "elements requires selector"
			return result
		}
		els, err := b.Elements(args[0])
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["count"] = len(els)
			result["items"] = els
		}

	case "eval":
		if len(args) < 1 {
			result["status"] = "error"
			result["error"] = "eval requires JavaScript"
			return result
		}
		r, err := b.Eval(strings.Join(args, " "))
		if err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		} else {
			result["value"] = r.Value
		}

	case "wait":
		if len(args) < 1 {
			result["status"] = "error"
			result["error"] = "wait requires selector"
			return result
		}
		if err := b.Wait(args[0], timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "scroll":
		direction := "down"
		if len(args) > 0 {
			direction = args[0]
		}
		if err := b.Scroll(direction, 300); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "sleep":
		// Sleep for manual operations (e.g., waiting for user to login)
		duration := 30 * time.Second
		if len(args) > 0 {
			if d, err := time.ParseDuration(args[0]); err == nil {
				duration = d
			}
		}
		time.Sleep(duration)
		result["duration"] = duration.String()

	case "back":
		if err := b.Back(timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "forward":
		if err := b.Forward(timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	case "reload":
		if err := b.Reload(timeout); err != nil {
			result["status"] = "error"
			result["error"] = err.Error()
		}

	default:
		result["status"] = "error"
		result["error"] = fmt.Sprintf("unknown action: %s", cmdName)
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
			// Find closing quote
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
				// Handle case where the same arg has both opening and closing quote
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
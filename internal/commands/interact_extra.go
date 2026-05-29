package commands

import (
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <selector> <file>",
	Short: "Upload a file to a file input element",
	Long: `Upload a file to a file input element on the page.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector for the file input element
  file     - Path to the file to upload

EXAMPLES:
  browser-cli upload "#file-input" ./document.pdf
  browser-cli upload "input[type=file]" /tmp/image.png
  browser-cli upload "#resume" ./resume.docx`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("upload", map[string]interface{}{
			"selector": args[0],
			"path":     args[1],
		})
	},
}

var pdfCmd = &cobra.Command{
	Use:   "pdf [file]",
	Short: "Save the current page as a PDF file (Chromium only)",
	Long: `Save the current page as a PDF file.

Note: PDF generation is only supported in Chromium-based browsers.

The browser server is auto-started if not running.

ARGUMENTS:
  file - Optional output file path (default: output.pdf)

FLAGS:
  --landscape - Use landscape orientation (default: false)
  --format    - Paper format: A4, Letter, etc. (default: A4)

EXAMPLES:
  browser-cli pdf
  browser-cli pdf report.pdf
  browser-cli pdf --landscape --format Letter page.pdf`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "output.pdf"
		if len(args) > 0 {
			path = args[0]
		}
		landscape, _ := cmd.Flags().GetBool("landscape")
		format, _ := cmd.Flags().GetString("format")
		return sendCommand("pdf", map[string]interface{}{
			"path":      path,
			"landscape": landscape,
			"format":    format,
		})
	},
}

var keyboardCmd = &cobra.Command{
	Use:   "keyboard <key>",
	Short: "Press a keyboard key or key combination",
	Long: `Press a keyboard key or key combination on the page.

The browser server is auto-started if not running.

ARGUMENTS:
  key - Key or key combination to press (e.g., "Enter", "Ctrl+A", "Escape")

SUPPORTED KEYS:
  • Single keys: Enter, Escape, Tab, Backspace, Delete, Space
  • Modifiers: Ctrl+, Alt+, Shift+, Meta+
  • Arrows: ArrowUp, ArrowDown, ArrowLeft, ArrowRight
  • Function keys: F1-F12
  • Combos: Ctrl+A, Ctrl+C, Ctrl+V, Ctrl+S, Alt+F4

EXAMPLES:
  browser-cli keyboard "Enter"
  browser-cli keyboard "Ctrl+A"
  browser-cli keyboard "Escape"
  browser-cli keyboard "Ctrl+Shift+I"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("keyboard", map[string]interface{}{
			"key": args[0],
		})
	},
}

var rightClickCmd = &cobra.Command{
	Use:   "right-click <selector>",
	Short: "Right-click an element on the page",
	Long: `Right-click an element identified by a CSS selector.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector to identify the element

EXAMPLES:
  browser-cli right-click "#menu-item"
  browser-cli right-click ".context-target"
  browser-cli right-click "td.cell"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("right-click", map[string]interface{}{
			"selector": args[0],
		})
	},
}

var dblclickCmd = &cobra.Command{
	Use:   "dblclick <selector>",
	Short: "Double-click an element on the page",
	Long: `Double-click an element identified by a CSS selector.

The browser server is auto-started if not running.

ARGUMENTS:
  selector - CSS selector to identify the element

EXAMPLES:
  browser-cli dblclick "#item"
  browser-cli dblclick ".folder"
  browser-cli dblclick "td.editable"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return sendCommand("dblclick", map[string]interface{}{
			"selector": args[0],
		})
	},
}

func init() {
	pdfCmd.Flags().Bool("landscape", false, "Use landscape orientation")
	pdfCmd.Flags().String("format", "A4", "Paper format (A4, Letter, etc.)")

	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(pdfCmd)
	rootCmd.AddCommand(keyboardCmd)
	rootCmd.AddCommand(rightClickCmd)
	rootCmd.AddCommand(dblclickCmd)
}

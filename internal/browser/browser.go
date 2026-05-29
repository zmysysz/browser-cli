package browser

import (
	"github.com/playwright-community/playwright-go"
)

// Browser wraps Playwright page for action helpers (used within Server's SessionState)
type Browser struct {
	page playwright.Page
}
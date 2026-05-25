package browser

import (
	"time"

	"github.com/playwright-community/playwright-go"
)

// Browser wraps Playwright browser instance
type Browser struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	context playwright.BrowserContext
	page    playwright.Page
}

// Config holds browser configuration
type Config struct {
	Browser  string // chromium, firefox, webkit
	Headless bool
	Timeout  time.Duration
}

// New creates a new browser instance
func New(cfg Config) (*Browser, error) {
	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	// Launch browser
	var browser playwright.Browser
	switch cfg.Browser {
	case "firefox":
		browser, err = pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	case "webkit":
		browser, err = pw.WebKit.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	default: // chromium
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(cfg.Headless),
		})
	}
	if err != nil {
		pw.Stop()
		return nil, err
	}

	// Create context and page
	context, err := browser.NewContext()
	if err != nil {
		browser.Close()
		pw.Stop()
		return nil, err
	}

	// Auto-load saved cookies
	cookies, err := GetCookieStorage().LoadAll()
	if err == nil && len(cookies) > 0 {
		// Convert Cookie to OptionalCookie
		optionalCookies := make([]playwright.OptionalCookie, len(cookies))
		for i, c := range cookies {
			optionalCookies[i] = c.ToOptionalCookie()
		}
		context.AddCookies(optionalCookies)
	}

	page, err := context.NewPage()
	if err != nil {
		context.Close()
		browser.Close()
		pw.Stop()
		return nil, err
	}

	return &Browser{
		pw:      pw,
		browser: browser,
		context: context,
		page:    page,
	}, nil
}

// Close closes the browser and cleans up
func (b *Browser) Close() error {
	// Auto-save cookies before closing
	cookies, err := b.context.Cookies()
	if err == nil && len(cookies) > 0 {
		GetCookieStorage().SaveAll(cookies)
	}

	if b.page != nil {
		b.page.Close()
	}
	if b.context != nil {
		b.context.Close()
	}
	if b.browser != nil {
		b.browser.Close()
	}
	if b.pw != nil {
		b.pw.Stop()
	}
	return nil
}

// Page returns the current page
func (b *Browser) Page() playwright.Page {
	return b.page
}

// SetTimeout sets the default timeout
func (b *Browser) SetTimeout(d time.Duration) {
	b.page.SetDefaultTimeout(float64(d.Milliseconds()))
}
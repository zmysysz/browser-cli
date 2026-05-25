package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// NavigateResult holds navigation result
type NavigateResult struct {
	URL     string
	Title   string
	Status  int
	LoadTime time.Duration
}

// Navigate navigates to a URL
func (b *Browser) Navigate(url string, timeout time.Duration) (*NavigateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	
	resp, err := b.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page to be ready
	b.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})

	title, _ := b.page.Title()

	result := &NavigateResult{
		URL:      b.page.URL(),
		Title:    title,
		LoadTime: time.Since(start),
	}

	if resp != nil {
		result.Status = int(resp.Status())
	}

	// Check context
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("navigation timeout")
	default:
		return result, nil
	}
}

// ScreenshotResult holds screenshot result
type ScreenshotResult struct {
	Path string
	Size int64
}

// Screenshot takes a screenshot
func (b *Browser) Screenshot(path string) (*ScreenshotResult, error) {
	if path == "" {
		path = "screenshot.png"
	}

	_, err := b.page.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
		FullPage: playwright.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	return &ScreenshotResult{
		Path: path,
	}, nil
}

// TextResult holds text extraction result
type TextResult struct {
	Text string
}

// Text extracts page text content
func (b *Browser) Text() (*TextResult, error) {
	text, err := b.page.TextContent("body")
	if err != nil {
		return nil, fmt.Errorf("text extraction failed: %w", err)
	}

	return &TextResult{
		Text: text,
	}, nil
}

// ElementInfo holds element information
type ElementInfo struct {
	Tag      string
	Text     string
	ID       string
	Class    string
	Href     string
	Visible  bool
}

// Elements finds elements by selector
func (b *Browser) Elements(selector string) ([]ElementInfo, error) {
	elements, err := b.page.QuerySelectorAll(selector)
	if err != nil {
		return nil, fmt.Errorf("element query failed: %w", err)
	}

	var results []ElementInfo
	for _, el := range elements {
		info := ElementInfo{}
		
		tag, _ := el.Evaluate("el => el.tagName", nil)
		if s, ok := tag.(string); ok {
			info.Tag = s
		}

		text, _ := el.TextContent()
		info.Text = text

		id, _ := el.GetAttribute("id")
		info.ID = id

		class, _ := el.GetAttribute("class")
		info.Class = class

		href, _ := el.GetAttribute("href")
		info.Href = href

		visible, _ := el.IsVisible()
		info.Visible = visible

		results = append(results, info)
	}

	return results, nil
}

// EvalResult holds JavaScript evaluation result
type EvalResult struct {
	Value interface{}
}

// Eval executes JavaScript
func (b *Browser) Eval(js string) (*EvalResult, error) {
	result, err := b.page.Evaluate(js)
	if err != nil {
		return nil, fmt.Errorf("eval failed: %w", err)
	}

	return &EvalResult{
		Value: result,
	}, nil
}
package browser

import (
	"fmt"
	"time"

	"github.com/playwright-community/playwright-go"
)

// Click clicks an element
func (b *Browser) Click(selector string, timeout time.Duration) error {
	err := b.page.Click(selector, playwright.PageClickOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("click failed: %w", err)
	}
	return nil
}

// Fill fills an input field
func (b *Browser) Fill(selector, value string, timeout time.Duration) error {
	err := b.page.Fill(selector, value, playwright.PageFillOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("fill failed: %w", err)
	}
	return nil
}

// Select selects an option from a dropdown
func (b *Browser) Select(selector, value string, timeout time.Duration) error {
	values := []string{value}
	_, err := b.page.SelectOption(selector, playwright.SelectOptionValues{
		Values: &values,
	}, playwright.PageSelectOptionOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("select failed: %w", err)
	}
	return nil
}

// Type types text into an element
func (b *Browser) Type(selector, text string, delay int, timeout time.Duration) error {
	err := b.page.Type(selector, text, playwright.PageTypeOptions{
		Delay:   playwright.Float(float64(delay)),
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("type failed: %w", err)
	}
	return nil
}

// Wait waits for an element
func (b *Browser) Wait(selector string, timeout time.Duration) error {
	_, err := b.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("wait failed: %w", err)
	}
	return nil
}

// Scroll scrolls the page
func (b *Browser) Scroll(direction string, distance int) error {
	var script string
	if direction == "down" {
		script = fmt.Sprintf("window.scrollBy(0, %d)", distance)
	} else {
		script = fmt.Sprintf("window.scrollBy(0, -%d)", distance)
	}

	_, err := b.page.Evaluate(script)
	if err != nil {
		return fmt.Errorf("scroll failed: %w", err)
	}
	return nil
}

// Back navigates back
func (b *Browser) Back(timeout time.Duration) error {
	_, err := b.page.GoBack(playwright.PageGoBackOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("back failed: %w", err)
	}
	return nil
}

// Forward navigates forward
func (b *Browser) Forward(timeout time.Duration) error {
	_, err := b.page.GoForward(playwright.PageGoForwardOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("forward failed: %w", err)
	}
	return nil
}

// Reload reloads the page
func (b *Browser) Reload(timeout time.Duration) error {
	_, err := b.page.Reload(playwright.PageReloadOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("reload failed: %w", err)
	}
	return nil
}
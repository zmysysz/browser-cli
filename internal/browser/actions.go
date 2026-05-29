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
		Force:   playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("click failed: %w", err)
	}
	return nil
}

// SmartClick intelligently clicks an element, handling Web Components automatically
// It tries multiple methods: regular click, Shadow DOM click, auto-detected callbacks, custom events
func (b *Browser) SmartClick(selector string, timeout time.Duration) error {
	script := `
		(selector) => {
			const element = document.querySelector(selector);
			if (!element) return { success: false, error: 'Element not found' };
			
			// Helper: auto-detect callable methods on element
			function findCallableMethods(obj) {
				const methods = [];
				const patterns = ['_on', '_handle', 'handle', 'on', '_click', '_submit', '_action'];
				for (const key of Object.keys(obj)) {
					if (typeof obj[key] === 'function') {
						for (const pattern of patterns) {
							if (key.toLowerCase().startsWith(pattern.toLowerCase())) {
								methods.push(key);
								break;
							}
						}
					}
				}
				return methods;
			}
			
			// Method 1: Regular click
			try {
				element.click();
				if (element.onclick || element.hasAttribute('onclick')) {
					return { success: true, method: 'regular_click' };
				}
			} catch (e) {}
			
			// Method 2: Shadow DOM click - find button inside shadow root
			if (element.shadowRoot) {
				const innerBtn = element.shadowRoot.querySelector('button, [role="button"], .btn, [type="button"]');
				if (innerBtn) {
					try {
						innerBtn.click();
						return { success: true, method: 'shadow_dom_click' };
					} catch (e) {}
				}
			}
			
			// Method 3: Auto-detect and call internal methods
			const methods = findCallableMethods(element);
			for (const method of methods) {
				try {
					element[method]();
					return { success: true, method: 'auto_detected_' + method };
				} catch (e) {}
			}
			
			// Method 4: Check for internal button/light DOM
			const innerBtn = element.querySelector('button, [role="button"], .btn');
			if (innerBtn) {
				try {
					innerBtn.click();
					return { success: true, method: 'inner_button_click' };
				} catch (e) {}
			}
			
			// Method 5: Dispatch custom events
			const events = ['click', 'tap', 'action', 'submit'];
			for (const event of events) {
				try {
					element.dispatchEvent(new CustomEvent(event, { bubbles: true, composed: true }));
					element.dispatchEvent(new MouseEvent(event, { bubbles: true }));
				} catch (e) {}
			}
			
			// Method 6: Trigger framework click if detected
			if (element.__vue__ || element._reactInternals) {
				try {
					element.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true }));
					return { success: true, method: 'framework_click' };
				} catch (e) {}
			}
			
			return { success: true, method: 'all_methods_attempted', detected_methods: methods };
		}
	`
	
	result, err := b.page.Evaluate(script)
	if err != nil {
		return fmt.Errorf("smart click evaluation failed: %w", err)
	}
	
	// Parse result
	if resultMap, ok := result.(map[string]interface{}); ok {
		if success, ok := resultMap["success"].(bool); ok && !success {
			if errMsg, ok := resultMap["error"].(string); ok {
				return fmt.Errorf("smart click failed: %s", errMsg)
			}
		}
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

// Upload sets files on a file input element
func (b *Browser) Upload(selector, path string, timeout time.Duration) error {
	el, err := b.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	if err != nil {
		return fmt.Errorf("element not found: %w", err)
	}
	if err := el.SetInputFiles(path); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}
	return nil
}

// PDF saves the current page as a PDF file (Chromium only)
func (b *Browser) PDF(path string, landscape bool, format string) error {
	_, err := b.page.PDF(playwright.PagePdfOptions{
		Path:      playwright.String(path),
		Landscape: playwright.Bool(landscape),
		Format:    playwright.String(format),
	})
	if err != nil {
		return fmt.Errorf("PDF generation failed (Chromium only): %w", err)
	}
	return nil
}

// KeyboardPress presses a keyboard key or key combination
func (b *Browser) KeyboardPress(key string) error {
	err := b.page.Keyboard().Press(key)
	if err != nil {
		return fmt.Errorf("keyboard press failed: %w", err)
	}
	return nil
}
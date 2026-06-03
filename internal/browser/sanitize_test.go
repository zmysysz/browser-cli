package browser

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeError_NilReturnsEmpty(t *testing.T) {
	if got := sanitizeError(nil); got != "" {
		t.Errorf("sanitizeError(nil) = %q, want empty string", got)
	}
}

func TestSanitizeError_TruncatesAtNewline(t *testing.T) {
	in := errors.New("first line\nsecond line with /etc/passwd inside")
	got := sanitizeError(in)
	if strings.Contains(got, "\n") {
		t.Errorf("expected single-line output, got %q", got)
	}
	if !strings.HasPrefix(got, "first line") {
		t.Errorf("expected prefix 'first line', got %q", got)
	}
	if strings.Contains(got, "second line") {
		t.Errorf("expected newline content stripped, got %q", got)
	}
}

func TestSanitizeError_StripsAbsolutePaths(t *testing.T) {
	// Playwright sometimes embeds cache paths like
	// "/home/user/.cache/ms-playwright/chromium-1234/chrome-linux/..."
	cases := []struct {
		name string
		in   string
		want string // substring that must NOT appear
	}{
		{
			name: "unix cache path",
			in:   "browser launch failed: /home/alice/.cache/ms-playwright/chromium-1193/chrome-linux/chrome: exec format error",
			want: "/home/alice/.cache",
		},
		{
			name: "windows path",
			in:   `failed to launch: C:\Users\bob\AppData\Local\ms-playwright\chromium-1193\chrome.exe: access denied`,
			want: `C:\Users\bob`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeError(errors.New(tc.in))
			if strings.Contains(got, tc.want) {
				t.Errorf("path %q leaked through sanitize: %q", tc.want, got)
			}
		})
	}
}

func TestSanitizeError_StripsGoLineRefs(t *testing.T) {
	in := errors.New("nil page: internal/browser/server.go:519 in eval()")
	got := sanitizeError(in)
	if strings.Contains(got, "server.go:519") {
		t.Errorf("go line ref leaked through sanitize: %q", got)
	}
}

func TestSanitizeError_PreservesUsefulMessage(t *testing.T) {
	// The whole point is to keep the action's failure context. Make
	// sure a clean playwright-style message still reads correctly.
	in := errors.New("element not found: timeout 30000ms exceeded")
	got := sanitizeError(in)
	if got != "element not found: timeout 30000ms exceeded" {
		t.Errorf("clean message got mangled: %q", got)
	}
}

func TestSanitizeError_KeepsHTTPSURLs(t *testing.T) {
	// The path-stripping regex intentionally avoids /https?:\/\// by
	// requiring the character after the path to be whitespace, colon,
	// or quote. Verify a URL in an error message survives.
	in := errors.New("navigate to https://example.com blocked: net::ERR_BLOCKED_BY_CLIENT")
	got := sanitizeError(in)
	if !strings.Contains(got, "https://example.com") {
		t.Errorf("URL got stripped from error: %q", got)
	}
}

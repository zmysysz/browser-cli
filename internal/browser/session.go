package browser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// SessionCookieStorage manages cookie persistence for a specific session
type SessionCookieStorage struct {
	basePath string
}

// NewSessionCookieStorage creates a cookie storage for a specific session
func NewSessionCookieStorage(sessionID string) *SessionCookieStorage {
	if sessionID == "" {
		sessionID = "default"
	}
	basePath := filepath.Join("/tmp", "browser-cli", "cookies", sessionID)
	os.MkdirAll(basePath, 0755)
	return &SessionCookieStorage{basePath: basePath}
}

// safeDomain converts a cookie domain into something safe to use as a
// filename. Real domains are usually fine ("example.com"), but cookies
// can carry ports in their Domain field (e.g. "localhost:8080") and a
// colon is not legal in filenames on Windows. We replace `:` with `_`.
// The original domain is preserved inside the JSON payload, so list and
// load still report the right thing.
func safeDomain(domain string) string {
	return strings.ReplaceAll(domain, ":", "_")
}

// SaveAll saves all cookies to storage (grouped by domain)
func (cs *SessionCookieStorage) SaveAll(cookies []playwright.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}

	// Group cookies by domain
	domainCookies := make(map[string][]playwright.Cookie)
	for _, c := range cookies {
		domain := c.Domain
		if len(domain) > 0 && domain[0] == '.' {
			domain = domain[1:]
		}
		domainCookies[domain] = append(domainCookies[domain], c)
	}

	for domain, cookies := range domainCookies {
		path := filepath.Join(cs.basePath, safeDomain(domain)+".json")
		data, err := json.MarshalIndent(cookies, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

// LoadAll loads all cookies from storage
func (cs *SessionCookieStorage) LoadAll() ([]playwright.Cookie, error) {
	var allCookies []playwright.Cookie

	files, err := filepath.Glob(filepath.Join(cs.basePath, "*.json"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var cookies []playwright.Cookie
		if err := json.Unmarshal(data, &cookies); err != nil {
			continue
		}
		allCookies = append(allCookies, cookies...)
	}

	return allCookies, nil
}

// Clear clears cookies for a specific domain or all domains
func (cs *SessionCookieStorage) Clear(domain string) error {
	if domain == "" {
		files, err := filepath.Glob(filepath.Join(cs.basePath, "*.json"))
		if err != nil {
			return err
		}
		for _, file := range files {
			os.Remove(file)
		}
		return nil
	}

	path := filepath.Join(cs.basePath, safeDomain(domain)+".json")
	os.Remove(path)
	path = filepath.Join(cs.basePath, safeDomain("."+domain)+".json")
	os.Remove(path)
	return nil
}

// List returns all saved cookie domains
func (cs *SessionCookieStorage) List() ([]CookieInfo, error) {
	files, err := filepath.Glob(filepath.Join(cs.basePath, "*.json"))
	if err != nil {
		return nil, err
	}

	var infos []CookieInfo
	for _, file := range files {
		domain := filepath.Base(file)
		domain = domain[:len(domain)-5]

		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var cookies []playwright.Cookie
		if err := json.Unmarshal(data, &cookies); err != nil {
			continue
		}

		infos = append(infos, CookieInfo{
			Domain: domain,
			Count:  len(cookies),
		})
	}
	return infos, nil
}

// CookieInfo represents cookie metadata for a domain
type CookieInfo struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

// GetSessionDir returns the session directory (for compatibility)
func GetSessionDir(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join("/tmp", "browser-cli", "sessions", sessionID)
}

package browser

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/playwright-community/playwright-go"
)

// CookieStorage manages cookie persistence (auto-save/load like real browser)
type CookieStorage struct {
	basePath string
}

// NewCookieStorage creates a new cookie storage
func NewCookieStorage() *CookieStorage {
	// Always use /tmp directory
	basePath := filepath.Join("/tmp", "browser-cli", "cookies")
	os.MkdirAll(basePath, 0755)
	return &CookieStorage{basePath: basePath}
}

// SaveAll saves all cookies to storage (grouped by domain)
func (cs *CookieStorage) SaveAll(cookies []playwright.Cookie) error {
	if len(cookies) == 0 {
		return nil
	}

	// Group cookies by domain
	domainCookies := make(map[string][]playwright.Cookie)
	for _, c := range cookies {
		domain := c.Domain
		// Remove leading dot if present
		if len(domain) > 0 && domain[0] == '.' {
			domain = domain[1:]
		}
		domainCookies[domain] = append(domainCookies[domain], c)
	}

	// Save each domain's cookies
	for domain, cookies := range domainCookies {
		path := filepath.Join(cs.basePath, domain+".json")
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
func (cs *CookieStorage) LoadAll() ([]playwright.Cookie, error) {
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
func (cs *CookieStorage) Clear(domain string) error {
	if domain == "" {
		// Clear all cookies
		files, err := filepath.Glob(filepath.Join(cs.basePath, "*.json"))
		if err != nil {
			return err
		}
		for _, file := range files {
			os.Remove(file)
		}
		return nil
	}

	// Clear specific domain (try both with and without leading dot)
	path := filepath.Join(cs.basePath, domain+".json")
	os.Remove(path)
	// Also try with leading dot
	path = filepath.Join(cs.basePath, "."+domain+".json")
	os.Remove(path)
	return nil
}

// List returns all saved cookie domains
func (cs *CookieStorage) List() ([]CookieInfo, error) {
	files, err := filepath.Glob(filepath.Join(cs.basePath, "*.json"))
	if err != nil {
		return nil, err
	}

	var infos []CookieInfo
	for _, file := range files {
		domain := filepath.Base(file)
		domain = domain[:len(domain)-5] // remove .json

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

// Global cookie storage
var globalCookieStorage = NewCookieStorage()

// GetCookieStorage returns the global cookie storage
func GetCookieStorage() *CookieStorage {
	return globalCookieStorage
}

// GetSessionDir returns the session directory
func GetSessionDir(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join("/tmp", "browser-cli", "sessions", sessionID)
}

// ListSessions returns all active sessions
func ListSessions() ([]string, error) {
	sessionsDir := filepath.Join("/tmp", "browser-cli", "sessions")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sessions []string
	for _, entry := range entries {
		if entry.IsDir() {
			socketPath := filepath.Join(sessionsDir, entry.Name(), "server.sock")
			if _, err := os.Stat(socketPath); err == nil {
				sessions = append(sessions, entry.Name())
			}
		}
	}
	return sessions, nil
}
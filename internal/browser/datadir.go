package browser

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// DataDir is the base directory for all browser-cli persistent state
// (socket, cookies, sessions, state files). It is resolved once at init
// time from --data-dir flag, BROWSER_CLI_HOME env, or the default.
//
// Default priority:
//  1. --data-dir flag (if set)
//  2. $BROWSER_CLI_HOME env var
//  3. $XDG_DATA_HOME/browser-cli (or ~/.local/share/browser-cli)
//  4. /tmp/browser-cli (fallback for environments without a home dir)
var DataDir string

// resolveDataDir determines the data directory using the same priority
// as XDG Base Directory spec. Called from init() so all packages see
// the resolved path via DataDir.
func resolveDataDir(flagDir string) string {
	// 1. Explicit flag
	if flagDir != "" {
		return flagDir
	}

	// 2. Environment variable
	if env := os.Getenv("BROWSER_CLI_HOME"); env != "" {
		return env
	}

	// 3. XDG data home
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "browser-cli")
	}

	// 4. ~/.local/share/browser-cli
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".local", "share", "browser-cli")
	}

	// 5. Fallback (no home dir, e.g. container running as nobody)
	return filepath.Join(os.TempDir(), "browser-cli")
}

// InitDataDir resolves and creates the data directory. Called once at startup.
func InitDataDir(flagDir string) error {
	DataDir = resolveDataDir(flagDir)
	if err := os.MkdirAll(DataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory %s: %w", DataDir, err)
	}
	return nil
}

// SocketPath returns the unix socket path inside the data directory.
func SocketPath() string {
	return filepath.Join(DataDir, "server.sock")
}

// CookieDir returns the cookie storage directory for a session.
func CookieDir(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(DataDir, "cookies", sessionID)
}

// SessionDir returns the session data directory.
func SessionDir(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(DataDir, "sessions", sessionID)
}

// StateFilePath returns the default storage state path for a session.
func StateFilePath(sessionID string) string {
	if sessionID == "" {
		sessionID = "default"
	}
	return filepath.Join(DataDir, "state", sessionID+".json")
}

// DefaultStatePath returns the default state file path (backward compat).
func DefaultStatePath() string {
	return StateFilePath("default")
}

// GetSocketPath returns the socket path. Kept for backward compatibility;
// new code should use SocketPath() directly.
//
// The sessionID parameter is ignored - there is a single socket for all
// sessions. It remains in the signature to avoid breaking callers.
func GetSocketPath(sessionID string) string {
	return SocketPath()
}

// GetSessionDir returns the session directory. Backward compat wrapper.
func GetSessionDir(sessionID string) string {
	return SessionDir(sessionID)
}

// Logger is the structured logger used throughout the server.
// Defaults to a text handler writing to stderr at Info level.
// Call InitLogger to reconfigure.
var Logger *slog.Logger

// LogFormat controls log output format.
type LogFormat string

const (
	LogFormatText LogFormat = "text"
	LogFormatJSON LogFormat = "json"
)

// InitLogger configures the global logger.
func InitLogger(format LogFormat, level slog.Level) {
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if format == LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// init sets up sensible defaults so packages that use Logger before
// InitLogger is called still get output.
func init() {
	Logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// maskPath replaces the data directory prefix in a string with a
// placeholder, so logs don't leak absolute home paths.
func maskPath(p string) string {
	if DataDir != "" {
		return strings.ReplaceAll(p, DataDir, "<datadir>")
	}
	return p
}

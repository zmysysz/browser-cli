# Changelog

All notable changes to browser-cli are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-06-03

### Added

- `eval-file <path>` command — read JavaScript from a file and execute
  it in the browser context. The recommended way to run multi-line
  scripts that contain single quotes, double quotes, or other awkward
  characters. Client-side alias: reads the file, sends it as a normal
  `eval` action, server unchanged. Five unit tests cover happy path,
  missing path, unreadable file, and multi-line IIFE preservation.
- `session-close` command — close the current session and persist
  cookies (was previously only reachable via `stop` which closed
  the whole server).
- `server-start` / `status` / `run` commands are now documented in
  the README All Commands table (they were always implemented, just
  missing from the docs).

### Changed

- **Concurrency model** rewritten. The previous design held a single
  `Server.mu` for the entire command dispatch path, which serialised
  every command across every session. Commands targeting different
  sessions now run in parallel; commands targeting the same session
  serialise on a per-session `SessionState.mu` that is held only for
  in-memory reads and writes, never across a playwright round-trip.
  See `.gal/design.md` for the full concurrency model and the rules
  handlers must follow.
- `idleMonitor` ticker period is now adaptive:
  `max(1s, min(60s, timeout/10))`. A 30-second timeout now polls
  every 3s (previously 1 minute, an undercount of 100-200%); a 24h
  timeout polls at the 60s ceiling. State is read via `atomic.Bool`
  / `atomic.Int64`, so the monitor never contends with dispatch.
- `Stop()` and `session_close` now snapshot the session map under
  `sessionsMu`, clear the map inside the lock, and run I/O
  (cookie save, page close, context close) outside the lock.
- `tab-close` response now includes a `new_active_tab` field so
  callers can update their local state without a follow-up
  `tab-list` call.
- `status` response now emits `last_activity` as an RFC3339 UTC
  timestamp (was a Go time.Time default format).
- `lastActivity` is now touched exactly once per command (was
  redundantly touched in both `handleConnection` and the
  dispatcher).

### Fixed

- Errors returned via `fail(err)` are now sanitised: multi-line
  playwright stack traces are truncated to the first line, absolute
  filesystem paths are replaced with `<path>`, Go-style
  `file.go:123` references are replaced with `<loc>`, while URLs
  (`http://`, `https://`, `file://`, `data://`) are preserved
  intact. This prevents leaking internal cache paths, user home
  directories, and Go source locations to the client. Six unit
  tests cover nil, multi-line, Unix path, Windows path, Go line
  ref, and URL preservation cases.
- A redundant `time.Sleep(100ms)` race workaround in the server
  shutdown path was replaced with `sync.Once` + `connWG.Wait()` in
  an earlier commit; v0.2.0 carries the final cleanup that removes
  the now-stale comment.

### Tests

- `tests/regression.sh` — 32 end-to-end checks (was 31 in v0.1.0,
  one added for `eval-file`). All pass against a freshly built
  binary on Linux.
- `internal/browser/sanitize_test.go` — 6 unit tests for the new
  error sanitiser.
- `internal/commands/run_test.go` — 5 unit tests for the new
  `eval-file` client-side alias.
- Existing `params_test.go` and `lifecycle_test.go` continue to
  pass under `-race`.

## [0.1.0] - 2026-06-02

### Added

- Initial release. Core browser automation commands
  (navigate / back / forward / reload, click / click-js / smart-click,
  hover / fill / type / select / keyboard / upload, screenshot / text
  / elements / eval / pdf / pick), tab management
  (tab-new / tab-list / tab-switch / tab-close), dialog detection
  (dialog-status / dialog-accept / dialog-dismiss), cookie
  persistence, multi-session isolation via `--session`, proxy
  support via `--proxy`, batch execution via `run`, and an
  auto-managed Unix-socket server.
- AI agent integration files for Claude Code, OpenAI Codex, GAL,
  and `AGENTS.md` for generic agents.

[Unreleased]: https://github.com/zmysysz/browser-cli/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/zmysysz/browser-cli/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/zmysysz/browser-cli/releases/tag/v0.1.0

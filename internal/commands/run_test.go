package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseActionToMap_Eval(t *testing.T) {
	got, err := parseActionToMap(`eval document.title`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["action"] != "eval" {
		t.Errorf("action = %q, want %q", got["action"], "eval")
	}
	params := got["params"].(map[string]interface{})
	if params["script"] != "document.title" {
		t.Errorf("script = %q, want %q", params["script"], "document.title")
	}
}

func TestParseActionToMap_EvalFile(t *testing.T) {
	dir := t.TempDir()
	jsPath := filepath.Join(dir, "snippet.js")
	want := "JSON.stringify({title: document.title})"
	if err := os.WriteFile(jsPath, []byte(want), 0644); err != nil {
		t.Fatalf("write tmp js: %v", err)
	}

	got, err := parseActionToMap("eval-file " + jsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// eval-file is a client-side alias: we keep the same wire shape as
	// eval, so the server doesn't need a new case statement.
	if got["action"] != "eval" {
		t.Errorf("action = %q, want %q (eval-file should map to eval)", got["action"], "eval")
	}
	params := got["params"].(map[string]interface{})
	if params["script"] != want {
		t.Errorf("script = %q, want %q", params["script"], want)
	}
}

func TestParseActionToMap_EvalFile_MissingPath(t *testing.T) {
	_, err := parseActionToMap("eval-file")
	if err == nil {
		t.Fatal("expected error for missing path, got nil")
	}
	if !strings.Contains(err.Error(), "requires a path") {
		t.Errorf("error = %q, want substring %q", err.Error(), "requires a path")
	}
}

func TestParseActionToMap_EvalFile_BadPath(t *testing.T) {
	_, err := parseActionToMap("eval-file /nonexistent/path/never-exists.js")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Errorf("error = %q, want substring %q", err.Error(), "read")
	}
}

func TestParseActionToMap_EvalFile_MultiLine(t *testing.T) {
	// Real-world shape: IIFE that returns a JSON string. Make sure newlines
	// and single quotes survive intact (the whole reason eval-file exists).
	dir := t.TempDir()
	jsPath := filepath.Join(dir, "multi.js")
	js := "(function(){\n  var t = document.querySelector('h1');\n  return JSON.stringify({h1: t ? t.innerText : null});\n})()"
	if err := os.WriteFile(jsPath, []byte(js), 0644); err != nil {
		t.Fatalf("write tmp js: %v", err)
	}

	got, err := parseActionToMap("eval-file " + jsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	params := got["params"].(map[string]interface{})
	if params["script"] != js {
		t.Errorf("multi-line script not preserved verbatim\nwant: %q\ngot:  %q", js, params["script"])
	}
}

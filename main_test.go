package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return dir
}

func createPart(t *testing.T, home, facet, name, content string) {
	t.Helper()
	dir := filepath.Join(home, ".prompt-cli", "parts", facet)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func createTemplate(t *testing.T, home, name, yamlContent string) {
	t.Helper()
	dir := filepath.Join(home, ".prompt-cli", "templates", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "template.yaml"), []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := fn()
	w.Close()
	os.Stdout = old
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	return string(buf[:n]), err
}

// --- init ---

func TestCmdInit_AlreadyExists(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "existing", "persona: '@x'\n")

	err := cmdInit("existing")
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

// --- run ---

func TestCmdRun_PartRef(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "helper", "You are helpful.")
	createTemplate(t, home, "t", "persona: '@helper'\n")

	out, err := captureStdout(t, func() error { return cmdRun("t") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>\nYou are helpful.\n</persona>") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestCmdRun_InlineText(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "persona: Custom bot\n")

	out, err := captureStdout(t, func() error { return cmdRun("t") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>\nCustom bot\n</persona>") {
		t.Errorf("unexpected output:\n%s", out)
	}
}

func TestCmdRun_AllFacets(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "helper", "Helper.")
	createTemplate(t, home, "t", `persona: '@helper'
policy: Be concise.
instruction: '@summarize'
output-contract: Output markdown.
`)

	out, err := captureStdout(t, func() error { return cmdRun("t") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>\nHelper.\n</persona>") {
		t.Error("expected persona block")
	}
	if !strings.Contains(out, "<policy>\nBe concise.\n</policy>") {
		t.Error("expected inline policy block")
	}
}

func TestCmdRun_SkipsEmpty(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "bot", "Bot.")
	createTemplate(t, home, "t", "persona: '@bot'\n")

	out, err := captureStdout(t, func() error { return cmdRun("t") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>") {
		t.Error("expected persona")
	}
	if strings.Contains(out, "<policy>") {
		t.Error("expected no policy")
	}
}

func TestCmdRun_MissingPart(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "persona: '@nonexistent'\n")

	err := cmdRun("t")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}
}

func TestCmdRun_TemplateNotFound(t *testing.T) {
	setupTestEnv(t)
	err := cmdRun("nope")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}
}

func TestCmdRun_WithBuiltinParts(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "persona: '@software-engineer'\npolicy: '@concise'\n")

	out, err := captureStdout(t, func() error { return cmdRun("t") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>") || !strings.Contains(out, "<policy>") {
		t.Errorf("expected builtin content:\n%s", out)
	}
}

// --- delete / list ---

func TestCmdDelete(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "")
	if err := cmdDelete("t"); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(home, ".prompt-cli", "templates", "t")
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("expected deleted")
	}
}

func TestCmdList(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "a", "")
	createTemplate(t, home, "b", "")

	out, err := captureStdout(t, func() error { return cmdList() })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "a") || !strings.Contains(out, "b") {
		t.Errorf("expected both templates:\n%s", out)
	}
}

// --- parts ---

func TestCmdPartList_All(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "custom", "x")

	out, err := captureStdout(t, func() error { return cmdPartList("") })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "persona/custom") {
		t.Error("expected user part")
	}
	if !strings.Contains(out, "software-engineer (builtin)") {
		t.Error("expected builtin part")
	}
}

func TestCmdPartDelete(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "x", "content")
	if err := cmdPartDelete("persona", "x"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(partPath("persona", "x")); !os.IsNotExist(err) {
		t.Error("expected deleted")
	}
}

// --- resolve ---

func TestResolveFacetContent(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "test", "Part content.")

	c, err := resolveFacetContent("@test", "persona")
	if err != nil || c != "Part content." {
		t.Errorf("part: got %q, %v", c, err)
	}

	c, err = resolveFacetContent("Inline text", "persona")
	if err != nil || c != "Inline text" {
		t.Errorf("inline: got %q, %v", c, err)
	}

	c, err = resolveFacetContent("", "persona")
	if err != nil || c != "" {
		t.Errorf("empty: got %q, %v", c, err)
	}

	c, err = resolveFacetContent("@software-engineer", "persona")
	if err != nil || c == "" {
		t.Errorf("builtin: got %q, %v", c, err)
	}

	_, err = resolveFacetContent("@nonexistent", "persona")
	if err == nil {
		t.Error("expected error for missing part")
	}
}

// --- validate ---

func TestValidateTemplate(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "x", "content")

	if err := validateTemplate(Template{Persona: "@x"}); err != nil {
		t.Errorf("valid part: %v", err)
	}
	if err := validateTemplate(Template{Persona: "inline text"}); err != nil {
		t.Errorf("inline text: %v", err)
	}
	if err := validateTemplate(Template{Persona: "@software-engineer"}); err != nil {
		t.Errorf("builtin: %v", err)
	}
	if err := validateTemplate(Template{}); err != nil {
		t.Errorf("empty: %v", err)
	}
	if err := validateTemplate(Template{Persona: "@nonexistent"}); err == nil {
		t.Error("expected error for missing part")
	}
}

// --- helpers ---

func TestIsValidFacet(t *testing.T) {
	for _, f := range validFacets {
		if !isValidFacet(f) {
			t.Errorf("%q should be valid", f)
		}
	}
	if isValidFacet("invalid") {
		t.Error("'invalid' should be invalid")
	}
}


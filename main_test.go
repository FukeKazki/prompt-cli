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

func TestCmdInit_WithParts(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "engineer", "You are an engineer.")

	if err := cmdInit("my-tmpl", []string{"persona=engineer"}); err != nil {
		t.Fatal(err)
	}

	tmpl, err := loadTemplate("my-tmpl")
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Persona != "@engineer" {
		t.Errorf("expected @engineer, got %q", tmpl.Persona)
	}
}

func TestCmdInit_EmptyTemplate(t *testing.T) {
	home := setupTestEnv(t)
	if err := cmdInit("empty", []string{}); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".prompt-cli", "templates", "empty", "template.yaml")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected template.yaml to exist")
	}
}

func TestCmdInit_AlreadyExists(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "existing", "persona: '@x'\n")

	err := cmdInit("existing", nil)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestCmdInit_InvalidPart(t *testing.T) {
	setupTestEnv(t)
	err := cmdInit("bad", []string{"persona=nonexistent"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCmdInit_WithBuiltinPart(t *testing.T) {
	setupTestEnv(t)
	if err := cmdInit("b", []string{"persona=code-reviewer"}); err != nil {
		t.Fatal(err)
	}
	tmpl, _ := loadTemplate("b")
	if tmpl.Persona != "@code-reviewer" {
		t.Errorf("expected @code-reviewer, got %q", tmpl.Persona)
	}
}

// --- run ---

func TestCmdRun_PartRef(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "helper", "You are helpful.")
	createTemplate(t, home, "t", "persona: '@helper'\n")

	out, err := captureStdout(t, func() error { return cmdRun("t", nil) })
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

	out, err := captureStdout(t, func() error { return cmdRun("t", nil) })
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

	out, err := captureStdout(t, func() error { return cmdRun("t", nil) })
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

	out, err := captureStdout(t, func() error { return cmdRun("t", nil) })
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

	err := cmdRun("t", nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}
}

func TestCmdRun_TemplateNotFound(t *testing.T) {
	setupTestEnv(t)
	err := cmdRun("nope", nil)
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}
}

func TestCmdRun_Override(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "helper", "Helper.")
	createTemplate(t, home, "t", "persona: '@helper'\n")

	out, err := captureStdout(t, func() error {
		return cmdRun("t", []string{"instruction=Do something"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>\nHelper.\n</persona>") {
		t.Error("expected persona from template")
	}
	if !strings.Contains(out, "<instruction>\nDo something\n</instruction>") {
		t.Error("expected instruction from override")
	}
}

func TestCmdRun_OverrideReplaces(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "helper", "Helper.")
	createTemplate(t, home, "t", "persona: '@helper'\n")

	out, err := captureStdout(t, func() error {
		return cmdRun("t", []string{"persona=Custom"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Custom") {
		t.Error("expected override content")
	}
	if strings.Contains(out, "Helper.") {
		t.Error("expected original to be replaced")
	}
}

func TestCmdRun_OverrideWithPartRef(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "persona: '@code-reviewer'\n")

	out, err := captureStdout(t, func() error {
		return cmdRun("t", []string{"persona=@software-engineer"})
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "シニアソフトウェアエンジニア") {
		t.Errorf("expected software-engineer content, got:\n%s", out)
	}
}

func TestCmdRun_OverrideInvalidFacet(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "")
	err := cmdRun("t", []string{"unknown=value"})
	if err == nil || !strings.Contains(err.Error(), "unknown facet") {
		t.Errorf("expected 'unknown facet', got: %v", err)
	}
}

func TestCmdRun_WithBuiltinParts(t *testing.T) {
	home := setupTestEnv(t)
	createTemplate(t, home, "t", "persona: '@software-engineer'\npolicy: '@concise'\n")

	out, err := captureStdout(t, func() error { return cmdRun("t", nil) })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "<persona>") || !strings.Contains(out, "<policy>") {
		t.Errorf("expected builtin content:\n%s", out)
	}
}

// --- edit ---

func TestCmdEdit_UpdatesFacet(t *testing.T) {
	home := setupTestEnv(t)
	createPart(t, home, "persona", "new", "New.")
	createTemplate(t, home, "t", "persona: '@code-reviewer'\n")

	if err := cmdEdit("t", []string{"persona=new"}); err != nil {
		t.Fatal(err)
	}
	tmpl, _ := loadTemplate("t")
	if tmpl.Persona != "@new" {
		t.Errorf("expected @new, got %q", tmpl.Persona)
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

func TestParseFacetArgs(t *testing.T) {
	m, err := parseFacetArgs([]string{"persona=x", "policy=y"})
	if err != nil {
		t.Fatal(err)
	}
	if m["persona"] != "x" || m["policy"] != "y" {
		t.Errorf("got %v", m)
	}

	_, err = parseFacetArgs([]string{"bad"})
	if err == nil {
		t.Error("expected error")
	}
	_, err = parseFacetArgs([]string{"unknown=v"})
	if err == nil {
		t.Error("expected error")
	}
}

func TestApplyFacetArgs_AutoPrefix(t *testing.T) {
	var tmpl Template
	applyFacetArgs(&tmpl, map[string]string{"persona": "engineer"})
	if tmpl.Persona != "@engineer" {
		t.Errorf("expected @engineer, got %q", tmpl.Persona)
	}

	tmpl = Template{}
	applyFacetArgs(&tmpl, map[string]string{"persona": "@engineer"})
	if tmpl.Persona != "@engineer" {
		t.Errorf("expected @engineer, got %q", tmpl.Persona)
	}
}

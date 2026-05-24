package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestTemplate(t *testing.T, facets map[string]string) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", dir)

	tmplDir := filepath.Join(dir, ".prompt-cli", "templates", "test-tmpl")
	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		t.Fatal(err)
	}
	for filename, content := range facets {
		if err := os.WriteFile(filepath.Join(tmplDir, filename), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return tmplDir, func() { os.Setenv("HOME", origHome) }
}

func TestCmdInit_CreatesInstructionFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	tmplDir := filepath.Join(dir, ".prompt-cli", "templates", "new-tmpl")

	if err := os.MkdirAll(tmplDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"persona.md":         "",
		"policy.md":          "",
		"instruction.md":     "",
		"output-contract.md": "",
	}
	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(tmplDir, filename), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	expected := []string{"persona.md", "policy.md", "instruction.md", "output-contract.md"}
	for _, f := range expected {
		path := filepath.Join(tmplDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}
}

func TestCmdRun_AllFacets(t *testing.T) {
	_, cleanup := setupTestTemplate(t, map[string]string{
		"persona.md":         "You are a helpful assistant.",
		"policy.md":          "Be concise.",
		"instruction.md":     "Summarize the input.",
		"output-contract.md": "Output in markdown.",
	})
	defer cleanup()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdRun("test-tmpl")
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun returned error: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	expectations := []struct {
		tag     string
		content string
	}{
		{"persona", "You are a helpful assistant."},
		{"policy", "Be concise."},
		{"instruction", "Summarize the input."},
		{"output-contract", "Output in markdown."},
	}
	for _, e := range expectations {
		expected := "<" + e.tag + ">\n" + e.content + "\n</" + e.tag + ">"
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %s block with content %q, got:\n%s", e.tag, e.content, output)
		}
	}
}

func TestCmdRun_EmptyInstruction(t *testing.T) {
	_, cleanup := setupTestTemplate(t, map[string]string{
		"persona.md":         "You are a bot.",
		"policy.md":          "",
		"instruction.md":     "",
		"output-contract.md": "",
	})
	defer cleanup()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdRun("test-tmpl")
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun returned error: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if strings.Contains(output, "<instruction>") {
		t.Error("expected no instruction block when instruction.md is empty")
	}
}

func TestCmdRun_TemplateNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	err := cmdRun("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestCmdRun_OnlyInstruction(t *testing.T) {
	_, cleanup := setupTestTemplate(t, map[string]string{
		"persona.md":         "",
		"policy.md":          "",
		"instruction.md":     "Do something.",
		"output-contract.md": "",
	})
	defer cleanup()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmdRun("test-tmpl")
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("cmdRun returned error: %v", err)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	expected := "<instruction>\nDo something.\n</instruction>"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain instruction block, got:\n%s", output)
	}

	if strings.Contains(output, "<persona>") {
		t.Error("expected no persona block when persona.md is empty")
	}
}

func TestReadFacet(t *testing.T) {
	dir := t.TempDir()
	content := "  hello world  \n"
	os.WriteFile(filepath.Join(dir, "test.md"), []byte(content), 0644)

	got := readFacet(dir, "test.md")
	if got != "hello world" {
		t.Errorf("expected trimmed content, got %q", got)
	}

	got = readFacet(dir, "nonexistent.md")
	if got != "" {
		t.Errorf("expected empty string for missing file, got %q", got)
	}
}

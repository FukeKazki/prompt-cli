package main

import (
	"os"
	"strings"
	"testing"
)

const releaseWorkflowPath = ".github/workflows/release.yml"

func TestReleaseWorkflow_Exists(t *testing.T) {
	info, err := os.Stat(releaseWorkflowPath)
	if err != nil {
		t.Fatalf("release workflow file does not exist: %v", err)
	}
	if info.IsDir() {
		t.Fatal("expected a file, got a directory")
	}
}

func TestReleaseWorkflow_Structure(t *testing.T) {
	content, err := os.ReadFile(releaseWorkflowPath)
	if err != nil {
		t.Fatalf("failed to read release workflow: %v", err)
	}

	body := string(content)

	required := []struct {
		key   string
		value string
	}{
		{"name", "name: release"},
		{"trigger", "branches: [main]"},
		{"permissions", "contents: write"},
		{"runner", "runs-on: ubuntu-latest"},
		{"checkout", "uses: actions/checkout@v4"},
		{"fetch-depth", "fetch-depth: 0"},
		{"tag-action", "uses: mathieudutour/github-tag-action@v6.2"},
		{"default_bump", "default_bump: patch"},
		{"tag_prefix", "tag_prefix: v"},
	}

	for _, r := range required {
		if !strings.Contains(body, r.value) {
			t.Errorf("missing required %s: expected to contain %q", r.key, r.value)
		}
	}
}

func TestReleaseWorkflow_TriggerOnPush(t *testing.T) {
	content, err := os.ReadFile(releaseWorkflowPath)
	if err != nil {
		t.Fatalf("failed to read release workflow: %v", err)
	}

	body := string(content)

	if !strings.Contains(body, "on:\n  push:") {
		t.Error("workflow should trigger on push event")
	}
}

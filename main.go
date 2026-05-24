package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kazki/prompt-cli/internal/tui"
)

func templateDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".prompt-cli", "templates")
}

func cmdInit(name string) error {
	dir := filepath.Join(templateDir(), name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("template %q already exists", name)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	files := map[string]string{
		"persona.md":         "",
		"policy.md":          "",
		"output-contract.md": "",
	}
	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "Created template: %s\n", dir)
	return cmdEdit(name)
}

func cmdRun(name string, instruction string) error {
	dir := filepath.Join(templateDir(), name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("template %q not found", name)
	}

	if instruction == "" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		instruction = strings.TrimSpace(string(b))
	}
	if instruction == "" {
		return fmt.Errorf("instruction is required (as argument or via stdin)")
	}

	type facet struct {
		file string
		tag  string
	}
	facets := []facet{
		{"persona.md", "persona"},
		{"policy.md", "policy"},
	}

	var parts []string
	for _, f := range facets {
		content, err := os.ReadFile(filepath.Join(dir, f.file))
		if err != nil {
			continue
		}
		if s := strings.TrimSpace(string(content)); s != "" {
			parts = append(parts, fmt.Sprintf("<%s>\n%s\n</%s>", f.tag, s, f.tag))
		}
	}

	parts = append(parts, fmt.Sprintf("<instruction>\n%s\n</instruction>", instruction))

	content, err := os.ReadFile(filepath.Join(dir, "output-contract.md"))
	if err == nil {
		if s := strings.TrimSpace(string(content)); s != "" {
			parts = append(parts, fmt.Sprintf("<output-contract>\n%s\n</output-contract>", s))
		}
	}

	fmt.Println(strings.Join(parts, "\n\n"))
	return nil
}

func cmdList() error {
	dir := templateDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, "No templates found.")
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			fmt.Println(e.Name())
		}
	}
	return nil
}

func readFacet(dir, filename string) string {
	content, err := os.ReadFile(filepath.Join(dir, filename))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

func cmdEdit(name string) error {
	dir := filepath.Join(templateDir(), name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("template %q not found", name)
	}

	data := tui.FormData{
		Persona:        readFacet(dir, "persona.md"),
		Policy:         readFacet(dir, "policy.md"),
		OutputContract: readFacet(dir, "output-contract.md"),
	}

	result, err := tui.RunForm(data)
	if err != nil {
		return err
	}

	files := map[string]string{
		"persona.md":         result.Persona,
		"policy.md":          result.Policy,
		"output-contract.md": result.OutputContract,
	}
	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content+"\n"), 0644); err != nil {
			return err
		}
	}
	fmt.Fprintf(os.Stderr, "Saved template: %s\n", name)
	return nil
}

func cmdDelete(name string) error {
	dir := filepath.Join(templateDir(), name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("template %q not found", name)
	}
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Deleted template: %s\n", name)
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  prompt-cli init <name>              Create a new template
  prompt-cli edit <name>              Edit a template
  prompt-cli run <name> [instruction] Run a template with instruction
  prompt-cli list                     List all templates
  prompt-cli delete <name>            Delete a template`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "init":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli init <name>")
			os.Exit(1)
		}
		err = cmdInit(os.Args[2])
	case "edit":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli edit <name>")
			os.Exit(1)
		}
		err = cmdEdit(os.Args[2])
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli run <name> [instruction]")
			os.Exit(1)
		}
		instruction := ""
		if len(os.Args) >= 4 {
			instruction = strings.Join(os.Args[3:], " ")
		}
		err = cmdRun(os.Args[2], instruction)
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli delete <name>")
			os.Exit(1)
		}
		err = cmdDelete(os.Args[2])
	case "list":
		err = cmdList()
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

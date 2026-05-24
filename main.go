package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kazki/prompt-cli/internal/tui"
	"gopkg.in/yaml.v3"
)

//go:embed builtin/parts
var builtinParts embed.FS

var validFacets = []string{"persona", "policy", "instruction", "output-contract"}

type Template struct {
	Persona        string `yaml:"persona"`
	Policy         string `yaml:"policy"`
	Instruction    string `yaml:"instruction"`
	OutputContract string `yaml:"output-contract"`
}

func baseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".prompt-cli")
}

func templateDir() string {
	return filepath.Join(baseDir(), "templates")
}

func partsDir() string {
	return filepath.Join(baseDir(), "parts")
}

func isValidFacet(facet string) bool {
	for _, f := range validFacets {
		if f == facet {
			return true
		}
	}
	return false
}

func partPath(facet, name string) string {
	return filepath.Join(partsDir(), facet, name+".md")
}

func loadTemplate(name string) (Template, error) {
	path := filepath.Join(templateDir(), name, "template.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Template{}, fmt.Errorf("template %q not found", name)
	}
	var tmpl Template
	if err := yaml.Unmarshal(data, &tmpl); err != nil {
		return Template{}, fmt.Errorf("invalid template %q: %w", name, err)
	}
	return tmpl, nil
}

func saveTemplate(name string, tmpl Template) error {
	dir := filepath.Join(templateDir(), name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := yaml.Marshal(tmpl)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "template.yaml"), data, 0644)
}

func readPart(facet, name string) (string, error) {
	if content, err := os.ReadFile(partPath(facet, name)); err == nil {
		return strings.TrimSpace(string(content)), nil
	}
	if content, err := builtinParts.ReadFile("builtin/parts/" + facet + "/" + name + ".md"); err == nil {
		return strings.TrimSpace(string(content)), nil
	}
	return "", fmt.Errorf("part %s/%s not found", facet, name)
}

func partExists(facet, name string) bool {
	if _, err := os.Stat(partPath(facet, name)); err == nil {
		return true
	}
	if _, err := builtinParts.ReadFile("builtin/parts/" + facet + "/" + name + ".md"); err == nil {
		return true
	}
	return false
}

// @name → part content, otherwise inline text
func resolveFacetContent(value, facet string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(value, "@") {
		return readPart(facet, strings.TrimPrefix(value, "@"))
	}
	return value, nil
}

func validateTemplate(tmpl Template) error {
	check := func(facet, value string) error {
		value = strings.TrimSpace(value)
		if !strings.HasPrefix(value, "@") {
			return nil
		}
		name := strings.TrimPrefix(value, "@")
		if name == "" {
			return nil
		}
		if !partExists(facet, name) {
			return fmt.Errorf("part %s/%s not found", facet, name)
		}
		return nil
	}
	if err := check("persona", tmpl.Persona); err != nil {
		return err
	}
	if err := check("policy", tmpl.Policy); err != nil {
		return err
	}
	if err := check("instruction", tmpl.Instruction); err != nil {
		return err
	}
	return check("output-contract", tmpl.OutputContract)
}

// --- Part commands ---

func cmdPartAdd(facet, name string) error {
	if !isValidFacet(facet) {
		return fmt.Errorf("invalid facet %q (valid: %s)", facet, strings.Join(validFacets, ", "))
	}
	path := partPath(facet, name)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("part %s/%s already exists", facet, name)
	}

	result, err := tui.RunSingleEditor(facet+"/"+name, "")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(partsDir(), facet), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(result+"\n"), 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Created part: %s/%s\n", facet, name)
	return nil
}

func cmdPartEdit(facet, name string) error {
	if !isValidFacet(facet) {
		return fmt.Errorf("invalid facet %q (valid: %s)", facet, strings.Join(validFacets, ", "))
	}
	path := partPath(facet, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("part %s/%s not found", facet, name)
	}

	result, err := tui.RunSingleEditor(facet+"/"+name, strings.TrimSpace(string(content)))
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, []byte(result+"\n"), 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Saved part: %s/%s\n", facet, name)
	return nil
}

func cmdPartList(facet string) error {
	listFacet := func(f string) {
		seen := make(map[string]bool)
		dir := filepath.Join(partsDir(), f)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					name := strings.TrimSuffix(e.Name(), ".md")
					seen[name] = true
					if facet != "" {
						fmt.Println(name)
					} else {
						fmt.Printf("%s/%s\n", f, name)
					}
				}
			}
		}
		if entries, err := builtinParts.ReadDir("builtin/parts/" + f); err == nil {
			for _, e := range entries {
				name := strings.TrimSuffix(e.Name(), ".md")
				if !seen[name] {
					if facet != "" {
						fmt.Printf("%s (builtin)\n", name)
					} else {
						fmt.Printf("%s/%s (builtin)\n", f, name)
					}
				}
			}
		}
	}

	if facet != "" {
		if !isValidFacet(facet) {
			return fmt.Errorf("invalid facet %q (valid: %s)", facet, strings.Join(validFacets, ", "))
		}
		listFacet(facet)
		return nil
	}
	for _, f := range validFacets {
		listFacet(f)
	}
	return nil
}

func cmdPartDelete(facet, name string) error {
	if !isValidFacet(facet) {
		return fmt.Errorf("invalid facet %q (valid: %s)", facet, strings.Join(validFacets, ", "))
	}
	path := partPath(facet, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("part %s/%s not found", facet, name)
	}
	if err := os.Remove(path); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Deleted part: %s/%s\n", facet, name)
	return nil
}

// --- Template commands ---

func collectAvailableParts() tui.AvailableParts {
	collect := func(facet string) []tui.PartInfo {
		var parts []tui.PartInfo
		seen := make(map[string]bool)
		dir := filepath.Join(partsDir(), facet)
		if entries, err := os.ReadDir(dir); err == nil {
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					name := strings.TrimSuffix(e.Name(), ".md")
					seen[name] = true
					parts = append(parts, tui.PartInfo{Name: name})
				}
			}
		}
		if entries, err := builtinParts.ReadDir("builtin/parts/" + facet); err == nil {
			for _, e := range entries {
				name := strings.TrimSuffix(e.Name(), ".md")
				if !seen[name] {
					parts = append(parts, tui.PartInfo{Name: name, Builtin: true})
				}
			}
		}
		return parts
	}
	return tui.AvailableParts{
		Persona:        collect("persona"),
		Policy:         collect("policy"),
		Instruction:    collect("instruction"),
		OutputContract: collect("output-contract"),
	}
}

func templateToFormData(tmpl Template) tui.TemplateFormData {
	return tui.TemplateFormData{
		Persona:        tmpl.Persona,
		Policy:         tmpl.Policy,
		Instruction:    tmpl.Instruction,
		OutputContract: tmpl.OutputContract,
	}
}

func formDataToTemplate(data tui.TemplateFormData) Template {
	return Template{
		Persona:        data.Persona,
		Policy:         data.Policy,
		Instruction:    data.Instruction,
		OutputContract: data.OutputContract,
	}
}

func cmdInit(name string) error {
	dir := filepath.Join(templateDir(), name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("template %q already exists", name)
	}

	result, err := tui.RunTemplateForm(tui.TemplateFormData{}, collectAvailableParts())
	if err != nil {
		return err
	}
	tmpl := formDataToTemplate(result)

	if err := validateTemplate(tmpl); err != nil {
		return err
	}
	if err := saveTemplate(name, tmpl); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Created template: %s\n", name)
	return nil
}

func cmdEdit(name string) error {
	tmpl, err := loadTemplate(name)
	if err != nil {
		return err
	}

	result, err := tui.RunTemplateForm(templateToFormData(tmpl), collectAvailableParts())
	if err != nil {
		return err
	}
	tmpl = formDataToTemplate(result)

	if err := validateTemplate(tmpl); err != nil {
		return err
	}
	if err := saveTemplate(name, tmpl); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Saved template: %s\n", name)
	return nil
}

func cmdRun(name string) error {
	tmpl, err := loadTemplate(name)
	if err != nil {
		return err
	}

	type facetDef struct {
		value string
		facet string
		tag   string
	}
	defs := []facetDef{
		{tmpl.Persona, "persona", "persona"},
		{tmpl.Policy, "policy", "policy"},
		{tmpl.Instruction, "instruction", "instruction"},
		{tmpl.OutputContract, "output-contract", "output-contract"},
	}

	var sections []string
	for _, d := range defs {
		content, err := resolveFacetContent(d.value, d.facet)
		if err != nil {
			return err
		}
		if content != "" {
			sections = append(sections, fmt.Sprintf("<%s>\n%s\n</%s>", d.tag, content, d.tag))
		}
	}

	fmt.Println(strings.Join(sections, "\n\n"))
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
  prompt-cli run <name>               Run a template
  prompt-cli list                     List all templates
  prompt-cli delete <name>            Delete a template

  prompt-cli part add <facet> <name>        Create a reusable part
  prompt-cli part edit <facet> <name>       Edit a part
  prompt-cli part list [facet]              List parts
  prompt-cli part delete <facet> <name>     Delete a part

Facets: persona, policy, instruction, output-contract`)
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
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli run <name>")
			os.Exit(1)
		}
		err = cmdRun(os.Args[2])
	case "delete":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: prompt-cli delete <name>")
			os.Exit(1)
		}
		err = cmdDelete(os.Args[2])
	case "list":
		err = cmdList()
	case "part":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "add":
			if len(os.Args) < 5 {
				fmt.Fprintln(os.Stderr, "Usage: prompt-cli part add <facet> <name>")
				os.Exit(1)
			}
			err = cmdPartAdd(os.Args[3], os.Args[4])
		case "edit":
			if len(os.Args) < 5 {
				fmt.Fprintln(os.Stderr, "Usage: prompt-cli part edit <facet> <name>")
				os.Exit(1)
			}
			err = cmdPartEdit(os.Args[3], os.Args[4])
		case "list":
			facet := ""
			if len(os.Args) >= 4 {
				facet = os.Args[3]
			}
			err = cmdPartList(facet)
		case "delete":
			if len(os.Args) < 5 {
				fmt.Fprintln(os.Stderr, "Usage: prompt-cli part delete <facet> <name>")
				os.Exit(1)
			}
			err = cmdPartDelete(os.Args[3], os.Args[4])
		default:
			usage()
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

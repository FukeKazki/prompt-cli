package tui

import (
	"testing"
)

func TestNewFormModel_FieldCount(t *testing.T) {
	m := newFormModel(FormData{Persona: "p", Policy: "po", Instruction: "i", OutputContract: "oc"})
	if len(m.areas) != focusCount {
		t.Errorf("expected %d areas, got %d", focusCount, len(m.areas))
	}
}

func TestFormModel_Data_RoundTrip(t *testing.T) {
	data := FormData{Persona: "a", Policy: "b", Instruction: "c", OutputContract: "d"}
	got := newFormModel(data).Data()
	if got != data {
		t.Errorf("expected %+v, got %+v", data, got)
	}
}

func TestFormModel_Data_TrimsWhitespace(t *testing.T) {
	data := FormData{Persona: "  x  ", Policy: "  y  ", Instruction: "  z  ", OutputContract: "  w  "}
	got := newFormModel(data).Data()
	if got.Persona != "x" || got.Policy != "y" || got.Instruction != "z" || got.OutputContract != "w" {
		t.Errorf("not trimmed: %+v", got)
	}
}

func TestFocusOrder(t *testing.T) {
	if focusPersona != 0 || focusPolicy != 1 || focusInstruction != 2 || focusOutputContract != 3 {
		t.Error("unexpected focus order")
	}
}

func TestTemplateFormModel_RoundTrip(t *testing.T) {
	data := TemplateFormData{
		Persona:  "@engineer",
		Policy:   "Be concise.",
		Instruction: "@code-review",
	}
	m := newTemplateFormModel(data, AvailableParts{})
	got := m.Data()

	if got.Persona != "@engineer" {
		t.Errorf("Persona: %q", got.Persona)
	}
	if got.Policy != "Be concise." {
		t.Errorf("Policy: %q", got.Policy)
	}
	if got.Instruction != "@code-review" {
		t.Errorf("Instruction: %q", got.Instruction)
	}
	if got.OutputContract != "" {
		t.Errorf("OutputContract: %q", got.OutputContract)
	}
}

func TestTemplateFormModel_Completing(t *testing.T) {
	available := AvailableParts{
		Persona: []PartInfo{
			{Name: "code-reviewer", Builtin: true},
			{Name: "software-engineer", Builtin: true},
		},
	}
	data := TemplateFormData{Persona: "@co"}
	m := newTemplateFormModel(data, available)

	if !m.isCompleting() {
		t.Error("expected completing for @co")
	}

	filtered := m.filteredParts()
	if len(filtered) != 1 || filtered[0].Name != "code-reviewer" {
		t.Errorf("expected [code-reviewer], got %v", filtered)
	}
}

func TestTemplateFormModel_NotCompleting(t *testing.T) {
	data := TemplateFormData{Persona: "plain text"}
	m := newTemplateFormModel(data, AvailableParts{})

	if m.isCompleting() {
		t.Error("should not be completing for plain text")
	}
}

func TestTemplateFormModel_NotCompleting_Multiline(t *testing.T) {
	data := TemplateFormData{Persona: "@code\nmore text"}
	m := newTemplateFormModel(data, AvailableParts{})

	if m.isCompleting() {
		t.Error("should not be completing for multiline starting with @")
	}
}

func TestTemplateFormModel_FilterAll(t *testing.T) {
	available := AvailableParts{
		Persona: []PartInfo{
			{Name: "a", Builtin: false},
			{Name: "b", Builtin: true},
		},
	}
	data := TemplateFormData{Persona: "@"}
	m := newTemplateFormModel(data, available)

	filtered := m.filteredParts()
	if len(filtered) != 2 {
		t.Errorf("expected 2, got %d", len(filtered))
	}
}

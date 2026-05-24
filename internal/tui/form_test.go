package tui

import (
	"testing"
)

func TestNewFormModel_FieldCount(t *testing.T) {
	data := FormData{
		Persona:        "p",
		Policy:         "po",
		Instruction:    "i",
		OutputContract: "oc",
	}
	m := newFormModel(data)

	if len(m.areas) != focusCount {
		t.Errorf("expected %d areas, got %d", focusCount, len(m.areas))
	}
}

func TestFormModel_Data_RoundTrip(t *testing.T) {
	data := FormData{
		Persona:        "persona text",
		Policy:         "policy text",
		Instruction:    "instruction text",
		OutputContract: "output contract text",
	}
	m := newFormModel(data)
	got := m.Data()

	if got.Persona != data.Persona {
		t.Errorf("Persona: expected %q, got %q", data.Persona, got.Persona)
	}
	if got.Policy != data.Policy {
		t.Errorf("Policy: expected %q, got %q", data.Policy, got.Policy)
	}
	if got.Instruction != data.Instruction {
		t.Errorf("Instruction: expected %q, got %q", data.Instruction, got.Instruction)
	}
	if got.OutputContract != data.OutputContract {
		t.Errorf("OutputContract: expected %q, got %q", data.OutputContract, got.OutputContract)
	}
}

func TestFormModel_Data_TrimsWhitespace(t *testing.T) {
	data := FormData{
		Persona:        "  persona  ",
		Policy:         "  policy  ",
		Instruction:    "  instruction  ",
		OutputContract: "  output  ",
	}
	m := newFormModel(data)
	got := m.Data()

	if got.Persona != "persona" {
		t.Errorf("Persona not trimmed: got %q", got.Persona)
	}
	if got.Policy != "policy" {
		t.Errorf("Policy not trimmed: got %q", got.Policy)
	}
	if got.Instruction != "instruction" {
		t.Errorf("Instruction not trimmed: got %q", got.Instruction)
	}
	if got.OutputContract != "output" {
		t.Errorf("OutputContract not trimmed: got %q", got.OutputContract)
	}
}

func TestFieldLabels_IncludesInstruction(t *testing.T) {
	if focusCount != 4 {
		t.Fatalf("expected focusCount=4, got %d", focusCount)
	}
	if fieldLabels[focusInstruction] != "Instruction" {
		t.Errorf("expected label %q at focusInstruction, got %q", "Instruction", fieldLabels[focusInstruction])
	}
}

func TestFocusOrder(t *testing.T) {
	if focusPersona != 0 {
		t.Errorf("focusPersona should be 0, got %d", focusPersona)
	}
	if focusPolicy != 1 {
		t.Errorf("focusPolicy should be 1, got %d", focusPolicy)
	}
	if focusInstruction != 2 {
		t.Errorf("focusInstruction should be 2, got %d", focusInstruction)
	}
	if focusOutputContract != 3 {
		t.Errorf("focusOutputContract should be 3, got %d", focusOutputContract)
	}
}

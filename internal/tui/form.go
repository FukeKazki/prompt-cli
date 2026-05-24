package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	colAccent = lipgloss.Color("13")
	colMuted  = lipgloss.Color("240")

	labelFocus  = lipgloss.NewStyle().Bold(true).Foreground(colAccent)
	labelBlur   = lipgloss.NewStyle().Bold(true).Foreground(colMuted)
	footerStyle = lipgloss.NewStyle().Faint(true)
)

const (
	focusPersona = iota
	focusPolicy
	focusInstruction
	focusOutputContract
	focusCount
)

var fieldLabels = [focusCount]string{
	"Persona",
	"Policy",
	"Instruction",
	"Output Contract",
}

type FormData struct {
	Persona        string
	Policy         string
	Instruction    string
	OutputContract string
}

type formModel struct {
	areas     [focusCount]textarea.Model
	focus     int
	submitted bool
	canceled  bool
	width     int
	height    int
}

func newFormModel(data FormData) formModel {
	makeArea := func(value string) textarea.Model {
		ta := textarea.New()
		ta.ShowLineNumbers = false
		ta.SetValue(value)
		ta.CharLimit = 0
		return ta
	}

	m := formModel{
		areas: [focusCount]textarea.Model{
			makeArea(data.Persona),
			makeArea(data.Policy),
			makeArea(data.Instruction),
			makeArea(data.OutputContract),
		},
	}
	m.areas[m.focus].Focus()
	return m
}

func (m formModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeAreas()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.submitted = true
			return m, tea.Quit
		case "esc":
			m.canceled = true
			return m, tea.Quit
		case "tab":
			m.areas[m.focus].Blur()
			m.focus = (m.focus + 1) % focusCount
			m.areas[m.focus].Focus()
			return m, nil
		case "shift+tab":
			m.areas[m.focus].Blur()
			m.focus = (m.focus - 1 + focusCount) % focusCount
			m.areas[m.focus].Focus()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.areas[m.focus], cmd = m.areas[m.focus].Update(msg)
	return m, cmd
}

func (m formModel) View() string {
	if m.width == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("\n")

	for i := 0; i < focusCount; i++ {
		label := fieldLabels[i]
		if i == m.focus {
			b.WriteString(labelFocus.Render("▸ " + label))
		} else {
			b.WriteString(labelBlur.Render("  " + label))
		}
		b.WriteString("\n")
		b.WriteString(m.areas[i].View())
		b.WriteString("\n\n")
	}

	b.WriteString(footerStyle.Render("tab/shift+tab: switch • ctrl+s: save • esc: cancel"))
	return b.String()
}

func (m formModel) resizeAreas() formModel {
	w := m.width - 4
	if w < 20 {
		w = 20
	}

	areaHeight := (m.height - focusCount*2 - 4) / focusCount
	if areaHeight < 3 {
		areaHeight = 3
	}

	for i := range m.areas {
		m.areas[i].SetWidth(w)
		m.areas[i].SetHeight(areaHeight)
	}
	return m
}

func (m formModel) Data() FormData {
	return FormData{
		Persona:        strings.TrimSpace(m.areas[focusPersona].Value()),
		Policy:         strings.TrimSpace(m.areas[focusPolicy].Value()),
		Instruction:    strings.TrimSpace(m.areas[focusInstruction].Value()),
		OutputContract: strings.TrimSpace(m.areas[focusOutputContract].Value()),
	}
}

func RunForm(data FormData) (FormData, error) {
	m := newFormModel(data)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return data, fmt.Errorf("running form: %w", err)
	}

	fm := result.(formModel)
	if fm.canceled {
		return data, fmt.Errorf("canceled")
	}
	return fm.Data(), nil
}

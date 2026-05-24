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

type singleEditorModel struct {
	area      textarea.Model
	label     string
	submitted bool
	canceled  bool
	width     int
	height    int
}

func newSingleEditorModel(label, content string) singleEditorModel {
	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.SetValue(content)
	ta.CharLimit = 0
	ta.Focus()
	return singleEditorModel{area: ta, label: label}
}

func (m singleEditorModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m singleEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		w := m.width - 4
		if w < 20 {
			w = 20
		}
		h := m.height - 6
		if h < 3 {
			h = 3
		}
		m.area.SetWidth(w)
		m.area.SetHeight(h)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.submitted = true
			return m, tea.Quit
		case "esc":
			m.canceled = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	return m, cmd
}

func (m singleEditorModel) View() string {
	if m.width == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(labelFocus.Render("▸ " + m.label))
	b.WriteString("\n")
	b.WriteString(m.area.View())
	b.WriteString("\n\n")
	b.WriteString(footerStyle.Render("ctrl+s: save • esc: cancel"))
	return b.String()
}

func RunSingleEditor(label, content string) (string, error) {
	m := newSingleEditorModel(label, content)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return content, fmt.Errorf("running editor: %w", err)
	}
	fm := result.(singleEditorModel)
	if fm.canceled {
		return content, fmt.Errorf("canceled")
	}
	return strings.TrimSpace(fm.area.Value()), nil
}

// --- Template form ---

type PartInfo struct {
	Name    string
	Builtin bool
}

type AvailableParts struct {
	Persona        []PartInfo
	Policy         []PartInfo
	Instruction    []PartInfo
	OutputContract []PartInfo
}

type TemplateFormData struct {
	Persona        string
	Policy         string
	Instruction    string
	OutputContract string
}

type templateFormModel struct {
	areas       [focusCount]textarea.Model
	focus       int
	submitted   bool
	canceled    bool
	width       int
	height      int
	available   [focusCount][]PartInfo
	completeIdx int
}

func newTemplateFormModel(data TemplateFormData, available AvailableParts) templateFormModel {
	values := [focusCount]string{
		data.Persona,
		data.Policy,
		data.Instruction,
		data.OutputContract,
	}

	var areas [focusCount]textarea.Model
	for i, v := range values {
		ta := textarea.New()
		ta.ShowLineNumbers = false
		ta.SetValue(v)
		ta.CharLimit = 0
		ta.Placeholder = "text or @part"
		areas[i] = ta
	}

	m := templateFormModel{
		areas: areas,
		available: [focusCount][]PartInfo{
			available.Persona,
			available.Policy,
			available.Instruction,
			available.OutputContract,
		},
	}
	m.areas[0].Focus()
	return m
}

func (m templateFormModel) isCompleting() bool {
	value := m.areas[m.focus].Value()
	return strings.HasPrefix(value, "@") && !strings.Contains(value, "\n")
}

func (m templateFormModel) filteredParts() []PartInfo {
	filter := strings.ToLower(strings.TrimPrefix(m.areas[m.focus].Value(), "@"))
	parts := m.available[m.focus]
	if filter == "" {
		return parts
	}
	var result []PartInfo
	for _, p := range parts {
		if strings.Contains(strings.ToLower(p.Name), filter) {
			result = append(result, p)
		}
	}
	return result
}

func (m templateFormModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m templateFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resize()
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+s":
			m.submitted = true
			return m, tea.Quit
		case "esc":
			if m.isCompleting() {
				m.areas[m.focus].SetValue("")
				return m, nil
			}
			m.canceled = true
			return m, tea.Quit
		case "tab":
			m.areas[m.focus].Blur()
			m.focus = (m.focus + 1) % focusCount
			m.areas[m.focus].Focus()
			m.completeIdx = 0
			return m, nil
		case "shift+tab":
			m.areas[m.focus].Blur()
			m.focus = (m.focus - 1 + focusCount) % focusCount
			m.areas[m.focus].Focus()
			m.completeIdx = 0
			return m, nil
		case "up":
			if m.isCompleting() {
				if m.completeIdx > 0 {
					m.completeIdx--
				}
				return m, nil
			}
		case "down":
			if m.isCompleting() {
				filtered := m.filteredParts()
				if m.completeIdx < len(filtered)-1 {
					m.completeIdx++
				}
				return m, nil
			}
		case "enter":
			if m.isCompleting() {
				filtered := m.filteredParts()
				if len(filtered) > 0 && m.completeIdx < len(filtered) {
					m.areas[m.focus].SetValue("@" + filtered[m.completeIdx].Name)
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	m.areas[m.focus], cmd = m.areas[m.focus].Update(msg)
	if m.isCompleting() {
		m.completeIdx = 0
	}
	return m, cmd
}

func (m templateFormModel) resize() templateFormModel {
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

func (m templateFormModel) View() string {
	if m.width == 0 {
		return ""
	}

	completing := m.isCompleting()

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
		b.WriteString("\n")

		if completing && i == m.focus {
			filtered := m.filteredParts()
			if len(filtered) == 0 {
				b.WriteString(footerStyle.Render("  (no matches)"))
				b.WriteString("\n")
			} else {
				for j, p := range filtered {
					name := p.Name
					if p.Builtin {
						name += "*"
					}
					if j == m.completeIdx {
						b.WriteString(labelFocus.Render("  ▸ " + name))
					} else {
						b.WriteString(footerStyle.Render("    " + name))
					}
					b.WriteString("\n")
				}
			}
		}
		b.WriteString("\n")
	}

	if completing {
		b.WriteString(footerStyle.Render("↑↓: select • enter: confirm • esc: clear"))
	} else {
		b.WriteString(footerStyle.Render("@...: part ref • tab/shift+tab: switch • ctrl+s: save • esc: cancel"))
	}
	return b.String()
}

func (m templateFormModel) Data() TemplateFormData {
	return TemplateFormData{
		Persona:        strings.TrimSpace(m.areas[focusPersona].Value()),
		Policy:         strings.TrimSpace(m.areas[focusPolicy].Value()),
		Instruction:    strings.TrimSpace(m.areas[focusInstruction].Value()),
		OutputContract: strings.TrimSpace(m.areas[focusOutputContract].Value()),
	}
}

func RunTemplateForm(data TemplateFormData, available AvailableParts) (TemplateFormData, error) {
	m := newTemplateFormModel(data, available)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return data, fmt.Errorf("running form: %w", err)
	}
	fm := result.(templateFormModel)
	if fm.canceled {
		return data, fmt.Errorf("canceled")
	}
	return fm.Data(), nil
}

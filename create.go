package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	fieldTitle = iota
	fieldDesc
	fieldTeam
	fieldCount
)

type createForm struct {
	teams      []TeamInfo
	teamIdx    int
	focusField int
	title      string
	desc       string
	cursorPos  int
	submitting bool
	err        error
}

func newCreateForm(teams []TeamInfo) *createForm {
	idx := 0
	if saved := loadLastTeam(); saved != "" {
		for i, t := range teams {
			if t.ID == saved {
				idx = i
				break
			}
		}
	}
	return &createForm{teams: teams, teamIdx: idx}
}

type createResultMsg struct {
	result *CreateIssueResult
	err    error
}

type submitCreateMsg struct {
	title   string
	desc    string
	teamID  string
	stateID string
}

func (f *createForm) Update(msg tea.Msg) (*createForm, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if f.submitting {
			return f, nil
		}

		switch msg.String() {
		case "esc":
			return nil, nil
		case "tab":
			f.focusField = (f.focusField + 1) % fieldCount
			f.cursorPos = f.activeLen()
			return f, nil
		case "shift+tab":
			f.focusField = (f.focusField - 1 + fieldCount) % fieldCount
			f.cursorPos = f.activeLen()
			return f, nil
		case "enter":
			return f, f.submit()
		}

		if f.focusField == fieldTeam {
			switch msg.String() {
			case "left", "h":
				if f.teamIdx > 0 {
					f.teamIdx--
				} else {
					f.teamIdx = len(f.teams) - 1
				}
			case "right", "l":
				if f.teamIdx < len(f.teams)-1 {
					f.teamIdx++
				} else {
					f.teamIdx = 0
				}
			}
			return f, nil
		}

		switch msg.String() {
		case "backspace":
			if f.cursorPos > 0 {
				r := []rune(f.activeGet())
				f.activeSet(string(r[:f.cursorPos-1]) + string(r[f.cursorPos:]))
				f.cursorPos--
			}
		case "left":
			if f.cursorPos > 0 {
				f.cursorPos--
			}
		case "right":
			if f.cursorPos < f.activeLen() {
				f.cursorPos++
			}
		case "ctrl+a", "home":
			f.cursorPos = 0
		case "ctrl+e", "end":
			f.cursorPos = f.activeLen()
		case "ctrl+k":
			r := []rune(f.activeGet())
			f.activeSet(string(r[:f.cursorPos]))
		case "ctrl+u":
			r := []rune(f.activeGet())
			f.activeSet(string(r[f.cursorPos:]))
			f.cursorPos = 0
		default:
			if msg.Text != "" && msg.Text != " " || msg.String() == "space" {
				ch := msg.Text
				if msg.String() == "space" {
					ch = " "
				}
				r := []rune(f.activeGet())
				f.activeSet(string(r[:f.cursorPos]) + ch + string(r[f.cursorPos:]))
				f.cursorPos += len([]rune(ch))
			}
		}
	}
	return f, nil
}

func (f *createForm) activeLen() int {
	return len([]rune(f.activeGet()))
}

func (f *createForm) activeGet() string {
	if f.focusField == fieldDesc {
		return f.desc
	}
	return f.title
}

func (f *createForm) activeSet(s string) {
	if f.focusField == fieldDesc {
		f.desc = s
	} else {
		f.title = s
	}
}

func (f *createForm) submit() tea.Cmd {
	title := strings.TrimSpace(f.title)
	if title == "" {
		f.err = fmt.Errorf("title is required")
		return nil
	}
	if len(f.teams) == 0 {
		f.err = fmt.Errorf("no teams available")
		return nil
	}
	f.submitting = true
	f.err = nil

	team := f.teams[f.teamIdx]
	saveLastTeam(team.ID)
	stateID := team.DefaultStateID()
	desc := strings.TrimSpace(f.desc)

	return func() tea.Msg {
		return submitCreateMsg{
			title:   title,
			desc:    desc,
			teamID:  team.ID,
			stateID: stateID,
		}
	}
}

// --- View ---

func (f *createForm) View(width, height int) string {
	formW := width * 3 / 4
	if formW > 80 {
		formW = 80
	}
	if formW < 30 {
		formW = 30
	}
	inputW := formW - 6
	b := border

	var lines []string

	// Top border
	titleText := formTitle.Render(" Create Issue ")
	titleVisW := lipgloss.Width(titleText)
	topFill := formW - titleVisW - 4
	if topFill < 0 {
		topFill = 0
	}
	lines = append(lines, b.Render("╭─")+titleText+b.Render(strings.Repeat("─", topFill)+"─╮"))

	row := func(content string) {
		lines = append(lines, b.Render("│")+fitWidth("  "+content, formW-2)+b.Render("│"))
	}
	emptyRow := func() { row("") }

	// Title
	emptyRow()
	if f.focusField == fieldTitle {
		row(formLabelFocused.Render("Title"))
	} else {
		row(formLabel.Render("Title"))
	}
	f.addInputRows(&lines, f.title, inputW, formW, f.focusField == fieldTitle, fieldTitle)

	// Description
	emptyRow()
	if f.focusField == fieldDesc {
		row(formLabelFocused.Render("Description"))
	} else {
		row(formLabel.Render("Description"))
	}
	f.addInputRows(&lines, f.desc, inputW, formW, f.focusField == fieldDesc, fieldDesc)

	// Team
	emptyRow()
	if f.focusField == fieldTeam {
		row(formLabelFocused.Render("Team"))
	} else {
		row(formLabel.Render("Team"))
	}
	teamName := "No teams"
	if len(f.teams) > 0 {
		teamName = f.teams[f.teamIdx].Name
	}
	leftArr := formArrow.Render("◂ ")
	rightArr := formArrow.Render(" ▸")
	if f.focusField == fieldTeam {
		row(leftArr + formTeamFocused.Render(teamName) + rightArr)
	} else {
		row(leftArr + formTeam.Render(teamName) + rightArr)
	}

	// Error / status
	if f.err != nil {
		emptyRow()
		row(formError.Render(f.err.Error()))
	} else if f.submitting {
		emptyRow()
		row(formLabel.Render("Creating…"))
	}

	emptyRow()

	// Separator
	lines = append(lines, b.Render("├"+strings.Repeat("─", formW-2)+"┤"))

	// Help bar
	help := formHelpKey.Render("tab") + " " + formHelpDesc.Render("next") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("enter") + " " + formHelpDesc.Render("submit") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("esc") + " " + formHelpDesc.Render("cancel")
	helpPad := formW - lipgloss.Width(help) - 4
	if helpPad < 0 {
		helpPad = 0
	}
	lines = append(lines, b.Render("│")+" "+help+strings.Repeat(" ", helpPad)+" "+b.Render("│"))

	// Bottom border
	lines = append(lines, b.Render("╰"+strings.Repeat("─", formW-2)+"╯"))

	// Center the form
	formContent := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, formContent)
}

func (f *createForm) addInputRows(lines *[]string, text string, inputW, formW int, focused bool, field int) {
	b := border
	style := formInputBorder
	if focused {
		style = formInputBorderFocused
	}

	// Top border of input
	*lines = append(*lines, b.Render("│")+"  "+style.Render("┌"+strings.Repeat("─", inputW)+"┐")+
		strings.Repeat(" ", max(0, formW-inputW-6))+b.Render("│"))

	content := text
	if focused {
		r := []rune(text)
		if f.cursorPos < len(r) {
			content = string(r[:f.cursorPos]) + formCursor.Render(string(r[f.cursorPos:f.cursorPos+1])) + string(r[f.cursorPos+1:])
		} else {
			content = text + formCursor.Render(" ")
		}
	}

	inputLine := style.Render("│") + fitWidth(content, inputW) + style.Render("│")
	pad := max(0, formW-inputW-6)
	*lines = append(*lines, b.Render("│")+"  "+inputLine+strings.Repeat(" ", pad)+b.Render("│"))

	// For description, add extra empty lines
	if field == fieldDesc {
		for i := 0; i < 2; i++ {
			emptyInput := style.Render("│") + strings.Repeat(" ", inputW) + style.Render("│")
			*lines = append(*lines, b.Render("│")+"  "+emptyInput+strings.Repeat(" ", pad)+b.Render("│"))
		}
	}

	// Bottom border of input
	*lines = append(*lines, b.Render("│")+"  "+style.Render("└"+strings.Repeat("─", inputW)+"┘")+
		strings.Repeat(" ", pad)+b.Render("│"))
}

// --- Team persistence ---

func stateDir() string {
	dir := os.Getenv("XDG_STATE_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(dir, "omarchy-linear-notifications")
}

func saveLastTeam(teamID string) {
	dir := stateDir()
	_ = os.MkdirAll(dir, 0o755)
	_ = os.WriteFile(filepath.Join(dir, "last-team"), []byte(teamID), 0o644)
}

func loadLastTeam() string {
	data, err := os.ReadFile(filepath.Join(stateDir(), "last-team"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// --- Post-create modal ---

func renderCreatedModal(issue *CreateIssueResult, width, height int) string {
	modalW := 64
	if modalW > width-4 {
		modalW = width - 4
	}
	b := border
	innerW := modalW - 2

	var lines []string

	titleText := formTitle.Render(" Issue Created ")
	titleVisW := lipgloss.Width(titleText)
	topFill := max(0, modalW-titleVisW-4)
	lines = append(lines, b.Render("╭─")+titleText+b.Render(strings.Repeat("─", topFill)+"─╮"))

	row := func(content string) {
		lines = append(lines, b.Render("│")+fitWidth("  "+content, innerW)+b.Render("│"))
	}

	row("")
	row(detailIdent.Render(issue.Identifier))
	row("")
	if issue.BranchName != "" {
		row(formLabel.Render("Branch  ") + detailValue.Render(truncate(issue.BranchName, innerW-12)))
	}
	row(formLabel.Render("URL     ") + detailValue.Render(truncate(issue.URL, innerW-12)))
	row("")

	lines = append(lines, b.Render("├"+strings.Repeat("─", innerW)+"┤"))

	help := formHelpKey.Render("o") + " " + formHelpDesc.Render("open") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("u") + " " + formHelpDesc.Render("copy url") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("y") + " " + formHelpDesc.Render("copy id") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("b") + " " + formHelpDesc.Render("copy branch") +
		formHelpSep.Render(" │ ") +
		formHelpKey.Render("esc") + " " + formHelpDesc.Render("close")
	helpPad := max(0, innerW-lipgloss.Width(help)-2)
	lines = append(lines, b.Render("│")+" "+help+strings.Repeat(" ", helpPad)+" "+b.Render("│"))
	lines = append(lines, b.Render("╰"+strings.Repeat("─", innerW)+"╯"))

	content := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

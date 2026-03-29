package main

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type model struct {
	client        *Client
	notifications []Notification
	unreadCount   int
	cursor        int
	width         int
	height        int
	loading       bool
	err           error
	openURL       string
	createForm    *createForm
	teamsCache    []TeamInfo
	viewerID      string
	flash         string
	createdIssue  *CreateIssueResult
}

type notificationsMsg struct {
	notifications []Notification
	unreadCount   int
	err           error
}

type markReadMsg struct {
	id  string
	err error
}

type markAllReadMsg struct {
	err error
}

type teamsMsg struct {
	teams    []TeamInfo
	viewerID string
	err      error
}

func newModel(client *Client) model {
	return model{client: client, loading: true}
}

func (m model) Init() tea.Cmd {
	if m.client == nil {
		return nil
	}
	return m.fetchNotifications()
}

func (m model) fetchNotifications() tea.Cmd {
	return func() tea.Msg {
		notifs, _, err := m.client.FetchNotifications("")
		if err != nil {
			return notificationsMsg{err: err}
		}
		count, _ := m.client.FetchUnreadCount()
		return notificationsMsg{notifications: notifs, unreadCount: count}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle post-create modal
	if m.createdIssue != nil {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			return m, nil
		case tea.KeyPressMsg:
			switch msg.String() {
			case "o":
				url := m.createdIssue.URL
				m.flash = "Created " + m.createdIssue.Identifier
				m.createdIssue = nil
				m.openURL = url
				return m, tea.Quit
			case "u":
				_ = copyToClipboard(m.createdIssue.URL)
				m.flash = "Copied URL"
				m.createdIssue = nil
			case "y":
				_ = copyToClipboard(m.createdIssue.Identifier)
				m.flash = "Copied " + m.createdIssue.Identifier
				m.createdIssue = nil
			case "b":
				_ = copyToClipboard(m.createdIssue.BranchName)
				m.flash = "Copied branch name"
				m.createdIssue = nil
			case "enter", "esc", "q":
				m.flash = "Created " + m.createdIssue.Identifier
				m.createdIssue = nil
			}
		}
		return m, nil
	}

	// Handle create form overlay
	if m.createForm != nil {
		switch msg := msg.(type) {
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			return m, nil
		case submitCreateMsg:
			client := m.client
			viewerID := m.viewerID
			return m, func() tea.Msg {
				result, err := client.CreateIssue(msg.title, msg.desc, msg.teamID, msg.stateID, viewerID)
				return createResultMsg{result: result, err: err}
			}
		case createResultMsg:
			if msg.err != nil {
				m.createForm.submitting = false
				m.createForm.err = msg.err
				return m, nil
			}
			m.createForm = nil
			m.createdIssue = msg.result
			return m, nil
		default:
			updated, cmd := m.createForm.Update(msg)
			m.createForm = updated
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case notificationsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.notifications = msg.notifications
		m.unreadCount = msg.unreadCount

	case markReadMsg:
		// State already updated optimistically on keypress

	case markAllReadMsg:
		// State already updated optimistically on keypress

	case teamsMsg:
		if msg.err != nil {
			m.flash = "Failed to load teams"
			return m, nil
		}
		m.teamsCache = msg.teams
		m.viewerID = msg.viewerID
		m.createForm = newCreateForm(m.teamsCache)
		return m, nil

	case tea.KeyPressMsg:
		m.flash = ""
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.notifications)-1 {
				m.cursor++
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
			}
		case "o", "enter":
			if m.client != nil && len(m.notifications) > 0 {
				m.openURL = notificationURL(m.notifications[m.cursor])
				return m, tea.Quit
			}
		case "r":
			if m.client != nil && len(m.notifications) > 0 && m.notifications[m.cursor].ReadAt == nil {
				id := m.notifications[m.cursor].ID
				now := time.Now().UTC().Format(time.RFC3339)
				m.notifications[m.cursor].ReadAt = &now
				m.unreadCount = max(0, m.unreadCount-1)
				client := m.client
				return m, func() tea.Msg {
					_ = client.MarkAsRead(id)
					return markReadMsg{id: id}
				}
			}
		case "R":
			if m.client == nil {
				break
			}
			now := time.Now().UTC().Format(time.RFC3339)
			ids := make([]string, 0)
			for i := range m.notifications {
				if m.notifications[i].ReadAt == nil {
					ids = append(ids, m.notifications[i].ID)
					m.notifications[i].ReadAt = &now
				}
			}
			m.unreadCount = 0
			if len(ids) > 0 {
				client := m.client
				return m, func() tea.Msg {
					for _, id := range ids {
						_ = client.MarkAsRead(id)
					}
					return markAllReadMsg{}
				}
			}
		case "c":
			if m.client == nil {
				break
			}
			if m.teamsCache != nil {
				m.createForm = newCreateForm(m.teamsCache)
				return m, nil
			}
			client := m.client
			return m, func() tea.Msg {
				teams, err := client.FetchTeams()
				if err != nil {
					return teamsMsg{err: err}
				}
				viewerID, _ := client.FetchViewerID()
				return teamsMsg{teams: teams, viewerID: viewerID}
			}
		}
	}
	return m, nil
}

func (m model) markAllAsRead() tea.Cmd {
	ids := make([]string, 0)
	for _, n := range m.notifications {
		if n.ReadAt == nil {
			ids = append(ids, n.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	client := m.client
	return func() tea.Msg {
		for _, id := range ids {
			_ = client.MarkAsRead(id)
		}
		return markAllReadMsg{}
	}
}

// --- View ---

func (m model) View() tea.View {
	var v tea.View
	v.AltScreen = true

	if m.width == 0 || m.height == 0 {
		v.SetContent("")
		return v
	}

	if m.createdIssue != nil {
		v.SetContent(renderCreatedModal(m.createdIssue, m.width, m.height))
		return v
	}

	if m.createForm != nil {
		v.SetContent(m.createForm.View(m.width, m.height))
		return v
	}

	if m.loading {
		v.SetContent(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, "Loading…"))
		return v
	}
	if m.err != nil {
		msg := lipgloss.NewStyle().Foreground(colorMauve).Render("Error: " + m.err.Error())
		v.SetContent(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg))
		return v
	}

	v.SetContent(m.renderFrame())
	return v
}

func (m model) renderFrame() string {
	w := m.width
	// 4 rows consumed: top border, separator, help line, bottom border
	contentH := m.height - 4
	if contentH < 1 {
		contentH = 1
	}

	leftW := w * 2 / 5
	if leftW < 10 {
		leftW = 10
	}
	// 3 border columns: left edge, centre divider, right edge
	rightW := w - leftW - 3
	if rightW < 10 {
		rightW = 10
	}

	b := border

	// Scrollable window: show a slice of notifications around the cursor
	visibleStart, visibleEnd := m.visibleRange(contentH)
	listLines := m.renderListLines(leftW, contentH, visibleStart, visibleEnd)
	detailLines := m.renderDetailLines(rightW, contentH)

	// Top border with title
	top := m.renderTopBorder(w, leftW, rightW)

	// Content rows
	rows := make([]string, contentH)
	for i := 0; i < contentH; i++ {
		l := fitWidth(listLines[i], leftW)
		r := fitWidth(detailLines[i], rightW)
		rows[i] = b.Render("│") + l + b.Render("│") + r + b.Render("│")
	}

	// Separator between content and help
	sep := b.Render("├") + b.Render(strings.Repeat("─", leftW)) +
		b.Render("┴") + b.Render(strings.Repeat("─", rightW)) + b.Render("┤")

	// Help bar
	helpContent := renderHelpContent()
	helpLine := b.Render("│") + " " + helpContent +
		strings.Repeat(" ", max(0, w-lipgloss.Width(helpContent)-4)) + " " + b.Render("│")

	// Bottom border
	bottom := b.Render("╰") + b.Render(strings.Repeat("─", w-2)) + b.Render("╯")

	lines := make([]string, 0, contentH+4)
	lines = append(lines, top)
	lines = append(lines, rows...)
	lines = append(lines, sep, helpLine, bottom)
	return strings.Join(lines, "\n")
}

func (m model) renderTopBorder(w, leftW, rightW int) string {
	b := border

	titleText := " Linear Notifications "
	unread := m.unreadCount
	var statusText string
	if m.flash != "" {
		statusText = " " + m.flash + " "
	} else if unread > 0 {
		statusText = fmt.Sprintf(" %d unread ", unread)
	}

	styledTitle := titleStyle.Render(titleText)
	styledStatus := subtitleStyle.Render(statusText)
	titleVisW := lipgloss.Width(styledTitle)
	statusVisW := lipgloss.Width(styledStatus)

	// ╭─ Title ───┬──── N unread ─╮
	leftFill := leftW - titleVisW - 1 // -1 for "─" before title
	if leftFill < 0 {
		leftFill = 0
	}
	rightFill := rightW - statusVisW
	if rightFill < 0 {
		rightFill = 0
	}

	return b.Render("╭─") + styledTitle + b.Render(strings.Repeat("─", leftFill)) +
		b.Render("┬") + b.Render(strings.Repeat("─", rightFill)) +
		styledStatus + b.Render("─╮")
}

func (m model) visibleRange(contentH int) (int, int) {
	total := len(m.notifications)
	if total <= contentH {
		return 0, total
	}
	start := m.cursor - contentH/2
	if start < 0 {
		start = 0
	}
	end := start + contentH
	if end > total {
		end = total
		start = end - contentH
	}
	return start, end
}

func (m model) renderListLines(w, h, visStart, visEnd int) []string {
	lines := make([]string, h)
	row := 0
	for i := visStart; i < visEnd && row < h; i++ {
		n := m.notifications[i]

		marker := "  "
		if n.ReadAt == nil {
			marker = "● "
		}

		label := notificationLabel(n, w-lipgloss.Width(marker)-1)
		text := marker + label

		if i == m.cursor {
			if n.ReadAt == nil {
				markerStr := unreadDot.Background(colorSurface0).Render("● ")
				labelStr := listItemSelected.Render(label)
				lines[row] = markerStr + labelStr
			} else {
				lines[row] = listItemSelected.Render(text)
			}
		} else {
			if n.ReadAt == nil {
				markerStr := unreadDot.Render("● ")
				labelStr := listItemNormal.Render(label)
				lines[row] = markerStr + labelStr
			} else {
				lines[row] = listItemNormal.Render(text)
			}
		}
		row++
	}
	for i := row; i < h; i++ {
		lines[i] = ""
	}
	return lines
}

func (m model) renderDetailLines(w, h int) []string {
	if len(m.notifications) == 0 {
		lines := make([]string, h)
		lines[0] = detailLabel.Render("No notifications")
		for i := 1; i < h; i++ {
			lines[i] = ""
		}
		return lines
	}

	n := m.notifications[m.cursor]
	var parts []string
	sep := detailSep.Render(strings.Repeat("─", w-1))

	// Type badge
	typeName := humanizeType(n.Type)
	parts = append(parts, detailType.Render(typeName))
	parts = append(parts, "")

	// Title section
	if n.Issue != nil {
		parts = append(parts, detailIdent.Render(n.Issue.Identifier)+" "+detailTitle.Render(truncate(n.Issue.Title, w-len(n.Issue.Identifier)-2)))
		parts = append(parts, "")

		if n.Issue.Team != nil {
			parts = append(parts, detailIcon.Render("󰏬 ")+detailLabel.Render("Team     ")+detailValue.Render(n.Issue.Team.Name))
		}
		if n.Issue.State != nil {
			parts = append(parts, detailIcon.Render("󰓦 ")+detailLabel.Render("State    ")+detailState.Render(n.Issue.State.Name))
		}
		if n.Issue.Assignee != nil {
			parts = append(parts, detailIcon.Render("󰀄 ")+detailLabel.Render("Assignee ")+detailActor.Render(n.Issue.Assignee.Name))
		}
		if n.Issue.Labels != nil && len(n.Issue.Labels.Nodes) > 0 {
			tags := make([]string, len(n.Issue.Labels.Nodes))
			for j, l := range n.Issue.Labels.Nodes {
				tags[j] = detailTag.Render(l.Name)
			}
			parts = append(parts, detailIcon.Render("󰓹 ")+strings.Join(tags, " "))
		}
	} else if n.PullRequest != nil {
		parts = append(parts, detailIdent.Render("⑂ ")+detailTitle.Render(truncate(n.PullRequest.Title, w-3)))
		parts = append(parts, "")

		if n.PullRequest.Status != "" {
			parts = append(parts, detailIcon.Render("󰓦 ")+detailLabel.Render("Status   ")+detailState.Render(n.PullRequest.Status))
		}
		if n.PullRequest.SourceBranch != "" {
			parts = append(parts, detailIcon.Render("󰘬 ")+detailLabel.Render("Branch   ")+detailValue.Render(
				truncate(n.PullRequest.SourceBranch, w-22))+detailLabel.Render(" → ")+detailValue.Render(n.PullRequest.TargetBranch))
		}
	} else if n.Project != nil {
		parts = append(parts, detailIdent.Render("◈ ")+detailTitle.Render(truncate(n.Project.Name, w-3)))
		parts = append(parts, "")
	} else if n.Title != "" {
		parts = append(parts, detailTitle.Render(truncate(n.Title, w-1)))
		parts = append(parts, "")
	}

	// Actor + time
	if n.Actor != nil || n.CreatedAt != "" {
		parts = append(parts, sep)
		if n.Actor != nil {
			parts = append(parts, detailIcon.Render("󰀄 ")+detailActor.Render(n.Actor.Name)+
				detailLabel.Render("  ·  ")+detailTime.Render(timeAgo(n.CreatedAt)))
		} else if n.CreatedAt != "" {
			parts = append(parts, detailTime.Render(timeAgo(n.CreatedAt)))
		}
	}

	// Comment quote block — from issue comment or PR subtitle
	commenter, body := extractComment(n)
	if body != "" {
		parts = append(parts, sep)
		parts = append(parts, detailCommenter.Render(commenter)+" "+detailLabel.Render("commented:"))
		parts = append(parts, "")

		body = cleanMarkdown(body)
		body = wrapText(body, w-4)
		maxLines := h - len(parts) - 1
		if maxLines < 3 {
			maxLines = 3
		}
		qbar := detailQuoteBar.Render("▎ ")
		for i, line := range strings.Split(body, "\n") {
			if i >= maxLines {
				parts = append(parts, qbar+commentBody.Render("…"))
				break
			}
			parts = append(parts, qbar+commentBody.Render(line))
		}
	}

	// Pad or trim to exactly h lines
	lines := make([]string, h)
	for i := 0; i < h; i++ {
		if i < len(parts) {
			lines[i] = " " + parts[i]
		} else {
			lines[i] = ""
		}
	}
	return lines
}

func renderHelpContent() string {
	keys := []struct{ key, desc string }{
		{"enter", "open"},
		{"r", "mark read"},
		{"R", "mark all read"},
		{"c", "create"},
		{"q", "quit"},
	}
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = helpKey.Render(k.key) + " " + helpDesc.Render(k.desc)
	}
	return strings.Join(parts, helpSep.Render(" │ "))
}

// --- Helpers ---

func copyToClipboard(text string) error {
	cmd := exec.Command("wl-copy", text)
	return cmd.Run()
}

func countUnread(ns []Notification) int {
	c := 0
	for _, n := range ns {
		if n.ReadAt == nil {
			c++
		}
	}
	return c
}

func notificationLabel(n Notification, maxW int) string {
	if n.Issue != nil {
		return n.Issue.Identifier + " " + truncate(n.Issue.Title, maxW-len(n.Issue.Identifier)-1)
	}
	if n.Title != "" {
		return truncate(n.Title, maxW)
	}
	return humanizeType(n.Type)
}

func extractComment(n Notification) (commenter string, body string) {
	// Prefer the structured issue comment
	if n.Comment != nil && n.Comment.Body != "" {
		commenter = "Someone"
		if n.Comment.User != nil {
			commenter = n.Comment.User.Name
		}
		return commenter, n.Comment.Body
	}
	// Fall back to parsing the subtitle ("author commented: body" / "author reviewed ...: body")
	if n.Subtitle != "" {
		if idx := strings.Index(n.Subtitle, " commented: "); idx > 0 {
			return n.Subtitle[:idx], n.Subtitle[idx+len(" commented: "):]
		}
		if idx := strings.Index(n.Subtitle, " reviewed "); idx > 0 {
			return n.Subtitle[:idx], n.Subtitle[idx+1:]
		}
	}
	return "", ""
}

func notificationURL(n Notification) string {
	if n.Issue != nil && n.Issue.URL != "" {
		return n.Issue.URL
	}
	if n.URL != "" {
		return n.URL
	}
	return "https://linear.app/inbox"
}

func humanizeType(t string) string {
	types := map[string]string{
		"issueAssignedToYou":  "Assigned to you",
		"issueComment":        "Comment",
		"issueMention":        "Mention",
		"issueCommentMention": "Comment mention",
		"issueStatusChanged":  "Status changed",
		"issueNewComment":     "New comment",
		"issuePriorityChanged": "Priority changed",
		"issueBlocking":       "Blocking",
		"issueUnblocked":      "Unblocked",
		"issueDue":            "Due date",
		"issueSubscribed":     "Subscribed",
		"issueCreated":        "Issue created",
		"issueUpdated":        "Issue updated",
		"projectUpdate":       "Project update",
		"pullRequestCommented": "New comment",
		"projectUpdatePrompt": "Project update reminder",
		"issueSlaHighRisk": "SLA high risk",
		"issueSlaBreached": "SLA breached",
	}
	if s, ok := types[t]; ok {
		return s
	}
	return t
}

func timeAgo(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

func truncate(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= maxW {
		return s
	}
	if maxW <= 1 {
		return "…"
	}
	return string(r[:maxW-1]) + "…"
}

func fitWidth(s string, w int) string {
	return lipgloss.NewStyle().Width(w).MaxWidth(w).Render(s)
}

func wrapText(s string, w int) string {
	if w <= 0 {
		return s
	}
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		if lipgloss.Width(paragraph) <= w {
			lines = append(lines, paragraph)
			continue
		}
		words := strings.Fields(paragraph)
		current := ""
		for _, word := range words {
			if current == "" {
				if lipgloss.Width(word) > w {
					current = truncate(word, w)
				} else {
					current = word
				}
			} else if lipgloss.Width(current+" "+word) <= w {
				current += " " + word
			} else {
				lines = append(lines, current)
				if lipgloss.Width(word) > w {
					current = truncate(word, w)
				} else {
					current = word
				}
			}
		}
		if current != "" {
			lines = append(lines, current)
		}
	}
	return strings.Join(lines, "\n")
}

var (
	reImage   = regexp.MustCompile(`!\[[^\]]*\]\([^)]+\)`)
	reLink    = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	reBareURL = regexp.MustCompile(`https?://\S{60,}`)
	reMultiNL = regexp.MustCompile(`\n{3,}`)
)

func cleanMarkdown(s string) string {
	s = reImage.ReplaceAllString(s, "[image]")
	s = reLink.ReplaceAllString(s, "$1")
	s = reBareURL.ReplaceAllStringFunc(s, func(u string) string {
		if len(u) > 50 {
			return u[:47] + "…"
		}
		return u
	})
	s = reMultiNL.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

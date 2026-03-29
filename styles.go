package main

import "charm.land/lipgloss/v2"

// Catppuccin Mocha palette
var (
	colorBase     = lipgloss.Color("#1E1E2E")
	colorSurface0 = lipgloss.Color("#313244")
	colorSurface1 = lipgloss.Color("#45475A")
	colorOverlay0 = lipgloss.Color("#6C7086")
	colorSubtext0 = lipgloss.Color("#A6ADC8")
	colorText     = lipgloss.Color("#CDD6F4")
	colorGreen    = lipgloss.Color("#A6E3A1")
	colorMauve    = lipgloss.Color("#CBA6F7")
	colorBlue     = lipgloss.Color("#89B4FA")
	colorPeach    = lipgloss.Color("#FAB387")
	colorYellow   = lipgloss.Color("#F9E2AF")
	colorTeal     = lipgloss.Color("#94E2D5")
	colorFlamingo = lipgloss.Color("#F2CDCD")
)

var (
	borderColor = colorSurface1

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorText)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	listItemNormal = lipgloss.NewStyle().
			Foreground(colorSubtext0)

	listItemSelected = lipgloss.NewStyle().
				Foreground(colorText).
				Background(colorSurface0).
				Bold(true)

	unreadDot = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	readDot = lipgloss.NewStyle().
		Foreground(colorSurface0)

	// Detail pane: notification type badge
	detailType = lipgloss.NewStyle().
			Foreground(colorBase).
			Background(colorMauve).
			Bold(true).
			Padding(0, 1)

	// Detail pane: issue/PR identifier (colored separately from title)
	detailIdent = lipgloss.NewStyle().
			Foreground(colorBlue).
			Bold(true)

	// Detail pane: main title text
	detailTitle = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	// Detail pane: field icons
	detailIcon = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	// Detail pane: field labels
	detailLabel = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	// Detail pane: field values
	detailValue = lipgloss.NewStyle().
			Foreground(colorSubtext0)

	// Detail pane: separator line
	detailSep = lipgloss.NewStyle().
			Foreground(colorSurface1)

	// Detail pane: actor name
	detailActor = lipgloss.NewStyle().
			Foreground(colorFlamingo)

	// Detail pane: time ago
	detailTime = lipgloss.NewStyle().
			Foreground(colorOverlay0).
			Italic(true)

	// Detail pane: comment attribution
	detailCommenter = lipgloss.NewStyle().
			Foreground(colorPeach).
			Bold(true)

	// Detail pane: comment quote bar
	detailQuoteBar = lipgloss.NewStyle().
			Foreground(colorSurface1)

	// Detail pane: comment body text
	commentBody = lipgloss.NewStyle().
			Foreground(colorSubtext0)

	// Detail pane: label tags
	detailTag = lipgloss.NewStyle().
			Foreground(colorBase).
			Background(colorSurface1).
			Padding(0, 1)

	// Detail pane: state with color
	detailState = lipgloss.NewStyle().
			Foreground(colorTeal)

	// Help bar
	helpKey = lipgloss.NewStyle().
		Foreground(colorText).
		Bold(true)

	helpDesc = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	helpSep = lipgloss.NewStyle().
		Foreground(colorSurface1)

	border = lipgloss.NewStyle().
		Foreground(borderColor)

	// Create form styles
	formTitle = lipgloss.NewStyle().
			Foreground(colorBase).
			Background(colorMauve).
			Bold(true).
			Padding(0, 1)

	formLabel = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	formLabelFocused = lipgloss.NewStyle().
				Foreground(colorMauve).
				Bold(true)

	formInputBorder = lipgloss.NewStyle().
			Foreground(colorSurface1)

	formInputBorderFocused = lipgloss.NewStyle().
				Foreground(colorMauve)

	formCursor = lipgloss.NewStyle().
			Foreground(colorBase).
			Background(colorText)

	formTeam = lipgloss.NewStyle().
			Foreground(colorSubtext0)

	formTeamFocused = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	formArrow = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	formError = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F38BA8"))

	formHelpKey = lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true)

	formHelpDesc = lipgloss.NewStyle().
			Foreground(colorOverlay0)

	formHelpSep = lipgloss.NewStyle().
			Foreground(colorSurface1)
)

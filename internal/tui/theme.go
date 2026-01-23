package tui

import "github.com/charmbracelet/lipgloss"

var (
	asciiBorder = lipgloss.Border{
		Top:         "-",
		Bottom:      "-",
		Left:        "|",
		Right:       "|",
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Middle:      "+",
		MiddleLeft:  "+",
		MiddleRight: "+",
	}

	panelStyle = lipgloss.NewStyle().Border(asciiBorder).Padding(0, 1)

	titleStyle = lipgloss.NewStyle().Bold(true)
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	goodStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#2E8540")).Bold(true)
	badStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#B71C1C")).Bold(true)
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F9A825")).Bold(true)
)

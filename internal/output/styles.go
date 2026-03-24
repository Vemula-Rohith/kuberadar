package output

import (
	"charm.land/lipgloss/v2"
)

// Icons for terminal output (Unicode symbols with broad terminal support)
const (
	IconSweep     = "⊙"  // Cluster sweep / scan
	IconSummary   = "▸"  // Summary section
	IconCritical  = "✗"  // Critical severity
	IconWarning   = "⚠"  // Warning severity
	IconOK        = "✓"  // No issues / success
	IconEvidence  = "▸"  // Evidence section (match Summary / Recommendation)
	IconRecommend = "▸"  // Recommendation section
	IconResource  = "◉"  // Resource / pod
)

// Colors
var (
	colorPurple   = lipgloss.Color("99")
	colorRed      = lipgloss.Color("9")
	colorYellow   = lipgloss.Color("11")
	colorGreen    = lipgloss.Color("10")
	colorGray     = lipgloss.Color("245")
	colorLightGray = lipgloss.Color("241")
)

// Styles for output
var (
	// TitleStyle for main headers (e.g. "Cluster Sweep Results")
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPurple)

	// SummaryHeaderStyle for "Summary" section header
	SummaryHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorPurple)

	// CriticalStyle for critical severity text
	CriticalStyle = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	// WarningStyle for warning severity text
	WarningStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	// SuccessStyle for no-issues / OK state
	SuccessStyle = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	// SectionStyle for section headers (Evidence, Recommendation)
	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorGray)

	// TableHeaderStyle for table column headers
	TableHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPurple)

	// OddRowStyle for alternating table rows (optional)
	OddRowStyle = lipgloss.NewStyle().Foreground(colorGray)

	// EvenRowStyle for alternating table rows (optional)
	EvenRowStyle = lipgloss.NewStyle().Foreground(colorLightGray)

	// RootCauseStyle highlights the identified root cause for faster debugging
	RootCauseStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorYellow).
			Background(lipgloss.Color("236"))
)

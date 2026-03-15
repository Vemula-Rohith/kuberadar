package output

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"charm.land/lipgloss/v2"

	"github.com/kuberadar/kuberadar/internal/constants"
	"github.com/kuberadar/kuberadar/internal/model"
)

const evidenceWrapWidth = 100 // wrap long lines at word boundaries

func uniqueResourceCount(issues []model.Issue) int {
	seen := make(map[string]bool)
	for _, i := range issues {
		key := i.ResourceKind + "/" + i.ResourceName
		seen[key] = true
	}
	return len(seen)
}

func severityCounts(issues []model.Issue) (critical, warning int) {
	for _, i := range issues {
		switch i.Severity {
		case constants.SeverityCritical:
			critical++
		case constants.SeverityWarning:
			warning++
		}
	}
	return critical, warning
}

// wrapEvidence wraps long lines at word boundaries. Uses rune-aware slicing to
// avoid breaking UTF-8 characters. Continuation lines align under the bullet content.
func wrapEvidence(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		maxWidth = evidenceWrapWidth
	}
	var out []string
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for _, line := range lines {
		// Preserve empty lines for subsection spacing
		if line == "" {
			out = append(out, "")
			continue
		}
		for runeCount(line) > maxWidth {
			// Find cut point: last space before maxWidth, or maxWidth (rune-aware)
			runes := []rune(line)
			cutRunes := maxWidth
			if len(runes) > maxWidth {
				// Look for word boundary in first maxWidth runes
				search := string(runes[:maxWidth])
				if idx := strings.LastIndex(search, " "); idx > maxWidth/2 {
					cutRunes = idx + 1
				}
			}
			out = append(out, string(runes[:cutRunes]))
			line = strings.TrimLeft(string(runes[cutRunes:]), " ")
		}
		out = append(out, line)
	}
	return out
}

func runeCount(s string) int { return utf8.RuneCountInString(s) }

func renderTitle(s string) string {
	return TitleStyle.Render(s)
}

func renderSummaryHeader() string {
	return SummaryHeaderStyle.Render(IconSummary + " Summary") + "\n─────────\n"
}

func renderSeverity(severity string) string {
	switch severity {
	case constants.SeverityCritical:
		return CriticalStyle.Render(IconCritical + " " + severity)
	case constants.SeverityWarning:
		return WarningStyle.Render(IconWarning + " " + severity)
	default:
		return severity
	}
}

// padCell pads s to visible width w (handles ANSI codes correctly).
func padCell(s string, w int) string {
	return lipgloss.NewStyle().Width(w).Render(s)
}

// truncateByWidth truncates s to max visible width, appending "..." if needed.
func truncateByWidth(s string, max int) string {
	if lipgloss.Width(s) <= max {
		return s
	}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		if lipgloss.Width(string(runes[:i])+"...") <= max {
			return string(runes[:i]) + "..."
		}
	}
	return "..."
}

// FormatAsTable renders a Diagnosis as a human-readable table.
func FormatAsTable(d *model.Diagnosis, opts Options) string {
	var b strings.Builder

	if len(d.Issues) == 0 {
		if !opts.SinglePod && d.PodsScanned > 0 {
			b.WriteString(renderTitle(IconSweep + " Cluster Sweep Results"))
			b.WriteString("\n\n")
			b.WriteString(renderSummaryHeader())
			b.WriteString(fmt.Sprintf("Pods scanned: %d   Affected: 0   Issues: 0   (%s Critical: 0, %s Warning: 0)\n\n",
				d.PodsScanned, IconCritical, IconWarning))
		} else {
			b.WriteString(SuccessStyle.Render(IconOK + " No issues found."))
		}
		return b.String()
	}

	// Diagnose mode: full evidence and recommendation with improved spacing
	if opts.SinglePod && opts.Diagnose {
		for i, issue := range d.Issues {
			if i > 0 {
				b.WriteString("\n\n")
			}
			resource := IconResource + " " + issue.ResourceKind + "/" + issue.ResourceName
			b.WriteString(TableHeaderStyle.Render(resource) + "\n")
			b.WriteString(fmt.Sprintf("%s Issue: %s (%s)\n\n", IconSweep, issue.ID, renderSeverity(issue.Severity)))
			b.WriteString(issue.Message + "\n\n")
			if issue.LikelyCause != "" {
				b.WriteString(SectionStyle.Render("Likely cause") + "\n")
				b.WriteString("────────────\n")
				for _, line := range wrapEvidence(issue.LikelyCause, evidenceWrapWidth) {
					b.WriteString("  " + line + "\n")
				}
				b.WriteString("\n")
			}
			if len(issue.Evidence) > 0 {
				b.WriteString(SectionStyle.Render(IconEvidence + " Evidence") + "\n")
				b.WriteString("────────\n")
				indent := "  "
				for _, e := range issue.Evidence {
					lines := wrapEvidence(e.Line, evidenceWrapWidth)
					for _, line := range lines {
						if line == "" {
							b.WriteString("\n")
						} else {
							if e.RootCause {
								b.WriteString(indent + RootCauseStyle.Render(line) + "\n")
							} else {
								b.WriteString(indent + line + "\n")
							}
						}
					}
				}
				b.WriteString("\n")
			}
			if issue.Recommendation != "" {
				b.WriteString(SectionStyle.Render(IconRecommend + " Recommendation") + "\n")
				b.WriteString("──────────────\n")
				b.WriteString(issue.Recommendation + "\n")
			}
		}
		return strings.TrimSuffix(b.String(), "\n")
	}

	// Summary header only for sweep-style (multiple pods)
	if !opts.SinglePod && d.PodsScanned > 0 {
		affected := uniqueResourceCount(d.Issues)
		critical, warning := severityCounts(d.Issues)
		b.WriteString(renderTitle(IconSweep + " Cluster Sweep Results"))
		b.WriteString("\n\n")
		b.WriteString(renderSummaryHeader())
		b.WriteString(fmt.Sprintf("Pods scanned: %d   Affected: %d   Issues: %d   (%s Critical: %d, %s Warning: %d)\n\n",
			d.PodsScanned, affected, len(d.Issues), IconCritical, critical, IconWarning, warning))
	}

	// Compact table (sweep or single-pod default)
	const (
		colResource = 27 // IconResource + space + "Pod/ns/name"
		colIssue    = 8
		colSeverity = 16 // "✗ Critical" / "⚠ Warning"
		colMessage  = 44
	)
	headerResource := padCell(TableHeaderStyle.Render("RESOURCE"), colResource)
	headerIssue := padCell(TableHeaderStyle.Render("ISSUE"), colIssue)
	headerSeverity := padCell(TableHeaderStyle.Render("SEVERITY"), colSeverity)
	headerMessage := TableHeaderStyle.Render("MESSAGE")
	if opts.SinglePod {
		b.WriteString(headerResource + " " + headerIssue + " " + headerSeverity + "\n")
		b.WriteString(strings.Repeat("─", colResource+colIssue+colSeverity+2) + "\n")
	} else {
		b.WriteString(headerResource + " " + headerIssue + " " + headerSeverity + "  " + headerMessage + "\n")
		b.WriteString(strings.Repeat("─", colResource+colIssue+colSeverity+colMessage+3) + "\n")
	}
	for _, issue := range d.Issues {
		resource := IconResource + " " + issue.ResourceKind + "/" + issue.ResourceName
		resource = truncateByWidth(resource, colResource)
		resource = padCell(resource, colResource)
		issueID := padCell(issue.ID, colIssue)
		sev := padCell(renderSeverity(issue.Severity), colSeverity)
		if opts.SinglePod {
			b.WriteString(resource + " " + issueID + " " + sev + "\n")
		} else {
			msg := issue.Message
			if len(msg) > colMessage {
				msg = msg[:colMessage-3] + "..."
			}
			b.WriteString(resource + " " + issueID + " " + sev + "  " + msg + "\n")
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

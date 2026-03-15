package model

import "strings"

// Evidence represents a single evidence line with optional root cause flag.
type Evidence struct {
	Line      string `json:"Line"`
	RootCause bool   `json:"RootCause,omitempty"`
}

// EvidenceFromBlock splits a multi-line block into Evidence items, marking the root cause line.
func EvidenceFromBlock(block, rootCause string) []Evidence {
	lines := strings.Split(block, "\n")
	out := make([]Evidence, 0, len(lines))
	for _, line := range lines {
		rc := rootCause != "" && (line == rootCause || strings.TrimSpace(line) == rootCause)
		out = append(out, Evidence{Line: line, RootCause: rc})
	}
	return out
}

// EvidenceLine creates a single Evidence item.
func EvidenceLine(line string, rootCause bool) Evidence {
	return Evidence{Line: line, RootCause: rootCause}
}

// Issue represents a detected problem with resource metadata.
type Issue struct {
	ID             string     `json:"ID"`
	Severity       string     `json:"Severity"`
	ResourceKind   string     `json:"ResourceKind"`
	ResourceName   string     `json:"ResourceName"`
	Message        string     `json:"Message"`
	LikelyCause    string     `json:"LikelyCause,omitempty"`
	Evidence       []Evidence `json:"Evidence"`
	Recommendation string     `json:"Recommendation"`
}

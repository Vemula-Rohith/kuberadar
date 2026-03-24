package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

const sweepFileName = "last-sweep.json"

type sweepFile struct {
	Version int         `json:"version"`
	Entries []sweepRef  `json:"entries"`
}

type sweepRef struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}

// WriteLastSweep saves pod refs from a sweep (same order as issue rows) for `kuberadar pod <n>`.
func WriteLastSweep(entries []SweepEntry) error {
	path, err := sweepStatePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	refs := make([]sweepRef, len(entries))
	for i := range entries {
		refs[i] = sweepRef{Namespace: entries[i].Namespace, Pod: entries[i].Pod}
	}
	data, err := json.MarshalIndent(sweepFile{Version: 1, Entries: refs}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// SweepEntry identifies a pod from a sweep row.
type SweepEntry struct {
	Namespace string
	Pod       string
}

// EntriesFromDiagnosis builds sweep rows in issue order for persistence.
func EntriesFromDiagnosis(d *model.Diagnosis) []SweepEntry {
	if d == nil || len(d.Issues) == 0 {
		return nil
	}
	out := make([]SweepEntry, 0, len(d.Issues))
	for _, issue := range d.Issues {
		if ns, pod, ok := ParseResourceName(issue.ResourceName); ok {
			out = append(out, SweepEntry{Namespace: ns, Pod: pod})
		}
	}
	return out
}

// ParseResourceName splits "namespace/pod" from issue ResourceName.
func ParseResourceName(resourceName string) (namespace, pod string, ok bool) {
	i := strings.Index(resourceName, "/")
	if i <= 0 || i >= len(resourceName)-1 {
		return "", "", false
	}
	return resourceName[:i], resourceName[i+1:], true
}

// ResolveSweepIndex returns namespace and pod for 1-based index from last sweep.
func ResolveSweepIndex(oneBased int) (namespace, pod string, err error) {
	if oneBased < 1 {
		return "", "", fmt.Errorf("sweep index must be at least 1")
	}
	path, err := sweepStatePath()
	if err != nil {
		return "", "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("no saved sweep; run kuberadar sweep in this namespace first")
		}
		return "", "", fmt.Errorf("read sweep state: %w", err)
	}
	var f sweepFile
	if err := json.Unmarshal(data, &f); err != nil {
		return "", "", fmt.Errorf("invalid sweep state file")
	}
	if oneBased > len(f.Entries) {
		return "", "", fmt.Errorf("sweep index %d out of range (last sweep had %d issue row(s))", oneBased, len(f.Entries))
	}
	e := f.Entries[oneBased-1]
	if e.Namespace == "" || e.Pod == "" {
		return "", "", fmt.Errorf("invalid sweep entry at index %d", oneBased)
	}
	return e.Namespace, e.Pod, nil
}

// IsSweepIndexSyntax returns true if s is all ASCII digits (e.g. "1", "12").
func IsSweepIndexSyntax(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// ParseSweepIndex parses "12" as 12.
func ParseSweepIndex(s string) (int, error) {
	return strconv.Atoi(s)
}

func sweepStatePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "kuberadar", sweepFileName), nil
}

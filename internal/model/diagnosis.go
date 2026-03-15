package model

// Diagnosis is the output of a diagnostic run.
type Diagnosis struct {
	Timestamp   string  `json:"Timestamp,omitempty"`
	PodsScanned int     `json:"PodsScanned,omitempty"`
	Issues      []Issue `json:"Issues"`
	Scope       Scope   `json:"Scope"`
}

// Add appends issues to the diagnosis.
func (d *Diagnosis) Add(issues ...Issue) {
	d.Issues = append(d.Issues, issues...)
}

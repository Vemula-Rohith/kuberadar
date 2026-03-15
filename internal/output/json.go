package output

import (
	"encoding/json"
	"io"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// WriteJSON writes a Diagnosis as JSON to w.
func WriteJSON(w io.Writer, d *model.Diagnosis) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(d)
}

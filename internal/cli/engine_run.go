package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/engine"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

func stderrForCmd(cmd *cobra.Command) *os.File {
	if cmd == nil {
		return os.Stderr
	}
	w := cmd.ErrOrStderr()
	if w == nil {
		return os.Stderr
	}
	if f, ok := w.(*os.File); ok {
		return f
	}
	return os.Stderr
}

// engineRun executes a diagnosis with a single-line spinner on stderr (TTY only).
func engineRun(cmd *cobra.Command, scope model.Scope) (*model.Diagnosis, error) {
	sp := NewSpinner(stderrForCmd(cmd))
	sp.Start("Scanning cluster...")
	defer sp.Stop()
	return app.engine.Run(context.Background(), scope, &engine.RunOpts{
		OnProgress: sp.SetMessage,
	})
}

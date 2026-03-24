package cli

import (
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Spinner renders a single updating status line on stderr; cleared on Stop (TTY only).
type Spinner struct {
	out     *os.File
	enabled bool
	msg     string
	mu      sync.Mutex
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewSpinner returns a spinner; animation runs only when out is a terminal.
func NewSpinner(out *os.File) *Spinner {
	if out == nil {
		return &Spinner{enabled: false}
	}
	enabled := term.IsTerminal(int(out.Fd()))
	return &Spinner{out: out, enabled: enabled}
}

// Start begins animation with the initial message. No-op if not a TTY.
func (s *Spinner) Start(message string) {
	s.SetMessage(message)
	if !s.enabled {
		return
	}
	s.stopCh = make(chan struct{})
	s.doneCh = make(chan struct{})
	_, _ = io.WriteString(s.out, "\r\x1b[K"+spinnerFrames[0]+" "+message)
	go s.loop()
}

// SetMessage updates the text after the spinner frame.
func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

func (s *Spinner) loop() {
	defer close(s.doneCh)
	tick := time.NewTicker(80 * time.Millisecond)
	defer tick.Stop()
	n := 0
	for {
		select {
		case <-s.stopCh:
			return
		case <-tick.C:
			s.mu.Lock()
			msg := s.msg
			s.mu.Unlock()
			frame := spinnerFrames[n%len(spinnerFrames)]
			n++
			_, _ = io.WriteString(s.out, "\r\x1b[K"+frame+" "+msg)
		}
	}
}

// Stop ends animation and clears the status line.
func (s *Spinner) Stop() {
	if !s.enabled || s.stopCh == nil {
		return
	}
	close(s.stopCh)
	<-s.doneCh
	_, _ = io.WriteString(s.out, "\r\x1b[K")
	s.stopCh = nil
	s.doneCh = nil
}

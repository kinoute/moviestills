package debug

import (
	"io"
	"os"
	"sync/atomic"
	"time"

	log "github.com/pterm/pterm"

	"github.com/gocolly/colly/v2/debug"
)

// PTermDebugger a a debugger which prints log messages to the STDERR
// with the PTerm design pattern.
type PTermDebugger struct {
	// Output is the log destination, anything can be used which implements them
	// io.Writer interface. Leave it blank to use STDERR
	Output io.Writer
	// Prefix appears at the beginning of each generated log line
	Prefix string
	// Flag defines the logging properties.
	counter int32
	start   time.Time
}

// Init initializes the pTermDebugger
func (l *PTermDebugger) Init() error {
	l.counter = 0
	l.start = time.Now()
	if l.Output == nil {
		l.Output = os.Stderr
	}
	return nil
}

// Event receives Collector events and prints them to STDERR
func (l *PTermDebugger) Event(e *debug.Event) {
	i := atomic.AddInt32(&l.counter, 1)
	log.Debug.Printf("[%06d] %d [%6d - %s] %q (%s)\n", i, e.CollectorID, e.RequestID, e.Type, e.Values, time.Since(l.start))
}

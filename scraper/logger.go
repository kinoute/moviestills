package scraper

import (
	"fmt"

	"github.com/pterm/pterm"
)

// Logger provides website-prefixed logging
type Logger struct {
	prefix string
}

// NewLogger creates a logger with a website prefix
func NewLogger(website string) *Logger {
	return &Logger{
		prefix: fmt.Sprintf("[%s]", website),
	}
}

// Info logs an info message with website prefix
func (l *Logger) Info(args ...interface{}) {
	pterm.Info.Println(append([]interface{}{pterm.Cyan(l.prefix)}, args...)...)
}

// Debug logs a debug message with website prefix
func (l *Logger) Debug(args ...interface{}) {
	pterm.Debug.Println(append([]interface{}{pterm.Cyan(l.prefix)}, args...)...)
}

// Error logs an error message with website prefix
func (l *Logger) Error(args ...interface{}) {
	pterm.Error.Println(append([]interface{}{pterm.Cyan(l.prefix)}, args...)...)
}

// Success logs a success message with website prefix
func (l *Logger) Success(args ...interface{}) {
	pterm.Success.Println(append([]interface{}{pterm.Cyan(l.prefix)}, args...)...)
}

// Warning logs a warning message with website prefix
func (l *Logger) Warning(args ...interface{}) {
	pterm.Warning.Println(append([]interface{}{pterm.Cyan(l.prefix)}, args...)...)
}

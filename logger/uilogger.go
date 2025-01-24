package logger

import (
	"fmt"
	"sync"
)

var GlobalUILogger *UILogger

// InitGlobalLogger initializes the global logger
func InitGlobalLogger(ui UI) {
	GlobalUILogger = NewUILogger(ui)
}

// UILogger implements io.Writer and safely forwards logs to the UI
type UILogger struct {
	ui UI
	mu sync.Mutex // Protects concurrent access to the UI

}

func NewUILogger(ui UI) *UILogger {
	return &UILogger{
		ui: ui,
	}
}

// Write implements io.Writer
func (l *UILogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ui != nil {
		msg := string(p)
		l.ui.DisplayLog("%s", msg)
	}
	return len(p), nil
}

func (l *UILogger) Log(format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ui != nil {
		msg := fmt.Sprintf(format, args...)
		l.ui.DisplayLog("%s", msg)
	}
}

func (l *UILogger) Error(format string, args ...interface{}) {
	l.Log("[red]"+format, args...)
}

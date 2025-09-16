package logger

import (
	"fmt"
	"log"
	"runtime"

	"golang.org/x/sys/windows/svc/eventlog"
)

// WindowsEventLogger wraps Windows Event Log functionality
type WindowsEventLogger struct {
	elog *eventlog.Log
}

// Event source name for the application
const EventSourceName = "ClusterEyeAgent"

// Event IDs for different log types
const (
	EventIDInfo    = 1
	EventIDWarning = 2
	EventIDError   = 3
	EventIDFatal   = 4
)

// NewWindowsEventLogger creates a new Windows Event Logger
func NewWindowsEventLogger() (*WindowsEventLogger, error) {
	// Only available on Windows
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("Windows Event Log is only available on Windows")
	}

	// Try to install the event source if it doesn't exist
	err := eventlog.InstallAsEventCreate(EventSourceName, eventlog.Info|eventlog.Warning|eventlog.Error)
	if err != nil {
		// If installation fails, it might already be installed - continue
		log.Printf("Event source installation failed (might already exist): %v", err)
	}

	// Open the event log
	elog, err := eventlog.Open(EventSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open Windows Event Log: %v", err)
	}

	return &WindowsEventLogger{elog: elog}, nil
}

// Close closes the Windows Event Log handle
func (w *WindowsEventLogger) Close() error {
	if w.elog != nil {
		return w.elog.Close()
	}
	return nil
}

// LogInfo logs an information message to Windows Event Log
func (w *WindowsEventLogger) LogInfo(msg string) error {
	if w.elog == nil {
		return fmt.Errorf("Windows Event Log not initialized")
	}
	return w.elog.Info(EventIDInfo, msg)
}

// LogWarning logs a warning message to Windows Event Log
func (w *WindowsEventLogger) LogWarning(msg string) error {
	if w.elog == nil {
		return fmt.Errorf("Windows Event Log not initialized")
	}
	return w.elog.Warning(EventIDWarning, msg)
}

// LogError logs an error message to Windows Event Log
func (w *WindowsEventLogger) LogError(msg string) error {
	if w.elog == nil {
		return fmt.Errorf("Windows Event Log not initialized")
	}
	return w.elog.Error(EventIDError, msg)
}

// LogFatal logs a fatal message to Windows Event Log
func (w *WindowsEventLogger) LogFatal(msg string) error {
	if w.elog == nil {
		return fmt.Errorf("Windows Event Log not initialized")
	}
	return w.elog.Error(EventIDFatal, msg)
}
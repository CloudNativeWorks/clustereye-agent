//go:build !windows
// +build !windows

package logger

import "fmt"

// WindowsEventLogger is a stub for non-Windows platforms
type WindowsEventLogger struct{}

// NewWindowsEventLogger returns an error on non-Windows platforms
func NewWindowsEventLogger() (*WindowsEventLogger, error) {
	return nil, fmt.Errorf("Windows Event Log is not available on this platform")
}

// Close is a no-op on non-Windows platforms
func (w *WindowsEventLogger) Close() error {
	return nil
}

// LogInfo is a no-op on non-Windows platforms
func (w *WindowsEventLogger) LogInfo(msg string) error {
	return fmt.Errorf("Windows Event Log is not available on this platform")
}

// LogWarning is a no-op on non-Windows platforms
func (w *WindowsEventLogger) LogWarning(msg string) error {
	return fmt.Errorf("Windows Event Log is not available on this platform")
}

// LogError is a no-op on non-Windows platforms
func (w *WindowsEventLogger) LogError(msg string) error {
	return fmt.Errorf("Windows Event Log is not available on this platform")
}

// LogFatal is a no-op on non-Windows platforms
func (w *WindowsEventLogger) LogFatal(msg string) error {
	return fmt.Errorf("Windows Event Log is not available on this platform")
}
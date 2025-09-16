package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

// FileLogger handles enhanced file logging
type FileLogger struct {
	logDir      string
	logFile     string
	maxLogSize  int64
	maxLogFiles int
}

// NewFileLogger creates a new file logger with improved directory structure
func NewFileLogger() (*FileLogger, error) {
	var logDir string
	var err error

	if runtime.GOOS == "windows" {
		// Windows: Use ProgramData directory for better organization
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = "C:\\ProgramData"
		}
		logDir = filepath.Join(programData, "ClusterEye", "Agent", "logs")
	} else {
		// Linux/Unix: Use /var/log/clustereye or user home as fallback
		logDir = "/var/log/clustereye"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			// Fallback to user home directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get user home directory: %v", err)
			}
			logDir = filepath.Join(homeDir, ".clustereye", "logs")
		}
	}

	// Create log directory if it doesn't exist
	if err = os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %v", logDir, err)
	}

	return &FileLogger{
		logDir:      logDir,
		logFile:     filepath.Join(logDir, "agent.log"),
		maxLogSize:  100 * 1024 * 1024, // 100MB
		maxLogFiles: 10,                 // Keep 10 log files
	}, nil
}

// GetLogFile returns the current log file path
func (f *FileLogger) GetLogFile() string {
	return f.logFile
}

// GetLogDir returns the log directory path
func (f *FileLogger) GetLogDir() string {
	return f.logDir
}

// RotateLogIfNeeded checks if log rotation is needed and performs it
func (f *FileLogger) RotateLogIfNeeded() error {
	// Check if current log file exists and its size
	info, err := os.Stat(f.logFile)
	if err != nil {
		// File doesn't exist, no rotation needed
		return nil
	}

	if info.Size() < f.maxLogSize {
		// File is within size limit, no rotation needed
		return nil
	}

	// Perform rotation
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := filepath.Join(f.logDir, fmt.Sprintf("agent_%s.log", timestamp))

	if err := os.Rename(f.logFile, backupFile); err != nil {
		return fmt.Errorf("log rotation failed: %v", err)
	}

	// Clean up old log files
	return f.cleanupOldLogs()
}

// cleanupOldLogs removes old log files keeping only the most recent ones
func (f *FileLogger) cleanupOldLogs() error {
	// Find all backup log files
	pattern := filepath.Join(f.logDir, "agent_*.log")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to find old log files: %v", err)
	}

	// Sort files by modification time (oldest first)
	sort.Slice(matches, func(i, j int) bool {
		infoI, errI := os.Stat(matches[i])
		infoJ, errJ := os.Stat(matches[j])

		if errI != nil || errJ != nil {
			return false
		}

		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Remove excess files
	if len(matches) > f.maxLogFiles {
		for i := 0; i < len(matches)-f.maxLogFiles; i++ {
			if err := os.Remove(matches[i]); err != nil {
				// Log the error but don't fail the cleanup
				fmt.Printf("Failed to remove old log file %s: %v\n", matches[i], err)
			}
		}
	}

	return nil
}

// OpenLogFile opens or creates the log file for writing
func (f *FileLogger) OpenLogFile() (*os.File, error) {
	// Rotate if needed before opening
	if err := f.RotateLogIfNeeded(); err != nil {
		return nil, err
	}

	// Open or create the log file
	file, err := os.OpenFile(f.logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %v", f.logFile, err)
	}

	return file, nil
}
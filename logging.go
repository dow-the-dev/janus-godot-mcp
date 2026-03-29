// logging.go
package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func setLogLevel(levelStr string) {
	switch strings.ToUpper(levelStr) {
	case "DEBUG":
		logLevel = 2
	case "INFO":
		logLevel = 1
	default:
		logLevel = 0 // Default to ERROR only
	}
}

func logMessage(level int, prefix, format string, args ...any) {
	if logLevel >= level {
		msg := fmt.Sprintf(format, args...)
		timestamp := time.Now().Format("15:04:05.000")
		fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", timestamp, prefix, msg)

		// --- THE FIX ---
		// Force the terminal to show the text NOW
		os.Stderr.Sync()
	}
}

func logError(format string, args ...any) { logMessage(0, "ERROR", format, args...) }
func logInfo(format string, args ...any)  { logMessage(1, "INFO ", format, args...) }
func logDebug(format string, args ...any) { logMessage(2, "DEBUG", format, args...) }
func logWarning(format string, args ...any) { logMessage(1, "WARN ", format, args...) }
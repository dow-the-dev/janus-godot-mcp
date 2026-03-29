// utils.go
package main

import (
	"os/exec"
	"syscall"
)

func detachProcess(cmd *exec.Cmd) {
	// This constant tells Windows to start the process in a new console/detached state
	const CREATE_BREAKAWAY_FROM_JOB = 0x01000000
	const DETACHED_PROCESS = 0x00000008

	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: DETACHED_PROCESS | CREATE_BREAKAWAY_FROM_JOB,
	}
}

// --- MCP Helpers ---

func mcpText(text string) any {
	return map[string]any{"content": []map[string]any{{"type": "text", "text": text}}}
}

func mcpError(errText string) any {
	logError("Tool execution failed: %s", errText)
	return map[string]any{
		"isError": true,
		"content": []map[string]any{{"type": "text", "text": errText}},
	}
}
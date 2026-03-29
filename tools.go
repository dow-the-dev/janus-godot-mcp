package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


// executeTool is the entry point called by handleRequest in main.go
func executeTool(paramsRaw json.RawMessage) any {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}

	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		logError("Failed to unmarshal tool params: %v", err)
		return mcpError("Failed to parse tool arguments")
	}

	logInfo("Tool Executing: [%s]", params.Name)
	logDebug("Arguments: %v", params.Arguments)

	// Route to the specific handler
	switch params.Name {
	case "read_godot_file":
		return handleRead(params.Arguments["filepath"])
	case "write_godot_file":
		return handleWrite(params.Arguments["filepath"], params.Arguments["content"])
	case "open_godot":
		return handleOpen(params.Arguments["project_path"])
	case "kill_godot":
		return handleKill()
	case "list_files":
		return handleList()
	case "get_scene_template":
		return handleTemplate(params.Arguments["node_type"], params.Arguments["node_name"])
	case "create_directory":
		return handleMkdir(params.Arguments["directory_path"])
	default:
		return mcpError("Unknown tool: " + params.Name)
	}
}

// --- Individual Handlers (Keeping things clean) ---

func handleRead(path string) any {
	full := filepath.Join(projectPath, filepath.Clean(path))
	data, err := os.ReadFile(full)
	if err != nil {
		return mcpError(err.Error())
	}
	return mcpText(string(data))
}

func handleWrite(path, content string) any {
	full := filepath.Join(projectPath, filepath.Clean(path))
	// Ensure folder exists before writing
	os.MkdirAll(filepath.Dir(full), os.ModePerm)
	err := os.WriteFile(full, []byte(content), 0644)
	if err != nil {
		return mcpError(err.Error())
	}
	return mcpText("Successfully wrote to " + path)
}

func handleOpen(customPath string) any {
	target := customPath
	if target == "" {
		target = projectPath
	}
	cmd := exec.Command(godotExe, "--path", target, "-e")
	detachProcess(cmd) // Found in utils.go
	if err := cmd.Start(); err != nil {
		return mcpError(err.Error())
	}
	return mcpText("Godot launched for: " + target)
}

func handleKill() any {
	logInfo("Closing Godot instances...")
	// Force close via Windows taskkill
	_ = exec.Command("taskkill", "/F", "/IM", "Godot*", "/T").Run()
	return mcpText("Godot editor instances closed.")
}

func handleList() any {
	var files []string
	filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !strings.Contains(path, ".godot") {
			rel, _ := filepath.Rel(projectPath, path)
			files = append(files, rel)
		}
		return nil
	})
	return mcpText(strings.Join(files, "\n"))
}

func handleTemplate(nodeType, nodeName string) any {
	// Generate a valid TSCN boilerplate with proper syntax
	template := fmt.Sprintf(`[gd_scene format=3]

[node name="%s" type="%s"]
`, nodeName, nodeType)
	return mcpText(template)
}

func handleMkdir(path string) any {
	full := filepath.Join(projectPath, path)
	os.MkdirAll(full, os.ModePerm)
	return mcpText("Created directory: " + path)
}
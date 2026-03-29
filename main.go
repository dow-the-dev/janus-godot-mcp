package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"syscall"
)

var (
	godotExe    string
	projectPath string
	logLevel    int // 0: Error, 1: Info, 2: Debug
)

type RPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type RPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

// --- Logging Helpers ---

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

// --- Main Loop ---

func main() {
	godotExe = os.Getenv("GODOT_EXE")
	projectPath = os.Getenv("GODOT_PROJECT")
	setLogLevel(os.Getenv("LOG_LEVEL"))

	logInfo("Starting Slim Godot MCP v0.1.3")
	logDebug("Environment - GODOT_PROJECT: %s", projectPath)
	logDebug("Environment - GODOT_EXE: %s", godotExe)
	logDebug("Environment - LOG_LEVEL: %d", logLevel)

// Check for Godot Executable (This is still pretty critical)
    if godotExe == "" {
        logError("GODOT_EXE is not set. The 'open_godot' tool will fail until configured.")
    }

    // Check for Project Path (Now a warning, not a crash)
    if projectPath == "" {
        logWarning("GODOT_PROJECT is not set. Project-specific tools will use relative paths or fail.")
    }

	scanner := bufio.NewScanner(os.Stdin)
	// Increase scanner buffer size just in case AI sends massive file writes
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		var req RPCRequest
		raw := scanner.Bytes()
		
		logDebug("<<< RAW INCOMING: %s", string(raw))
		
		if err := json.Unmarshal(raw, &req); err != nil {
			logError("Failed to parse incoming JSON: %v", err)
			continue
		}
		
		logInfo("Processing Request - Method: %s | ID: %v", req.Method, req.ID)
		handleRequest(req)
	}
	
	if err := scanner.Err(); err != nil {
		logError("Scanner fatal error: %v", err)
	}
}

func handleRequest(req RPCRequest) {
	var result any

	switch req.Method {
	case "initialize":
        logDebug("Handling 'initialize' handshake...")
        result = map[string]any{
            "protocolVersion": "2024-11-05",
            "serverInfo":      map[string]string{"name": "slim-godot-mcp", "version": "0.1.3"},
            "capabilities": map[string]any{
                "tools": map[string]any{}, // This tells the client we HAVE tools
            },
        }
	case "tools/list":
        logDebug("Handling 'tools/list' request...")
        // Use []any to be more flexible with the different map structures
        tools := []any{
            map[string]any{
                "name":        "read_godot_file",
                "description": "Read a .gd or .tscn file from the Godot project.",
                "inputSchema": map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "filepath": map[string]string{"type": "string", "description": "Path relative to the project root"},
                    },
                    "required": []string{"filepath"},
                },
            },
            map[string]any{
                "name":        "write_godot_file",
                "description": "Creates or updates scripts (.gd) and scenes (.tscn).",
                "inputSchema": map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "filepath": map[string]string{"type": "string"},
                        "content":  map[string]string{"type": "string"},
                    },
                    "required": []string{"filepath", "content"},
                },
            },
            map[string]any{
                "name":        "open_godot",
                "description": "Opens Godot.",
                "inputSchema": map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "project_path": map[string]string{"type": "string"},
                    },
                },
            },
            map[string]any{
                "name":        "init_godot_project",
                "description": "Initializes a project.godot file.",
                "inputSchema": map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "project_name": map[string]string{"type": "string"},
                    },
                },
            },
            map[string]any{
                "name":        "list_files",
                "description": "Lists all files in the project.",
                "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
            },
            map[string]any{
                "name":        "get_project_info",
                "description": "Reads project.godot settings.",
                "inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
            },
            map[string]any{
                "name":        "create_directory",
                "description": "Creates a directory path.",
                "inputSchema": map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "directory_path": map[string]string{"type": "string"},
                    },
                    "required": []string{"directory_path"},
                },
            },
			map[string]any{
				"name": "get_scene_template",
				"description": "Generates a boilerplate TSCN for ANY Godot node type.",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"node_type": map[string]string{"type": "string", "description": "e.g., CharacterBody2D, Sprite2D"},
						"node_name": map[string]string{"type": "string", "description": "The name for the root node"},
					},
					"required": []string{"node_type", "node_name"},
				},
			},
        }
        result = map[string]any{"tools": tools}
	case "prompts/list":
		logDebug("Handling 'prompts/list' request...")
		result = map[string]any{"prompts": []any{}}
		
	case "resources/list":
		logDebug("Handling 'resources/list' request...")
		result = map[string]any{"resources": []any{}}
		
	case "resources/templates/list":
		logDebug("Handling 'resources/templates/list' request...")
		result = map[string]any{"templates": []any{}}
		
	case "notifications/cancelled":
		logDebug("Client cancelled a request. Ignoring.")
		// Notifications do not require a response, so we explicitly return early
		return
	case "tools/call":
		result = executeTool(req.Params)
	}

	if result != nil {
		out, _ := json.Marshal(RPCResponse{JSONRPC: "2.0", ID: req.ID, Result: result})
		logDebug(">>> RAW OUTGOING: %s", string(out))
		fmt.Println(string(out))
	} else {
		logDebug("No result generated for ID: %v (Method: %s)", req.ID, req.Method)
	}
}

func executeTool(paramsRaw json.RawMessage) any {
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	
	if err := json.Unmarshal(paramsRaw, &params); err != nil {
		logError("Failed to unmarshal tool params. Raw: %s", string(paramsRaw))
		return mcpError("Failed to parse tool arguments: " + err.Error())
	}
	if progress, ok := params.Arguments["task_progress"]; ok {
        logInfo("AI Progress Update: \n%s", progress)
        // We leave it in the map so we don't crash, 
        // but we can now ignore it for our logic.
    }
	logInfo("Tool Executing: [%s]", params.Name)
	logDebug("Tool Arguments Parsed: %v", params.Arguments)

	safePath := ""
	if pathArg, exists := params.Arguments["filepath"]; exists {
		safePath = filepath.Clean(pathArg)
		if strings.Contains(safePath, "..") {
			logError("Security block: Directory traversal attempted with path: %s", safePath)
			return mcpError("Directory traversal not allowed.")
		}
	}
	
	fullPath := filepath.Join(projectPath, safePath)

	switch params.Name {
	case "read_godot_file":
		logDebug("Target full path for read: %s", fullPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			logError("Read failed: %v", err)
			return mcpError(err.Error())
		}
		logDebug("Successfully read %d bytes from %s", len(data), safePath)
		return mcpText(string(data))

	case "write_godot_file":
		logDebug("Target full path for write: %s", fullPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
			logError("MkdirAll failed for %s: %v", filepath.Dir(fullPath), err)
			return mcpError("Failed to create directories: " + err.Error())
		}
		
		contentBytes := []byte(params.Arguments["content"])
		logDebug("Writing %d bytes...", len(contentBytes))
		
		if err := os.WriteFile(fullPath, contentBytes, 0644); err != nil {
			logError("WriteFile failed: %v", err)
			return mcpError(err.Error())
		}
		logInfo("Successfully wrote file to %s", safePath)
		return mcpText("Successfully wrote file to " + safePath)
		
	case "open_godot":
        // 1. Determine the target path
        target := ""
        if val, ok := params.Arguments["project_path"]; ok && val != "" {
            target = val
        } else {
            // Default to our env projectPath if the AI didn't specify one
            target = projectPath 
        }

        var cmd *exec.Cmd
        if target == "" {
            logInfo("LAUNCH MODE: Project Manager (No path provided)")
            cmd = exec.Command(godotExe)
        } else {
            logInfo("LAUNCH MODE: Specific Project -> %s", target)
            cmd = exec.Command(godotExe, "--path", target, "-e")
        }

        // 2. Detach but keep logs flowing to Janus for 5 seconds
        stderr, _ := cmd.StderrPipe()
        
        if err := cmd.Start(); err != nil {
            logError("OS EXECUTION ERROR: %v", err)
            return mcpError(err.Error())
        }

        // 3. Background monitor for immediate crashes
        go func() {
            scanner := bufio.NewScanner(stderr)
            for scanner.Scan() {
                logError("GODOT-CRASH-LOG: %s", scanner.Text())
            }
        }()

        logInfo("Godot launched (PID %d). Standing by...", cmd.Process.Pid)
        return mcpText(fmt.Sprintf("Godot launched for: %s", target))
		
	case "init_godot_project":
        // LOGGING: See exactly what Cline sent us
        logDebug("Raw Arguments received: %v", params.Arguments)

        projectName := "Janus Project"
        // Only use the project_name if it's actually provided and not a progress string
        if val, ok := params.Arguments["project_name"]; ok && val != "" && !strings.HasPrefix(val, "- [") {
            projectName = val
        }

        projectFile := filepath.Join(projectPath, "project.godot")

        // IMPROVED LOGGING: Tell the AI EXACTLY which path we are looking at
        if _, err := os.Stat(projectFile); err == nil {
            logInfo("Init skipped: %s already exists", projectFile)
            return mcpText(fmt.Sprintf("SUCCESS: Project already exists at %s. You can proceed to 'open_godot' now.", projectPath))
        }

        // We need to define the file content here since it was missing in the previous snippet
        content := fmt.Sprintf(`; Engine configuration file.
; It's best edited using the editor UI and not directly.

config_version=5

[application]

config/name="%s"
`, projectName)

        logDebug("Ensuring directory exists: %s", projectPath)
        if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
            logError("MkdirAll failed: %v", err)
            return mcpError("Failed to create directories: " + err.Error())
        }
        
        logDebug("Writing project.godot file...")
        if err := os.WriteFile(projectFile, []byte(content), 0644); err != nil {
            logError("WriteFile failed for project.godot: %v", err)
            return mcpError(err.Error())
        }
        
        logInfo("Initialized new Godot project: %s", projectName)
        return mcpText(fmt.Sprintf("SUCCESS: Created new project '%s' at %s. You MUST now call 'open_godot' to see it.", projectName, projectPath))

	case "list_files":
        var files []string
        filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
            if err == nil && !info.IsDir() && !strings.Contains(path, ".godot") {
                rel, _ := filepath.Rel(projectPath, path)
                files = append(files, rel)
            }
            return nil
        })
        return mcpText(strings.Join(files, "\n"))

    case "get_project_info":
        data, err := os.ReadFile(filepath.Join(projectPath, "project.godot"))
        if err != nil { return mcpError("No project.godot found. Run init_godot_project first.") }
        return mcpText(string(data))

	case "create_directory":
        dirPath := filepath.Join(projectPath, params.Arguments["directory_path"])
        if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
            return mcpError(err.Error())
        }
        return mcpText("Created directory: " + params.Arguments["directory_path"])

	case "get_scene_template":
		nodeType := params.Arguments["node_type"]
		nodeName := params.Arguments["node_name"]
		
		// The Universal Godot 4 TSCN Boilerplate
		template := fmt.Sprintf(
			"[gd_scene format=3]\n\n[node name=\"%s\" type=\"%s\"]\n", 
			nodeName, 
			nodeType,
		)
		return mcpText(template)

	default:
		logError("Unknown tool requested: %s", params.Name)
		return mcpError("Unknown tool: " + params.Name)
	}
}
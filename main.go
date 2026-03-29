package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	godotExe = os.Getenv("GODOT_EXE")
	projectPath = os.Getenv("GODOT_PROJECT")
	setLogLevel(os.Getenv("LOG_LEVEL"))

	logInfo("Starting Slim Godot MCP v0.1.3")
	logDebug("Environment - GODOT_PROJECT: %s", projectPath)
	logDebug("Environment - GODOT_EXE: %s", godotExe)

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
				"name":        "get_scene_template",
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
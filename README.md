# Project Janus: Godot MCP Server

A lightweight, Go-based Model Context Protocol (MCP) server designed to act as a bridge between an AI assistant (like Cline) and the Godot Engine.

## Building the Server
To compile the server into an executable, run this in the project directory:
```bash
go build -o janus-mcp.exe .
```

## Launch janus manually
```sh
export GODOT_EXE="/c/Program Files/Godot/Godot_v4.5-dev4_win64.exe" && \
export GODOT_PROJECT="/c/Users/drewm/Desktop/Desktop/GodotProjects/DummyPong" && \
LOG_LEVEL="DEBUG" && \
./janus-mcp.exe
```
### Launching JUST godot
```sh
export GODOT_EXE="/c/Program Files/Godot/Godot_v4.5-dev4_win64.exe" && \
LOG_LEVEL="DEBUG" && \
./janus-mcp.exe
```

# Tests:

### Initialize connection
```json
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}
```

### List Available Tools
```json
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}
```

### Initialize a New Project
```json
{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}
```

### Open Godot
```json
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "open_godot", "arguments": {}}}
```

### Open Godot to a specific project
```json
{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": {"name": "open_godot", "arguments": { "project_path": "C:\\Users\\drewm\\Desktop\\Desktop\\GodotProjects/DummyPong" }}}
```

### Write a script
```json
{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": {"name": "write_godot_file", "arguments": {"filepath": "scripts/mcp_test.gd", "content": "extends Node\n\nfunc _ready():\n\tpass\n"}}}
```

### Read a script
```json
{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": {"name": "read_godot_file", "arguments": {"filepath": "scripts/mcp_test.gd"}}}
```

# Example C:\Users\drewm\AppData\Roaming\Code\User\mcp.json
```json
{
	"servers": {
		"janus-godot-mcp": {
			"type": "stdio",
			"command": "C:/Users/drewm/Desktop/Desktop/GodotProjects/janus-godot-mcp/janus-mcp.exe",
			"args": [],
			"env": {
				"GODOT_EXE": "C:/Program Files/Godot/Godot_v4.5-dev4_win64.exe",
				"GODOT_PROJECT": "C:/Users/drewm/Desktop/Desktop/GodotProjects/DummyPong",
				"LOG_LEVEL": "DEBUG"
			}
		}
	},
	"inputs": []
}
```
# Code Agent

A Go-based CLI tool for interacting with Anthropic's Claude AI assistant. This tool provides an interactive chat interface for code generation, debugging, and programming assistance with powerful file system tools.

## Project Summary

Code Agent is a Go-based CLI tool that enables seamless interaction with Anthropic's Claude AI assistant for programming-related tasks. The application provides a chat interface where users can communicate with Claude while allowing the AI to access and manipulate local files and directories through defined tools.

This project is based on the tutorial from [https://ampcode.com/how-to-build-an-agent](https://ampcode.com/how-to-build-an-agent).

### Key Features:
- Interactive CLI chat interface with Claude AI
- Persistence of conversation context across exchanges
- **Powerful tool execution capabilities** (read_file, list_files, edit_file)
- Secure API key management
- Colored terminal output for better user experience
- **File system integration** - Claude can read, list, and edit files directly

### Technical Implementation:
- Written in Go with a clean, modular structure
- Uses Anthropic's official SDK for API communication
- Implements a tool execution framework where Claude can request to access files
- Handles the message flow including tool use and results properly
- **Tool-based architecture** allowing easy extension with new capabilities

### Relation to State of the Art:

This project represents an evolution in AI coding assistants by:

1. **Tool Integration**: Unlike basic chat interfaces, Code Agent implements the tool-use capabilities of Claude 3.7 Sonnet, allowing the AI to interact with the user's file system. This bridges the gap between AI assistants and traditional development environments.

2. **Local-First Design**: Where many AI coding assistants rely on browser interfaces or cloud environments, Code Agent works directly in the user's terminal with access to local files, maintaining developer workflow and privacy.

3. **Simplified API Interaction**: The tool abstracts away the complexity of the Anthropic API, handling authentication, conversation management, and tool execution protocols, making advanced AI capabilities accessible through a simple CLI.

4. **Lightweight Approach**: Unlike heavier IDE extensions or plugins, this tool provides AI assistance through a minimal interface that integrates with existing development workflows rather than replacing them.

Code Agent sits at the intersection of traditional command-line tools and modern AI assistants, representing a pragmatic approach to incorporating AI into software development workflows. Its architecture demonstrates the potential for AI agents to operate with controlled access to local resources while maintaining security boundaries.

## Features

- 🤖 Interactive chat with Claude AI
- 💬 Conversation memory and context
- 🎨 Colored output for better UX
- 🔒 Secure API key management
- 🚀 Simple CLI interface
- 📁 **File system tools** - Read, list, and edit files
- 🔧 **Tool execution framework** - Claude can use tools automatically

## Available Tools

Code Agent provides Claude with several powerful tools to interact with your file system:

### 📖 `read_file` - Read File Contents
**Description**: Read the contents of a file in the working directory.

**Usage**: Claude can automatically use this tool when you ask about file contents.

**Example conversation**:
```
You: What's in main.go?
Claude: I'll check the contents of main.go for you.
tool: read_file({"path":"main.go"})
Claude: Here's what's in main.go:
[file contents displayed]
```

### 📋 `list_files` - List Directory Contents
**Description**: List files and directories at a given path. If no path is provided, lists files in the current directory.

**Usage**: Claude uses this to explore your project structure.

**Example conversation**:
```
You: What files are in this project?
Claude: Let me check what files are available in the current directory.
tool: list_files({"path":"."})
Claude: Here are the files in your project:
- main.go
- go.mod
- go.sum
- README.md
- config.env.example
```

### ✏️ `edit_file` - Edit File Contents
**Description**: Make edits to text files by replacing specific text. Can also create new files.

**Usage**: Claude can modify files directly based on your requests.

**Example conversation**:
```
You: Add a comment to the top of main.go explaining what it does
Claude: I'll add a comment to the top of main.go for you.
tool: edit_file({"path":"main.go","old_str":"package main","new_str":"// This is a CLI tool for chatting with Claude AI\npackage main"})
Claude: I've added the comment to main.go successfully.
```

**Parameters**:
- `path`: The file path to edit
- `old_str`: Text to search for (must match exactly)
- `new_str`: Text to replace it with
- If `old_str` is empty and the file doesn't exist, creates a new file with `new_str` content

## Prerequisites

- Go 1.19 or higher
- Anthropic API key

## Installation

1. Clone the repository:
```bash
git clone <your-repo-url>
cd code-agent
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up your API key:
```bash
# Copy the example config file
cp config.env.example config.env

# Edit config.env and add your actual API key
# Get your key from: https://console.anthropic.com/
```

## Usage

### Interactive Mode
```bash
go run main.go
```

### Build and Run
```bash
go build -o code-agent
./code-agent
```

### Example Workflows

**Code Review**:
```
You: Review the code in main.go and suggest improvements
Claude: I'll read the main.go file and provide a code review.
[Claude reads the file and provides detailed feedback]
```

**File Organization**:
```
You: List all Go files in the project and suggest a better structure
Claude: Let me explore your project structure first.
[Claude lists files and suggests improvements]
```

**Code Generation**:
```
You: Create a new file called utils.go with helper functions
Claude: I'll create a new utils.go file with some common helper functions.
[Claude creates the file with appropriate content]
```

## Configuration

The application reads your API key from either:
1. `ANTHROPIC_API_KEY` environment variable
2. `config.env` file

**Important**: Never commit your actual API key to version control!

## Features

- **Conversation Memory**: Claude remembers previous messages in the session
- **Error Handling**: Graceful handling of API overloads and network issues
- **Colored Output**: Blue for user messages, yellow for Claude responses, green for tool usage
- **Graceful Exit**: Use Ctrl+C or Ctrl+D to exit
- **Tool Integration**: Claude automatically uses appropriate tools when needed
- **File Safety**: Tools operate on relative paths within your working directory

## Development

### Project Structure
```
code-agent/
├── main.go           # Main application code with tool implementations
├── go.mod           # Go module file
├── go.sum           # Dependency checksums
├── config.env       # API key (not in git)
├── config.env.example # Example config
├── .gitignore       # Git ignore rules
└── README.md        # This file
```

### Adding New Tools

To add a new tool, follow this pattern:

1. Define the tool input structure:
```go
type MyToolInput struct {
    Param1 string `json:"param1" jsonschema_description:"Description of param1"`
    Param2 int    `json:"param2" jsonschema_description:"Description of param2"`
}
```

2. Create the tool function:
```go
func MyTool(input json.RawMessage) (string, error) {
    var myInput MyToolInput
    err := json.Unmarshal(input, &myInput)
    if err != nil {
        return "", fmt.Errorf("invalid input: %w", err)
    }
    
    // Tool logic here
    return "result", nil
}
```

3. Define the tool:
```go
var MyToolDefinition = ToolDefinition{
    Name:        "my_tool",
    Description: "Description of what this tool does",
    InputSchema: GenerateSchema[MyToolInput](),
    Function:    MyTool,
}
```

4. Add to the tools list in main():
```go
tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition, MyToolDefinition}
```

### Building
```bash
go build -o code-agent
```

### Testing
```bash
go test ./...
```

## Security Considerations

- **API keys** are stored locally in `config.env`
- The `.gitignore` file prevents accidental commits of sensitive data
- **Environment variables** are used for secure key management
- **Tool access** is limited to the working directory and subdirectories
- **File operations** use relative paths to prevent access to system files

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Troubleshooting

**"tool not found" error**: This means Claude tried to use a tool that isn't implemented. Check that all tools are properly added to the tools list in main().

**File permission errors**: Ensure the application has read/write permissions in the working directory.

**API key errors**: Verify your API key is correctly set in either the environment variable or config.env file. 
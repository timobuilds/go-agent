/*
Code Agent - A CLI tool for interacting with Claude AI with tool support

This file implements a command-line interface that allows users to chat with Claude AI
while providing access to various tools (like file reading). The agent can:

1. Load API credentials from environment variables or config files
2. Maintain conversation context across multiple exchanges
3. Execute tools requested by Claude (e.g., read_file)
4. Handle the tool use/result flow properly with the Anthropic API

Key components:
- Agent: Main conversation handler with tool execution capabilities
- ToolDefinition: Interface for defining tools that Claude can use
- Tool implementations: Concrete tools like read_file

Usage:
1. Set ANTHROPIC_API_KEY environment variable or create config.env file
2. Run: go run main.go
3. Chat with Claude - it can use available tools automatically
*/

package main

import (
	"bufio"   // For reading input line by line
	"context" // For context management and cancellation
	"encoding/json"
	"fmt" // For formatted output
	"os"  // For accessing stdin and environment variables
	"path"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go" // Anthropic's official Go SDK for Claude API
	"github.com/invopop/jsonschema"
)

// =============================================================================
// MAIN ENTRY POINT
// =============================================================================

func main() {
	// Initialize API client with credentials
	client, err := initializeClient()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	// Set up user input handling
	scanner := bufio.NewScanner(os.Stdin)
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	// Define available tools
	tools := []ToolDefinition{ReadFileDefinition, ListFilesDefinition, EditFileDefinition}

	// Create and run the agent
	agent := NewAgent(client, getUserMessage, tools)
	err = agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

// =============================================================================
// CLIENT INITIALIZATION
// =============================================================================

// initializeClient sets up the Anthropic API client with proper authentication
func initializeClient() (*anthropic.Client, error) {
	// Load API key from environment or config file
	apiKey := loadAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is required")
	}

	// Debug: Show partial key for verification
	fmt.Printf("API Key loaded: %s...\n", apiKey[:20])

	// Set environment variable for the client
	os.Setenv("ANTHROPIC_API_KEY", apiKey)

	// Create and return the client
	client := anthropic.NewClient()
	return &client, nil
}

// loadAPIKey attempts to load the API key from environment or config file
func loadAPIKey() string {
	// Try environment variable first
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey != "" {
		return apiKey
	}

	// Try config file as fallback
	if data, err := os.ReadFile("config.env"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "ANTHROPIC_API_KEY=") {
				return strings.TrimPrefix(line, "ANTHROPIC_API_KEY=")
			}
		}
	}

	// No key found
	fmt.Println("Error: ANTHROPIC_API_KEY is required")
	fmt.Println("Please either:")
	fmt.Println("1. Set environment variable: export ANTHROPIC_API_KEY=your_api_key_here")
	fmt.Println("2. Add your key to config.env file")
	return ""
}

// =============================================================================
// AGENT CORE STRUCTURE
// =============================================================================

// Agent represents the main conversation handler with tool execution capabilities
type Agent struct {
	client         *anthropic.Client     // Client for making API calls to Claude
	getUserMessage func() (string, bool) // Function to get user input
	tools          []ToolDefinition      // List of available tools
}

// NewAgent creates a new agent instance with the specified client and tools
func NewAgent(
	client *anthropic.Client,
	getUserMessage func() (string, bool),
	tools []ToolDefinition,
) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
	}
}

// =============================================================================
// CONVERSATION MANAGEMENT
// =============================================================================

// Run starts the main conversation loop and handles the chat flow
func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	readUserInput := true

	// Main conversation loop
	for {
		if readUserInput {
			// Get user input and add to conversation
			fmt.Print("\u001b[94mYou\u001b[0m: ")
			readUserInput = false

			userInput, ok := a.getUserMessage()
			if !ok {
				break
			}

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))
			conversation = append(conversation, userMessage)
		}

		// Get Claude's response
		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}

		// Add Claude's response to conversation history
		conversation = append(conversation, message.ToParam())

		// Process Claude's response for tool usage
		toolResults := a.processClaudeResponse(message)

		// Handle tool results if any
		if len(toolResults) > 0 {
			// Send tool results back to Claude as a user message
			toolResultMessage := anthropic.NewUserMessage(toolResults...)
			conversation = append(conversation, toolResultMessage)
			readUserInput = false
		} else {
			readUserInput = true
		}
	}

	return nil
}

// processClaudeResponse handles Claude's response and executes any requested tools
func (a *Agent) processClaudeResponse(message *anthropic.Message) []anthropic.ContentBlockParamUnion {
	toolResults := []anthropic.ContentBlockParamUnion{}

	for _, content := range message.Content {
		switch content.Type {
		case "text":
			fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
		case "tool_use":
			result := a.executeTool(content.ID, content.Name, content.Input)
			toolResults = append(toolResults, result)
		}
	}

	return toolResults
}

// =============================================================================
// API COMMUNICATION
// =============================================================================

// runInference sends the conversation to Claude and returns the response
func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	// Convert tool definitions to Anthropic's format
	anthropicTools := a.convertToolsToAnthropicFormat()

	// Make API call to Claude
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: int64(1024),
		Messages:  conversation,
		Tools:     anthropicTools,
	})

	return message, err
}

// convertToolsToAnthropicFormat converts our tool definitions to Anthropic's format
func (a *Agent) convertToolsToAnthropicFormat() []anthropic.ToolUnionParam {
	anthropicTools := []anthropic.ToolUnionParam{}

	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	return anthropicTools
}

// =============================================================================
// TOOL EXECUTION
// =============================================================================

// executeTool finds and executes the requested tool
func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	// Find the tool definition
	var toolDef ToolDefinition
	var found bool
	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}

	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	// Execute the tool
	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s)\n", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}

	return anthropic.NewToolResultBlock(id, response, false)
}

// =============================================================================
// TOOL DEFINITIONS
// =============================================================================

// ToolDefinition represents a tool that Claude can use
type ToolDefinition struct {
	Name        string                                      `json:"name"`
	Description string                                      `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam              `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error) `json:"-"`
}

// =============================================================================
// READ FILE TOOL IMPLEMENTATION
// =============================================================================

// ReadFileDefinition - Tool that allows Claude to read files
var ReadFileDefinition = ToolDefinition{
	Name:        "read_file",
	Description: "Read the contents of a given relative file path. Use this when you want to see what's inside a file. Do not use this with directory names.",
	InputSchema: ReadFileInputSchema,
	Function:    ReadFile,
}

// ReadFileInput defines the input structure for the read_file tool
type ReadFileInput struct {
	Path string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
}

// ReadFileInputSchema - Auto-generated JSON schema for ReadFileInput
var ReadFileInputSchema = GenerateSchema[ReadFileInput]()

// ReadFile executes the file reading functionality
func ReadFile(input json.RawMessage) (string, error) {
	// Parse the input
	readFileInput := ReadFileInput{}
	err := json.Unmarshal(input, &readFileInput)
	if err != nil {
		return "", fmt.Errorf("invalid input format: %w", err)
	}

	// Read the file
	content, err := os.ReadFile(readFileInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", readFileInput.Path, err)
	}

	return string(content), nil
}

// =============================================================================
// LIST FILE TOOL IMPLEMENTATION
// =============================================================================

// ListFileDefinition - Tool that allows Claude to list files in the working directory
var ListFilesDefinition = ToolDefinition{
	Name:        "list_files",
	Description: "List files and directories at a given path. If no path is provided, lists files in the current directory.",
	InputSchema: ListFilesInputSchema,
	Function:    ListFiles,
}

type ListFilesInput struct {
	Path string `json:"path,omitempty" jsonschema_description:"Optional relative path to list files from. Defaults to current directory if not provided."`
}

var ListFilesInputSchema = GenerateSchema[ListFilesInput]()

func ListFiles(input json.RawMessage) (string, error) {
	listFilesInput := ListFilesInput{}
	err := json.Unmarshal(input, &listFilesInput)
	if err != nil {
		panic(err)
	}

	dir := "."
	if listFilesInput.Path != "" {
		dir = listFilesInput.Path
	}

	var files []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if relPath != "." {
			if info.IsDir() {
				files = append(files, relPath+"/")
			} else {
				files = append(files, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	result, err := json.Marshal(files)
	if err != nil {
		return "", err
	}

	return string(result), nil
}

// =============================================================================
// EDIT FILE TOOL IMPLEMENTATION
// =============================================================================
var EditFileDefinition = ToolDefinition{
	Name: "edit_file",
	Description: `Make edits to a text file.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file specified with path doesn't exist, it will be created.
`,
	InputSchema: EditFileInputSchema,
	Function:    EditFile,
}

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var EditFileInputSchema = GenerateSchema[EditFileInput]()

func EditFile(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	err := json.Unmarshal(input, &editFileInput)
	if err != nil {
		return "", err
	}

	if editFileInput.Path == "" || editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			return createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", err
	}

	oldContent := string(content)
	newContent := strings.Replace(oldContent, editFileInput.OldStr, editFileInput.NewStr, -1)

	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	err = os.WriteFile(editFileInput.Path, []byte(newContent), 0644)
	if err != nil {
		return "", err
	}

	return "OK", nil
}

func createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================

// GenerateSchema creates a JSON schema for a given type using reflection
func GenerateSchema[T any]() anthropic.ToolInputSchemaParam {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T

	schema := reflector.Reflect(v)

	return anthropic.ToolInputSchemaParam{
		Properties: schema.Properties,
	}
}

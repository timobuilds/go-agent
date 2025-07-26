package main

import (
	"bufio"   // For reading input line by line
	"context" // For context management and cancellation
	"encoding/json"
	"fmt" // For formatted output
	"os"  // For accessing stdin and environment variables
	"strings"

	"github.com/anthropics/anthropic-sdk-go" // Anthropic's official Go SDK for Claude API
)

func main() {
	// Get API key from environment variable or config file
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		// Try to read from config file
		if data, err := os.ReadFile("config.env"); err == nil {
			// Simple parsing - look for ANTHROPIC_API_KEY=value
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "ANTHROPIC_API_KEY=") {
					apiKey = strings.TrimPrefix(line, "ANTHROPIC_API_KEY=")
					break
				}
			}
		}
	}

	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY is required")
		fmt.Println("Please either:")
		fmt.Println("1. Set environment variable: export ANTHROPIC_API_KEY=your_api_key_here")
		fmt.Println("2. Add your key to config.env file")
		os.Exit(1)
	}

	// Debug: Check if API key is being read correctly
	fmt.Printf("API Key loaded: %s...\n", apiKey[:20])

	// Set the API key in environment for the client to use
	os.Setenv("ANTHROPIC_API_KEY", apiKey)

	// Create a new Anthropic client instance
	// The client will automatically use the ANTHROPIC_API_KEY environment variable
	client := anthropic.NewClient()

	// Set up a scanner to read from standard input (keyboard)
	// This allows us to read user input line by line
	scanner := bufio.NewScanner(os.Stdin)

	// Define a function that gets the next user message
	// Returns the message text and a boolean indicating if input was successful
	getUserMessage := func() (string, bool) {
		if !scanner.Scan() { // Try to read the next line
			return "", false // Return empty string and false if no more input (EOF)
		}
		return scanner.Text(), true // Return the line text and true for success
	}

	// Create a new agent instance with the client and input function
	agent := NewAgent(&client, getUserMessage)

	// Start the agent and run it until completion or error
	err := agent.Run(context.TODO())
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}

// NewAgent creates and returns a new Agent instance
// Parameters:
//   - client: Pointer to the Anthropic client for API calls
//   - getUserMessage: Function that returns the next user input
func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool)) *Agent {
	return &Agent{
		client:         client,         // Store the API client
		getUserMessage: getUserMessage, // Store the input function
	}
}

// Agent represents our code generation assistant
// It holds the API client and input handling function
type Agent struct {
	client         *anthropic.Client     // Client for making API calls to Claude
	getUserMessage func() (string, bool) // Function to get user input
	tools          []ToolDefinition      // List of available tools
}

func (a *Agent) Run(ctx context.Context) error {
	// Initialize an empty conversation history
	// This will store all messages exchanged between user and Claude
	conversation := []anthropic.MessageParam{}

	// Display welcome message to the user
	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")

	// Main conversation loop - runs until user exits or error occurs
	for {
		// Display user prompt with blue color (\u001b[94m = blue, \u001b[0m = reset)
		fmt.Print("\u001b[94mYou\u001b[0m: ")

		// Get the next user input from stdin
		userInput, ok := a.getUserMessage()
		if !ok {
			// If getUserMessage returns false, it means EOF (end of input)
			// This happens when user presses Ctrl+D or input is redirected
			break
		}

		// Create a new user message for the conversation
		// NewTextBlock wraps the user input in the proper format for Claude
		userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(userInput))

		// Add the user message to our conversation history
		conversation = append(conversation, userMessage)

		// Send the conversation to Claude and get a response
		message, err := a.runInference(ctx, conversation)
		if err != nil {
			// Check for specific error types and provide helpful messages
			if strings.Contains(err.Error(), "overloaded_error") {
				fmt.Println("\n\u001b[93mClaude\u001b[0m: Sorry, I'm currently overloaded. Please try again in a moment.")
				continue
			}
			return err // Return other API errors
		}

		// Add Claude's response to the conversation history
		// This maintains context for future messages
		conversation = append(conversation, message.ToParam())

		// Display Claude's response
		// Iterate through all content blocks in the response
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				// Display text content with yellow color (\u001b[93m = yellow)
				fmt.Printf("\u001b[93mClaude\u001b[0m: %s\n", content.Text)
			}
		}
	}

	return nil
}

// runInference sends the conversation to Claude and returns the response
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - conversation: Array of all previous messages for context
//
// Returns:
//   - *anthropic.Message: Claude's response message
//   - error: Any API or network errors
func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	// Make API call to Claude with the conversation history
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest, // Use the latest Claude 3.5 Sonnet model
		MaxTokens: int64(1024),                          // Limit response to 1024 tokens
		Messages:  conversation,                         // Send full conversation for context
	})
	return message, err
}

type ToolDefinition struct {
	Name        string                         `json:"name"`
	Description string                         `json:"description"`
	InputSchema anthropic.ToolInputSchemaParam `json:"input_schema"`
	Function    func(input json.RawMessage) (string, error)
}

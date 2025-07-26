# Code Agent

A Go-based CLI tool for interacting with Anthropic's Claude AI assistant. This tool provides an interactive chat interface for code generation, debugging, and programming assistance.

## Features

- 🤖 Interactive chat with Claude AI
- 💬 Conversation memory and context
- 🎨 Colored output for better UX
- 🔒 Secure API key management
- 🚀 Simple CLI interface

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

## Configuration

The application reads your API key from either:
1. `ANTHROPIC_API_KEY` environment variable
2. `config.env` file

**Important**: Never commit your actual API key to version control!

## Features

- **Conversation Memory**: Claude remembers previous messages in the session
- **Error Handling**: Graceful handling of API overloads and network issues
- **Colored Output**: Blue for user messages, yellow for Claude responses
- **Graceful Exit**: Use Ctrl+C or Ctrl+D to exit

## Development

### Project Structure
```
code-agent/
├── main.go           # Main application code
├── go.mod           # Go module file
├── go.sum           # Dependency checksums
├── config.env       # API key (not in git)
├── config.env.example # Example config
├── .gitignore       # Git ignore rules
└── README.md        # This file
```

### Building
```bash
go build -o code-agent
```

### Testing
```bash
go test ./...
```

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## Security

- API keys are stored locally in `config.env`
- The `.gitignore` file prevents accidental commits of sensitive data
- Environment variables are used for secure key management 
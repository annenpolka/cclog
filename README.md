# cclog - Claude Conversation Log to Markdown Converter

<img width="1620" height="1052" alt="Image" src="https://github.com/user-attachments/assets/fd7c3707-1fb3-4831-847e-f2f2bf9c0b2d" />A Go command-line tool that parses Claude conversation logs (JSONL format) and converts them to human-readable Markdown format.

## Features

- **Interactive TUI Mode**: Browse and select conversation logs with an intuitive terminal interface
- **Markdown Output**: Converts conversations to clean, readable Markdown format

## Installation

```bash
# Install latest version from GitHub
go install github.com/annenpolka/cclog/cmd/cclog@latest

# Or build from source
go build -o cclog ./cmd/cclog/

# Run directly with Go
go run ./cmd/cclog/
```

## Usage

```
cclog [OPTIONS] [input]
```

### Arguments

- `[input]` - Path to JSONL file or directory containing JSONL files
  - If no input provided, opens interactive TUI mode with recursive search

### Options

- `-d, --directory` - Treat input as directory (parse all .jsonl files)
- `-o, --output FILE` - Write output to file instead of stdout
- `--include-all` - Include all messages (no filtering of empty/system messages)
- `--show-uuid` - Show UUID metadata for each message
- `--show-title` - Show conversation title as header
- `--tui` - Open interactive file picker (TUI mode)
- `-r, --recursive` - Recursively search for .jsonl files and open TUI mode
- `--path PATH` - Specify directory path for TUI mode
- `-h, --help` - Show help message

## Examples

### Interactive Mode (Default)
```bash
# Open interactive file picker with recursive search
cclog

# Open TUI in specific directory
cclog --path /path/to/logs

# Explicit TUI mode
cclog --tui
```

### Single File Processing
```bash
# Convert single file to stdout
cclog conversation.jsonl

# Convert single file to output file
cclog conversation.jsonl -o output.md

# Include all messages without filtering
cclog --include-all conversation.jsonl

# Show UUID metadata
cclog --show-uuid conversation.jsonl
```

### Directory Processing
```bash
# Convert all JSONL files in directory
cclog -d /path/to/logs -o combined.md

# Recursively find and process all JSONL files
cclog -r /path/to/logs
```

## Output Format

The tool converts Claude conversation logs into clean Markdown format with:

- **Timestamps**: Automatically converted to system timezone for readability
- **Message Filtering**: Removes system messages, API errors, and interrupted requests
- **Content Extraction**: Handles both simple and complex message structures
- **Readable Format**: Well-structured Markdown with proper formatting

## Architecture

The project follows Go's standard project layout with clear separation of concerns:

- **JSONL Parsing** (`internal/parser`) - Reads and parses conversation log files
- **Type System** (`pkg/types`) - Defines message structures and conversation logs
- **Message Filtering** (`internal/formatter/filter`) - Filters out noise and system messages
- **Markdown Formatting** (`internal/formatter/markdown`) - Converts parsed data to readable Markdown
- **CLI Interface** (`cmd/cclog` and `internal/cli`) - Provides command-line interface

## Development

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build the application
go build -o cclog ./cmd/cclog/
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

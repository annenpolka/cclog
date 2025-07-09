# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cclog is a Go command-line tool that parses Claude Code conversation logs (JSONL format) and converts them to human-readable Markdown. It features both a powerful interactive TUI mode and flexible CLI processing. The project follows test-driven development (TDD) practices with comprehensive test coverage.

## Development Commands

```bash
# Build commands
go build -o cclog ./cmd/cclog/
make build          # Alternative build command

# Testing commands
go test ./...                    # Run all tests
go test -cover ./...             # Run tests with coverage
go test -v ./...                 # Run tests with verbose output
go test ./pkg/types/             # Run tests for specific package
make test                        # Run all tests (via Makefile)
make test-coverage               # Run tests with coverage

# Code quality commands
make fmt             # Format code (go fmt ./...)
make vet             # Run linter (go vet ./...)
make deps            # Install and tidy dependencies

# Build variants
make build-all       # Build for multiple platforms (linux, darwin, windows)
make install         # Install binary to GOPATH/bin
make clean           # Clean build artifacts

# Run application
./cclog [arguments]
make run             # Build and run (starts TUI mode)
```

## Architecture

The codebase follows Go's standard project layout with layered architecture and clear separation of concerns:

### Core Data Flow
1. **JSONL Parsing** (`internal/parser`) - Reads and parses conversation log files
2. **Type System** (`pkg/types`) - Defines message structures and conversation logs
3. **Message Filtering** (`internal/formatter/filter`) - Filters out noise and system messages
4. **Markdown Formatting** (`internal/formatter/markdown`) - Converts parsed data to readable Markdown
5. **CLI Interface** (`cmd/cclog` and `internal/cli`) - Provides command-line interface and TUI orchestration
6. **TUI System** (`pkg/filepicker` and `pkg/terminal`) - Interactive file browser with live preview

### Key Components

- **Message Type System**: The `types.Message` struct handles complex JSONL structure from Claude conversations, including nested message content, timestamps, metadata, and title extraction
- **Parser Strategy**: Line-by-line JSONL parsing with buffer expansion (up to 1MB), proper error handling, and empty line skipping
- **Message Filtering**: Intelligent filtering that removes system messages, API errors, interrupted requests, command outputs, meta messages, and Bash inputs/outputs
- **Markdown Generation**: Time-sorted message processing with system timezone conversion and content extraction from Claude's complex message format
- **Content Extraction**: Handles both simple string content and complex array-based content structures from Claude's message format
- **CLI Features**: Supports single file/directory processing, output file specification, filtering options, UUID display, and TUI integration
- **TUI System**: Interactive file browser with live Markdown preview, conversation metadata display, clipboard integration, and `claude` CLI integration for conversation resumption

### TUI Architecture

The TUI system provides a rich interactive experience:
- **File Browser**: Recursive directory traversal with `.jsonl` file detection
- **Live Preview**: Real-time Markdown rendering with toggle functionality
- **Conversation Metadata**: Display of dates, project names, and extracted titles
- **Integration Features**: Session ID clipboard copy, conversation resumption via `claude` CLI, direct editor opening
- **State Management**: Uses Bubble Tea framework for robust TUI state handling

### TDD Approach

This project follows t-wada's TDD practices with the Red-Green-Refactor cycle:

1. **Red**: Write a failing test first
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Improve code quality while keeping tests green

Each package includes comprehensive test files following the `*_test.go` naming convention. Tests cover:
- Message unmarshaling and data integrity
- File and directory parsing with error cases
- Markdown formatting with various message types
- Content extraction from different message structures
- TUI interactions and state management

The test data in `testdata/sample.jsonl` represents actual Claude conversation log structure for realistic testing scenarios.

**TDD Development Flow:**
- Always write tests before implementation
- Run tests frequently during development (`go test ./...`)
- Ensure all tests pass before committing code
- Refactor only when tests are green

## CLI Usage Examples

```bash
# Interactive TUI mode (default when no arguments)
cclog

# TUI with specific directory (enables recursive search)
cclog --path /path/to/logs

# Convert single file to stdout
cclog conversation.jsonl

# Convert single file to output file
cclog conversation.jsonl -o output.md

# Convert all JSONL files in directory
cclog -d /path/to/logs -o combined.md

# Advanced filtering and metadata options
cclog --include-all --show-uuid --show-title conversation.jsonl

# Resume conversation from TUI or get session ID
# (Interactive: press 'r' in TUI, or 'c' to copy session ID)
```

## TUI Keybindings

| Key | Action |
|-----|--------|
| `↑`/`↓`/`j`/`k` | Navigate file list |
| `enter` | Enter directory or convert file and open in editor |
| `p` | Toggle live Markdown preview |
| `s` | Toggle message filtering |
| `c` | Copy session ID to clipboard |
| `r` | Resume conversation with `claude` CLI |
| `R` | Resume conversation with `--dangerously-skip-permissions` |
| `q`/`ctrl+c` | Quit application |

## Dependencies

The project uses both standard library and TUI-focused dependencies:

**Core Logic**: Uses only Go standard library for parsing and formatting
**TUI Components**: 
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `golang.design/x/clipboard` - Clipboard integration
- `golang.org/x/term` - Terminal handling

## Important Notes

- **Auto-detection**: Automatically detects Claude projects directory (`~/.claude/projects`) when no input provided
- **Timezone handling**: Converts timestamps to system local timezone for readability
- **Message filtering**: Intelligent filtering enabled by default, removes system noise while preserving meaningful conversation content
- **TUI as default**: When no arguments provided, starts in interactive TUI mode for better user experience
- **Integration**: Seamless integration with `claude` CLI for conversation resumption
- **Test data**: The `testdata/sample.jsonl` file contains realistic Claude conversation structure for testing
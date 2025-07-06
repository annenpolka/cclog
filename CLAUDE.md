# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cclog is a Go command-line tool that parses Claude conversation logs (JSONL format) and converts them to human-readable Markdown format. It follows test-driven development (TDD) practices with comprehensive test coverage.

## Development Commands

```bash
# Run all tests
go test ./...

# Run tests for specific package
go test ./pkg/types/
go test ./internal/parser/
go test ./internal/formatter/

# Run tests with coverage
go test -cover ./...

# Build the application
go build -o cclog ./cmd/cclog/

# Run the built binary
./cclog [arguments]
```

## Architecture

The codebase follows Go's standard project layout with clear separation of concerns:

### Core Data Flow
1. **JSONL Parsing** (`internal/parser`) - Reads and parses conversation log files
2. **Type System** (`pkg/types`) - Defines message structures and conversation logs
3. **Markdown Formatting** (`internal/formatter`) - Converts parsed data to readable Markdown
4. **CLI Interface** (`cmd/cclog`) - Provides command-line interface

### Key Components

- **Message Type System**: The `types.Message` struct handles the complex JSONL structure from Claude conversations, including nested message content, timestamps, and metadata
- **Parser Strategy**: Line-by-line JSONL parsing with proper error handling and empty line skipping
- **Markdown Generation**: Time-sorted message processing with JST timezone conversion and content extraction from Claude's complex message format
- **Content Extraction**: Handles both simple string content and complex array-based content structures from Claude's message format

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

The test data in `testdata/sample.jsonl` represents actual Claude conversation log structure for realistic testing scenarios.

**TDD Development Flow:**
- Always write tests before implementation
- Run tests frequently during development (`go test ./...`)
- Ensure all tests pass before committing code
- Refactor only when tests are green
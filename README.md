# cclog - Claude Code Conversation Log Viewer & Converter

<img width="1620" height="1052" alt="Image" src="https://github.com/user-attachments/assets/fd7c3707-1fb3-4831-847e-f2f2bf9c0b2d" />

A Go command-line tool that parses Claude Code conversation logs (JSONL format), converts them to human-readable Markdown, and provides a powerful TUI to browse, manage, and interact with your logs.

## Features

- **Powerful Interactive TUI Mode**: A rich terminal interface to browse, preview, and manage your conversation logs.
    - **Live Markdown Preview**: Instantly preview how your `.jsonl` file will look in Markdown (`p` key).
    - **Open in Editor**: Convert and open logs directly in your default text editor with a single keypress (`Enter` key).
    - **On-the-fly Filtering**: Toggle message filters dynamically to switch between clean and raw views (`s` key).
    - **Easy Navigation**: Browse through directories and files with familiar keybindings.
    - **Recursive File Search**: Easily find all `.jsonl` logs within nested directories.
    - **Session ID to Clipboard**: Quickly copy a conversation's `sessionId` (from the filename) for other uses (`c` key).
    - **`claude` CLI Integration**: Resume conversations directly by launching the `claude` CLI (`r` key).
- **Flexible CLI Mode**: Process files or entire directories directly from the command line for scripting and automation.
- **Clean Markdown Output**: Converts conversations into a beautifully formatted, readable Markdown format.

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

- `[input]` - Path to a JSONL file or a directory.
  - If no input is provided, `cclog` starts in TUI mode, recursively searching from `~/.claude/projects` (if it exists), or the current directory.

### Options

- `-d, --directory` - Treat the input path as a directory and process all `.jsonl` files within it (non-TUI mode).
- `-o, --output FILE` - Write output to a specific file instead of stdout.
- `--include-all` - Include all messages in the output (disables filtering of empty/system messages).
- `--show-uuid` - Show the UUID metadata for each message in the output.
- `--show-title` - Show the conversation title as a header in the output.
- `--tui` - Force the application to start in interactive TUI mode.
- `-r, --recursive` - Enable recursive search for `.jsonl` files within the TUI. This is on by default if no input path is given or if `--path` is used.
- `--path PATH` - Start the TUI in the specified directory path (enables recursive search by default).
- `-h, --help` - Show the help message.

## Interactive TUI Mode

Running `cclog` without arguments (or with `--tui`, `--path`, or `-r`) launches the interactive TUI. This mode is more than a file picker; it's a complete interface for managing your logs.

### Keybindings

| Key         | Action                                                              |
|:------------|:--------------------------------------------------------------------|
| `↑`/`↓`/`j`/`k` | Navigate the file list.                                             |
| `enter`     | On a directory, enters it. On a file, converts it to Markdown and opens it in your default editor (`$EDITOR`). |
| `p`         | Toggle the live Markdown preview pane for the selected file.        |
| `s`         | Toggle the message filter on/off for previews and opened files.     |
| `c`         | Copy the `sessionId` (from the filename) of the selected log to the clipboard. |
| `r` / `R`   | Resume the conversation using the `claude` CLI (`claude -r <sessionId>`). `R` uses the `--dangerously-skip-permissions` flag. |
| `q`, `ctrl+c` | Quit the application.                                               |

## Examples

### Interactive Mode

```bash
# Start TUI with recursive search from the default path
cclog

# Start TUI in a specific directory (with recursive search)
cclog --path /path/to/my/logs

# Start TUI in the current directory (explicitly)
cclog --tui
```

### Command-Line Processing

```bash
# Convert a single file and print to stdout
cclog conversation.jsonl

# Convert a single file and save to an output file
cclog conversation.jsonl -o output.md

# Convert all .jsonl files in a directory and combine them into a single output file
cclog -d /path/to/logs -o combined.md

# Convert a file, including all system messages, and show UUIDs
cclog --include-all --show-uuid conversation.jsonl
```

## Output Format

The tool converts Claude Code conversation logs into clean Markdown format with:

- **Timestamps**: Automatically converted to the system's timezone for readability.
- **Message Filtering**: Removes system messages, API errors, and interrupted requests by default.
- **Content Extraction**: Handles both simple and complex message structures.
- **Readable Format**: Well-structured Markdown with proper formatting.

## Architecture

The project follows Go's standard project layout with clear separation of concerns:

- **`cmd/cclog`**: Main application entry point.
- **`internal/cli`**: Defines the command-line interface, argument parsing, and TUI entry.
- **`internal/parser`**: Reads and parses `.jsonl` conversation log files.
- **`internal/formatter`**: Handles message filtering and conversion to Markdown.
- **`pkg/filepicker`**: Implements the interactive TUI, including file listing, preview, and keybindings.
- **`pkg/types`**: Defines the core data structures for messages and conversations.

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

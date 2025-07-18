package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
	"github.com/annenpolka/cclog/pkg/types"
)

// Config represents command-line configuration
type Config struct {
	InputPath   string
	OutputPath  string
	IsDirectory bool
	ShowHelp    bool
	IncludeAll  bool
	ShowUUID    bool
	TUIMode     bool
	Recursive   bool
	ShowTitle   bool
}

// ParseArgs parses command-line arguments and returns configuration
func ParseArgs(args []string) (Config, error) {
	config := Config{}
	hasPathOption := false

	// Check if --path option is used to determine default behavior
	for i := 1; i < len(args); i++ {
		if args[i] == "--path" {
			hasPathOption = true
			break
		}
	}

	// If no arguments provided or --path option is used, enable TUI mode and recursive mode by default
	if len(args) < 2 || hasPathOption {
		config.TUIMode = true
		config.Recursive = true
		// Continue to process default directory setup below
	}

	// Only process arguments if we have them
	if len(args) >= 2 {
		for i := 1; i < len(args); i++ {
			arg := args[i]

			switch arg {
			case "-h", "--help":
				config.ShowHelp = true
				return config, nil
			case "-d", "--directory":
				config.IsDirectory = true
			case "-o", "--output":
				if i+1 >= len(args) {
					return Config{}, fmt.Errorf("output flag requires a value")
				}
				config.OutputPath = args[i+1]
				i++ // Skip next argument as it's the output path
			case "--include-all":
				config.IncludeAll = true
			case "--show-uuid":
				config.ShowUUID = true
			case "--show-title":
				config.ShowTitle = true
			case "--tui":
				config.TUIMode = true
			case "-r", "--recursive":
				config.Recursive = true
				config.TUIMode = true
			case "--path":
				if i+1 >= len(args) {
					return Config{}, fmt.Errorf("path flag requires a value")
				}
				config.InputPath = args[i+1]
				i++ // Skip next argument as it's the input path
			default:
				if config.InputPath == "" {
					config.InputPath = arg
				}
			}
		}
	}

	if config.InputPath == "" && !config.ShowHelp && !config.TUIMode {
		return Config{}, fmt.Errorf("input path is required")
	}

	// Set default directory for TUI mode if no input path specified
	if config.TUIMode && config.InputPath == "" {
		defaultDir := getDefaultTUIDirectory()
		// Check if the directory exists
		if err := ensureDefaultDirectoryExists(defaultDir); err != nil {
			// If directory doesn't exist, fall back to current directory
			config.InputPath = "."
		} else {
			config.InputPath = defaultDir
		}
	}

	return config, nil
}

// getDefaultTUIDirectory returns the default directory for TUI mode
// First tries $HOME/.claude/projects, then falls back to $HOME/.config/claude/projects
func getDefaultTUIDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "." // Fallback to current directory
	}

	// First try $HOME/.claude/projects
	claudeDir := filepath.Join(home, ".claude", "projects")
	if _, err := os.Stat(filepath.Join(home, ".claude")); err == nil {
		return claudeDir
	}

	// Fallback to $HOME/.config/claude/projects
	return filepath.Join(home, ".config", "claude", "projects")
}

// ensureDefaultDirectoryExists checks if the directory exists without creating it
func ensureDefaultDirectoryExists(dir string) error {
	_, err := os.Stat(dir)
	return err
}

// RunCommand executes the main command logic
func RunCommand(config Config) (string, error) {
	if config.ShowHelp {
		return GetHelpText(), nil
	}

	if config.TUIMode {
		// TUI mode is handled externally, return empty
		return "", nil
	}

	// Validate input path exists
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input path does not exist: %s", config.InputPath)
	}

	var markdown string

	if config.IsDirectory {
		// Parse directory
		logs, err := parser.ParseJSONLDirectory(config.InputPath)
		if err != nil {
			return "", fmt.Errorf("failed to parse directory: %w", err)
		}

		// Apply filtering to all logs
		filteredLogs := make([]*types.ConversationLog, len(logs))
		for i, log := range logs {
			filteredLogs[i] = formatter.FilterConversationLog(log, !config.IncludeAll)
		}

		markdown = formatter.FormatMultipleConversationsToMarkdown(filteredLogs, formatter.FormatOptions{
			ShowUUID:         config.ShowUUID,
			ShowPlaceholders: config.IncludeAll,
		})

		// Add title if requested
		if config.ShowTitle && len(filteredLogs) > 0 {
			title := types.ExtractTitle(filteredLogs[0])
			markdown = fmt.Sprintf("# %s\n\n%s", title, markdown)
		}
	} else {
		// Parse single file
		log, err := parser.ParseJSONLFile(config.InputPath)
		if err != nil {
			return "", fmt.Errorf("failed to parse file: %w", err)
		}

		// Apply filtering
		filteredLog := formatter.FilterConversationLog(log, !config.IncludeAll)
		markdown = formatter.FormatConversationToMarkdown(filteredLog, formatter.FormatOptions{
			ShowUUID:         config.ShowUUID,
			ShowPlaceholders: config.IncludeAll,
		})

		// Add title if requested
		if config.ShowTitle {
			title := types.ExtractTitle(filteredLog)
			markdown = fmt.Sprintf("# %s\n\n%s", title, markdown)
		}
	}

	// Write output if specified
	if config.OutputPath != "" {
		// Create output directory if it doesn't exist
		outputDir := filepath.Dir(config.OutputPath)
		if outputDir != "." {
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create output directory: %w", err)
			}
		}

		if err := os.WriteFile(config.OutputPath, []byte(markdown), 0644); err != nil {
			return "", fmt.Errorf("failed to write output file: %w", err)
		}
	}

	return markdown, nil
}

// GetHelpText returns the help text for the command
func GetHelpText() string {
	return strings.TrimSpace(`
cclog - Claude Conversation Log to Markdown Converter

USAGE:
    cclog [OPTIONS] [input]

ARGUMENTS:
    [input]    Path to JSONL file or directory containing JSONL files
               (If no input provided, opens interactive TUI mode with recursive search)

OPTIONS:
    -d, --directory    Treat input as directory (parse all .jsonl files)
    -o, --output FILE  Write output to file instead of stdout
    --include-all      Include all messages (no filtering of empty/system messages)
    --show-uuid        Show UUID metadata for each message
    --show-title       Show conversation title as header
    --tui              Open interactive file picker (TUI mode)
    -r, --recursive    Recursively search for .jsonl files and open TUI mode
    --path PATH        Specify directory path for TUI mode
    -h, --help         Show this help message

EXAMPLES:
    # Open interactive file picker with recursive search (default behavior)
    cclog

    # Open TUI in specific directory with recursive search
    cclog --path /path/to/logs

    # Convert single file to stdout
    cclog conversation.jsonl

    # Convert single file to output file
    cclog conversation.jsonl -o output.md

    # Convert all JSONL files in directory
    cclog -d /path/to/logs -o combined.md

    # Recursively find and list all JSONL files (explicit recursive mode)
    cclog -r /path/to/logs

    # Open interactive file picker (explicit TUI mode)
    cclog --tui
`)
}

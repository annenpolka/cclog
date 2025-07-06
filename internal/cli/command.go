package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"cclog/internal/formatter"
	"cclog/internal/parser"
	"cclog/pkg/types"
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
	
	// If no arguments provided, enable TUI mode by default
	if len(args) < 2 {
		config.TUIMode = true
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
		// Ensure the directory exists
		if err := ensureDefaultDirectoryExists(defaultDir); err != nil {
			// If we can't create the directory, fall back to current directory
			config.InputPath = "."
		} else {
			config.InputPath = defaultDir
		}
	}

	return config, nil
}

// getDefaultTUIDirectory returns the default directory for TUI mode
func getDefaultTUIDirectory() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "." // Fallback to current directory
	}
	return filepath.Join(home, ".claude", "projects")
}

// ensureDefaultDirectoryExists creates the directory if it doesn't exist
func ensureDefaultDirectoryExists(dir string) error {
	return os.MkdirAll(dir, 0755)
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

		markdown = formatter.FormatMultipleConversationsToMarkdownWithOptions(filteredLogs, formatter.FormatOptions{ShowUUID: config.ShowUUID})
		
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
		markdown = formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{ShowUUID: config.ShowUUID})
		
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
               (If no input provided, opens interactive TUI mode)

OPTIONS:
    -d, --directory    Treat input as directory (parse all .jsonl files)
    -o, --output FILE  Write output to file instead of stdout
    --include-all      Include all messages (no filtering of empty/system messages)
    --show-uuid        Show UUID metadata for each message
    --show-title       Show conversation title as header
    --tui              Open interactive file picker (TUI mode)
    -r, --recursive    Recursively search for .jsonl files and open TUI mode
    -h, --help         Show this help message

EXAMPLES:
    # Open interactive file picker (default behavior)
    cclog

    # Convert single file to stdout
    cclog conversation.jsonl

    # Convert single file to output file
    cclog conversation.jsonl -o output.md

    # Convert all JSONL files in directory
    cclog -d /path/to/logs -o combined.md

    # Recursively find and list all JSONL files (TUI mode enabled automatically)
    cclog -r /path/to/logs

    # Open interactive file picker (explicit TUI mode)
    cclog --tui

    # Open TUI in specific directory
    cclog --tui /path/to/logs
`)
}
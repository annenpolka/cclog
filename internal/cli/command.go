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
}

// ParseArgs parses command-line arguments and returns configuration
func ParseArgs(args []string) (Config, error) {
	if len(args) < 2 {
		return Config{}, fmt.Errorf("insufficient arguments")
	}

	config := Config{}
	
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
		default:
			if config.InputPath == "" {
				config.InputPath = arg
			}
		}
	}

	if config.InputPath == "" && !config.ShowHelp {
		return Config{}, fmt.Errorf("input path is required")
	}

	return config, nil
}

// RunCommand executes the main command logic
func RunCommand(config Config) (string, error) {
	if config.ShowHelp {
		return GetHelpText(), nil
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
	} else {
		// Parse single file
		log, err := parser.ParseJSONLFile(config.InputPath)
		if err != nil {
			return "", fmt.Errorf("failed to parse file: %w", err)
		}

		// Apply filtering
		filteredLog := formatter.FilterConversationLog(log, !config.IncludeAll)
		markdown = formatter.FormatConversationToMarkdownWithOptions(filteredLog, formatter.FormatOptions{ShowUUID: config.ShowUUID})
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
    cclog [OPTIONS] <input>

ARGUMENTS:
    <input>    Path to JSONL file or directory containing JSONL files

OPTIONS:
    -d, --directory    Treat input as directory (parse all .jsonl files)
    -o, --output FILE  Write output to file instead of stdout
    --include-all      Include all messages (no filtering of empty/system messages)
    --show-uuid        Show UUID metadata for each message
    -h, --help         Show this help message

EXAMPLES:
    # Convert single file to stdout
    cclog conversation.jsonl

    # Convert single file to output file
    cclog conversation.jsonl -o output.md

    # Convert all JSONL files in directory
    cclog -d /path/to/logs -o combined.md

    # Convert directory to stdout
    cclog -d /path/to/logs
`)
}
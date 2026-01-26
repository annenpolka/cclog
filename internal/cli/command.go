package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/annenpolka/cclog/internal/usecase"
	"github.com/annenpolka/cclog/pkg/filepicker"
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
	Latest      bool
	AutoOutput  bool // When true with --latest, auto-generate output filename from title
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
				// Check if next arg exists and is not another flag
				if i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") {
					// No value provided - mark for auto-generation (used with --latest)
					config.AutoOutput = true
				} else {
					config.OutputPath = args[i+1]
					i++ // Skip next argument as it is the output path
				}
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
			case "--latest":
				config.Latest = true
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

	// --latest mode requires a directory path but not TUI mode
	if config.Latest {
		config.TUIMode = false
		if config.InputPath == "" {
			config.InputPath = getDefaultTUIDirectory()
		}
		return config, nil
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

	// Handle --latest flag: output session UUID and title of the most recent conversation
	if config.Latest {
		// When -o is used, convert latest to markdown and save to file
		if config.AutoOutput || config.OutputPath != "" {
			return convertLatestToFile(config.InputPath, config.OutputPath)
		}
		return getLatestSessionInfo(config.InputPath)
	}

	if config.TUIMode {
		// TUI mode is handled externally, return empty
		return "", nil
	}

	// Validate input path exists
	if _, err := os.Stat(config.InputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input path does not exist: %s", config.InputPath)
	}

	opts := usecase.Options{
		IncludeAll: config.IncludeAll,
		ShowUUID:   config.ShowUUID,
		ShowTitle:  config.ShowTitle,
	}

	var markdown string
	var err error
	if config.IsDirectory {
		markdown, err = usecase.GenerateMarkdownFromDirectory(config.InputPath, opts)
		if err != nil {
			return "", fmt.Errorf("failed to parse directory: %w", err)
		}
	} else {
		markdown, err = usecase.GenerateMarkdownFromFile(config.InputPath, opts)
		if err != nil {
			return "", fmt.Errorf("failed to parse file: %w", err)
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

// getLatestSessionInfo retrieves the session UUID and title of the most recent conversation.
// It recursively searches the given directory for JSONL files, sorted by modification time,
// and returns the session ID and extracted title from the latest file.
func getLatestSessionInfo(dirPath string) (string, error) {
	files, err := filepicker.GetFilesRecursive(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to search directory: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no conversation files found in: %s", dirPath)
	}

	// Files are already sorted by ModTime descending, so the first file is the latest
	latestFile := files[0]
	println("Latest file found:", latestFile.Path)
	// Get session ID by parsing the file
	sessionID, err := usecase.GetSessionID(latestFile.Path)
	if err != nil {
		return "", fmt.Errorf("failed to get session ID: %w", err)
	}

	// Use the title already extracted by filepicker
	title := latestFile.ConversationTitle

	return fmt.Sprintf("%s\t%s", sessionID, title), nil
}

// convertLatestToFile finds the latest conversation, converts it to markdown,
// and saves it to the specified filename or an auto-generated one based on the title.
// If the file already exists, a counter suffix is added to avoid overwriting.
func convertLatestToFile(dirPath string, outputPath string) (string, error) {
	files, err := filepicker.GetFilesRecursive(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to search directory: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no conversation files found in: %s", dirPath)
	}

	// Files are already sorted by ModTime descending, so the first file is the latest
	latestFile := files[0]

	// Use provided filename or generate from title
	if outputPath == "" {
		outputPath = generateFilenameFromTitle(latestFile.ConversationTitle)
	}

	// Avoid overwriting existing files by adding counter suffix
	outputPath = getUniqueFilename(outputPath)

	// Convert to markdown
	markdown, err := usecase.GenerateMarkdownFromFile(latestFile.Path, usecase.Options{
		ShowTitle: true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to convert conversation: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return "", fmt.Errorf("failed to write output file: %w", err)
	}

	return fmt.Sprintf("Output written to: %s", outputPath), nil
}

// getUniqueFilename returns a filename that doesn't exist by adding a counter suffix.
// Example: "hello_world.md" -> "hello_world_1.md" -> "hello_world_2.md"
func getUniqueFilename(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path // File doesn't exist, use as-is
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}

// generateFilenameFromTitle creates a filename from the first two words of a title.
// Words are connected with underscores and the .md extension is added.
// For CJK text without spaces, it takes the first 20 characters instead.
// Example: "hello world foo bar" -> "hello_world.md"
func generateFilenameFromTitle(title string) string {
	// Split title into words
	words := strings.Fields(title)

	// Take first two words (or fewer if title is shorter)
	var parts []string
	for i := 0; i < len(words) && i < 2; i++ {
		// Clean word: remove non-alphanumeric characters except unicode letters
		cleaned := sanitizeForFilename(words[i])
		if cleaned != "" {
			parts = append(parts, cleaned)
		}
	}

	// Fallback if no valid words
	if len(parts) == 0 {
		return "conversation.md"
	}

	filename := strings.Join(parts, "_")

	// Limit filename length (for CJK text that may not have spaces)
	runes := []rune(filename)
	if len(runes) > 20 {
		filename = string(runes[:20])
	}

	return filename + ".md"
}

// sanitizeForFilename removes characters that are unsafe for filenames.
func sanitizeForFilename(s string) string {
	var result strings.Builder
	for _, r := range s {
		// Keep letters (including unicode), numbers, and some safe characters
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' ||
			(r >= 0x4E00 && r <= 0x9FFF) || // CJK characters
			(r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) { // Katakana
			result.WriteRune(r)
		}
	}
	return result.String()
}

// applyOptionalTitle prefixes the markdown with a top-level title when enabled.
// For multiple logs, it uses the first conversation to extract the title,
// matching the previous behavior.

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
    --latest           Output the session UUID and title of the most recent conversation
                       (with -o: convert to markdown; omit FILE for auto-naming from title)
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

    # Get the latest session UUID and title
    cclog --latest

    # Convert latest conversation to markdown (auto-named from title)
    cclog --latest -o

    # Convert latest conversation to specific file
    cclog --latest -o output.md

    # Open interactive file picker (explicit TUI mode)
    cclog --tui
`)
}

package usecase

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/annenpolka/cclog/internal/domain"
	"github.com/annenpolka/cclog/internal/formatter"
	"github.com/annenpolka/cclog/internal/parser"
)

// Options controls rendering behavior.
type Options struct {
	IncludeAll bool // when true: no filtering; also enables placeholders in content
	ShowUUID   bool
	ShowTitle  bool
}

func resolveOptions(opts ...Options) Options {
	if len(opts) > 0 {
		return opts[0]
	}
	return Options{}
}

// GenerateMarkdownFromPath parses a file or directory and returns Markdown rendering.
func GenerateMarkdownFromPath(path string, opts ...Options) (string, error) {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("input path does not exist: %s", path)
		}
		return "", err
	}
	if st.IsDir() {
		return GenerateMarkdownFromDirectory(path, opts...)
	}
	return GenerateMarkdownFromFile(path, opts...)
}

// GenerateMarkdownFromFile parses a single JSONL file and returns its Markdown rendering.
func GenerateMarkdownFromFile(filePath string, opts ...Options) (string, error) {
	o := resolveOptions(opts...)
	log, err := parser.ParseJSONLFile(filePath)
	if err != nil {
		return "", err
	}
	enableFiltering := !o.IncludeAll
	filtered := formatter.FilterConversationLog(log, enableFiltering)
	md := formatter.FormatConversationToMarkdown(filtered, formatter.FormatOptions{
		ShowUUID:         o.ShowUUID,
		ShowPlaceholders: o.IncludeAll,
	})
	if o.ShowTitle {
		title := domain.ExtractTitle(filtered)
		md = fmt.Sprintf("# %s\n\n%s", title, md)
	}
	return md, nil
}

// GetSessionID extracts the session ID from the JSONL file name.
// Claude Code conversation files are named with their session UUID (e.g., "a1b2c3d4-...-.jsonl").
func GetSessionID(filePath string) (string, error) {
	base := filepath.Base(filePath)
	sessionID := strings.TrimSuffix(base, ".jsonl")
	if sessionID == "" || sessionID == base {
		return "", fmt.Errorf("invalid JSONL file name: %s", filePath)
	}
	return sessionID, nil
}

// GenerateMarkdownFromDirectory parses all JSONL files in a directory and returns a combined Markdown.
func GenerateMarkdownFromDirectory(dirPath string, opts ...Options) (string, error) {
	o := resolveOptions(opts...)
	logs, err := parser.ParseJSONLDirectory(dirPath)
	if err != nil {
		return "", err
	}
	enableFiltering := !o.IncludeAll
	filteredLogs := make([]*domain.ConversationLog, len(logs))
	for i, log := range logs {
		filteredLogs[i] = formatter.FilterConversationLog(log, enableFiltering)
	}
	md := formatter.FormatMultipleConversationsToMarkdown(filteredLogs, formatter.FormatOptions{
		ShowUUID:         o.ShowUUID,
		ShowPlaceholders: o.IncludeAll,
	})
	if o.ShowTitle && len(filteredLogs) > 0 {
		title := domain.ExtractTitle(filteredLogs[0])
		md = fmt.Sprintf("# %s\n\n%s", title, md)
	}
	return md, nil
}

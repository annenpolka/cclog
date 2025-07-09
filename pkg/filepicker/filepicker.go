package filepicker

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	
	"github.com/annenpolka/cclog/internal/parser"
	"github.com/annenpolka/cclog/pkg/types"
)

type FileInfo struct {
	Name              string
	Path              string
	IsDir             bool
	Size              int64
	ModTime           time.Time
	ConversationTitle string
	ProjectName       string
}

func (f FileInfo) FilterValue() string {
	return f.Name
}

func (f FileInfo) Title() string {
	if f.IsDir {
		return f.Name + "/"
	}
	
	// For JSONL files, display "date [project] title" format
	if filepath.Ext(f.Name) == ".jsonl" {
		dateStr := f.ModTime.Format("2006-01-02 15:04")
		
		// Add project name if available
		var projectPart string
		if f.ProjectName != "" {
			projectPart = " [" + f.ProjectName + "]"
		}
		
		// Add conversation title if available
		if f.ConversationTitle != "" {
			return dateStr + projectPart + " " + f.ConversationTitle
		}
		
		// If no title but has project name, show date [project]
		if f.ProjectName != "" {
			return dateStr + projectPart
		}
		
		return dateStr
	}
	
	return f.Name
}

func (f FileInfo) Description() string {
	// Return empty string for clean display - date is shown in Title for JSONL files
	return ""
}

func GetFiles(dir string) ([]FileInfo, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	
	// Add parent directory entry if not at root
	absDir, err := filepath.Abs(dir)
	if err == nil {
		parentDir := filepath.Dir(absDir)
		// Only add ".." if not at root and parent is different
		if parentDir != absDir && parentDir != "." {
			// Get actual modification time for parent directory
			var parentModTime time.Time
			if parentStat, err := os.Stat(parentDir); err == nil {
				parentModTime = parentStat.ModTime()
			}
			
			parentInfo := FileInfo{
				Name:    "..",
				Path:    parentDir,
				IsDir:   true,
				Size:    0,
				ModTime: parentModTime,
			}
			files = append(files, parentInfo)
		}
	}
	
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(dir, entry.Name()),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		
		// Extract conversation title and project name for JSONL files
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".jsonl" {
			title, projectName := extractConversationInfo(fileInfo.Path)
			// Skip empty files (when title extraction fails due to empty file)
			if title == "" {
				continue
			}
			fileInfo.ConversationTitle = title
			fileInfo.ProjectName = projectName
		}
		files = append(files, fileInfo)
	}

	// Sort files by modification time (newest first)
	// Keep parent directory at the beginning if it exists
	var parentDir *FileInfo
	var regularFiles []FileInfo
	
	for i, file := range files {
		if file.Name == ".." {
			parentDir = &files[i]
		} else {
			regularFiles = append(regularFiles, file)
		}
	}
	
	// Sort regular files by modification time (newest first)
	sort.Slice(regularFiles, func(i, j int) bool {
		return regularFiles[i].ModTime.After(regularFiles[j].ModTime)
	})
	
	// Rebuild files slice with parent directory first (if exists)
	var sortedFiles []FileInfo
	if parentDir != nil {
		sortedFiles = append(sortedFiles, *parentDir)
	}
	sortedFiles = append(sortedFiles, regularFiles...)
	
	return sortedFiles, nil
}

// extractConversationInfo extracts title and project name from JSONL conversation file
func extractConversationInfo(filePath string) (string, string) {
	// Parse the JSONL file to extract conversation information
	log, err := parser.ParseJSONLFile(filePath)
	if err != nil {
		return "", ""
	}
	
	// Skip empty files - return empty string to indicate this file should be filtered out
	if len(log.Messages) == 0 {
		return "", ""
	}
	
	// Extract project name from CWD field of the first message that has one
	var projectName string
	for _, msg := range log.Messages {
		if msg.CWD != "" {
			projectName = extractProjectName(msg.CWD)
			break
		}
	}
	
	// Apply filtering to check if any meaningful messages remain after filtering
	filteredLog := &types.ConversationLog{
		Messages: make([]types.Message, 0),
		FilePath: log.FilePath,
	}
	
	// Manually filter messages using the same logic as formatter.FilterConversationLog
	for _, msg := range log.Messages {
		// Apply the same filtering logic as IsContentfulMessage
		if isContentfulMessage(msg) {
			filteredLog.Messages = append(filteredLog.Messages, msg)
		}
	}
	
	// Skip files with no meaningful messages after filtering
	if len(filteredLog.Messages) == 0 {
		return "", ""
	}
	
	// Extract title using existing title extraction logic
	title := types.ExtractTitle(filteredLog)
	return title, projectName
}

// extractConversationTitle extracts title from JSONL conversation file (backward compatibility)
func extractConversationTitle(filePath string) string {
	title, _ := extractConversationInfo(filePath)
	return title
}

// isContentfulMessage replicates the filtering logic from formatter package
// to avoid circular imports while maintaining consistency
func isContentfulMessage(msg types.Message) bool {
	// Filter out system messages
	if msg.Type == "system" {
		return false
	}

	// Filter out summary messages
	if msg.Type == "summary" {
		return false
	}

	// Filter out meta messages
	if msg.IsMeta {
		return false
	}

	// Extract content and check if it's meaningful
	content := extractMessageContent(msg.Message)
	
	// Filter out empty messages
	if content == "" {
		return false
	}

	// Filter out API errors
	if strings.Contains(content, "API Error") {
		return false
	}

	// Filter out interrupted requests
	if strings.Contains(content, "[Request interrupted") {
		return false
	}

	// Filter out command messages
	if strings.Contains(content, "<command-name>") {
		return false
	}

	// Filter out bash inputs
	if strings.Contains(content, "<bash-input>") {
		return false
	}

	// Filter out command outputs
	if strings.Contains(content, "<local-command-stdout>") {
		return false
	}

	// Filter out system reminders and caveats
	if strings.Contains(content, "Caveat: The messages below were generated") {
		return false
	}

	return true
}

// extractMessageContent extracts string content from message
func extractMessageContent(message any) string {
	// Handle different message content types
	switch v := message.(type) {
	case map[string]any:
		if content, ok := v["content"]; ok {
			switch contentVal := content.(type) {
			case string:
				return contentVal
			case []any:
				// Handle array-based content (Claude's complex message format)
				var result strings.Builder
				for _, item := range contentVal {
					if itemMap, ok := item.(map[string]any); ok {
						if text, ok := itemMap["text"].(string); ok {
							result.WriteString(text)
						} else if itemType, ok := itemMap["type"].(string); ok && itemType == "text" {
							if text, ok := itemMap["text"].(string); ok {
								result.WriteString(text)
							}
						}
					}
				}
				return result.String()
			}
		}
	case string:
		return v
	}
	return ""
}

// GetFilesRecursive recursively collects all .jsonl files from a directory and its subdirectories
func GetFilesRecursive(rootDir string) ([]FileInfo, error) {
	var allFiles []FileInfo
	
	err := filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		// Skip directories
		if d.IsDir() {
			return nil
		}
		
		// Only include .jsonl files
		if filepath.Ext(d.Name()) != ".jsonl" {
			return nil
		}
		
		// Get file info for modification time
		info, err := d.Info()
		if err != nil {
			return err
		}
		
		fileInfo := FileInfo{
			Name:    d.Name(),
			Path:    path,
			IsDir:   false,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}
		
		// Extract conversation title and project name for JSONL files
		title, projectName := extractConversationInfo(path)
		// Skip empty files (when title extraction fails due to empty file)
		if title == "" {
			return nil
		}
		fileInfo.ConversationTitle = title
		fileInfo.ProjectName = projectName
		
		allFiles = append(allFiles, fileInfo)
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	// Sort by modification time (newest first)
	sort.Slice(allFiles, func(i, j int) bool {
		return allFiles[i].ModTime.After(allFiles[j].ModTime)
	})
	
	return allFiles, nil
}

// extractProjectName extracts project name from cwd path
func extractProjectName(cwd string) string {
	if cwd == "" || cwd == "/" {
		return ""
	}
	
	// Clean the path and get the base name
	cleanPath := filepath.Clean(cwd)
	projectName := filepath.Base(cleanPath)
	
	// Return empty string if it's root or dot
	if projectName == "/" || projectName == "." {
		return ""
	}
	
	return projectName
}
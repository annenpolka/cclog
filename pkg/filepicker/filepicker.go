package filepicker

import (
	"os"
	"path/filepath"
	"sort"
	"time"
)

type FileInfo struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
}

func (f FileInfo) FilterValue() string {
	return f.Name
}

func (f FileInfo) Title() string {
	if f.IsDir {
		return f.Name + "/"
	}
	return f.Name
}

func (f FileInfo) Description() string {
	// Display modification time in YYYY-MM-DD HH:MM format for both files and directories
	return f.ModTime.Format("2006-01-02 15:04")
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
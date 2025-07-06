package filepicker

import (
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
	Size  int64
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
	if f.IsDir {
		return "Directory"
	}
	if f.Size < 1024 {
		return "< 1KB"
	}
	return "1KB+" // 簡単な実装
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
			parentInfo := FileInfo{
				Name:  "..",
				Path:  parentDir,
				IsDir: true,
				Size:  0,
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
			Name:  entry.Name(),
			Path:  filepath.Join(dir, entry.Name()),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}
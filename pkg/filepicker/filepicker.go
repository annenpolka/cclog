package filepicker

import (
	"os"
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
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:  entry.Name(),
			Path:  dir + "/" + entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}
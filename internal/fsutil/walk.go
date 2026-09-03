package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnoreDirs = map[string]bool{
	".git":            true,
	".dart_tool":       true,
	".idea":            true,
	".vscode":          false,
	"build":            true,
	".gradle":          true,
	"Pods":             true,
	".symlinks":        true,
	"DerivedData":      true,
	"node_modules":     true,
	".fvm":             true,
}

var defaultIgnoreFiles = map[string]bool{
	".DS_Store":         true,
	".flutter-plugins":  true,
	".flutter-plugins-dependencies": true,
	".packages":         true,
}

const maxFileSizeBytes = 45 * 1024 * 1024

type FileEntry struct {
	AbsPath string
	RelPath string
	Size    int64
}

func CollectFiles(root string) ([]FileEntry, []string, error) {
	var files []FileEntry
	var skipped []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		if rel == "." {
			return nil
		}
		name := info.Name()

		if info.IsDir() {
			if defaultIgnoreDirs[name] {
				return filepath.SkipDir
			}
			return nil
		}

		if defaultIgnoreFiles[name] {
			return nil
		}
		if strings.HasSuffix(name, ".iml") {
			return nil
		}
		if info.Size() > maxFileSizeBytes {
			skipped = append(skipped, rel+" (too large, skipped)")
			return nil
		}

		files = append(files, FileEntry{
			AbsPath: path,
			RelPath: filepath.ToSlash(rel),
			Size:    info.Size(),
		})
		return nil
	})

	if err != nil {
		return nil, nil, err
	}
	return files, skipped, nil
}

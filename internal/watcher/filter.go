package watcher

import (
	"path/filepath"
	"strings"
)

// ShouldIgnore returns true if path should be ignored
// Ignores: .git/, node_modules/, *.tmp, *.swp, .DS_Store, hidden files
func ShouldIgnore(path string) bool {
	base := filepath.Base(path)
	
	// .git directory specifically (but allow .gitignore file)
	if strings.Contains(path, string(filepath.Separator)+".git"+string(filepath.Separator)) ||
	   strings.HasSuffix(path, string(filepath.Separator)+".git") {
		return true
	}
	
	// Hidden files and directories (starting with .) - except .gitignore
	if strings.HasPrefix(base, ".") && base != ".gitignore" {
		return true
	}
	
	// Common ignore directories
	ignoreDirs := []string{"node_modules", "vendor", "bin", "dist", "build"}
	for _, dir := range ignoreDirs {
		if strings.Contains(path, string(filepath.Separator)+dir+string(filepath.Separator)) ||
		   strings.HasSuffix(path, string(filepath.Separator)+dir) {
			return true
		}
	}
	
	// Temporary files
	ignoreExts := []string{".tmp", ".swp", ".swo", ".bak", ".log"}
	ext := filepath.Ext(base)
	for _, ignoreExt := range ignoreExts {
		if ext == ignoreExt {
			return true
		}
	}
	
	// Specific files
	ignoreFiles := []string{".DS_Store", "Thumbs.db"}
	for _, file := range ignoreFiles {
		if base == file {
			return true
		}
	}
	
	return false
}

// IsGoFile returns true if path is a .go file
func IsGoFile(path string) bool {
	return filepath.Ext(path) == ".go"
}

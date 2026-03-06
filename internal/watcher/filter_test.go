package watcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldIgnore(t *testing.T) {
	tests := []struct {
		path   string
		ignore bool
	}{
		{"/project/main.go", false},
		{"/project/.git/config", true},
		{"/project/.gitignore", false}, // Exception
		{"/project/node_modules/pkg/index.js", true},
		{"/project/file.swp", true},
		{"/project/file.tmp", true},
		{"/project/.DS_Store", true},
		{"/project/.hidden", true},
		{"/project/bin/server", true},
		{"/project/vendor/pkg", true},
		{"/project/dist/build.js", true},
		{"/project/build/output.o", true},
		{"/project/Thumbs.db", true},
		{"/project/normal.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := ShouldIgnore(tt.path)
			assert.Equal(t, tt.ignore, result, "Path: %s", tt.path)
		})
	}
}

func TestIsGoFile(t *testing.T) {
	tests := []struct {
		path     string
		isGoFile bool
	}{
		{"/project/main.go", true},
		{"/project/internal/handler.go", true},
		{"/project/test.js", false},
		{"/project/README.md", false},
		{"/project.go", true},
		{"/project.go.bak", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := IsGoFile(tt.path)
			assert.Equal(t, tt.isGoFile, result)
		})
	}
}

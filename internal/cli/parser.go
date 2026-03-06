package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Root      string // absolute path to project root
	BuildCmd  string // full build command
	ExecCmd   string // full execution command
}

// ParseArgs parses CLI arguments and returns validated config
// Returns error if any required flag is missing or invalid
func ParseArgs(args []string) (*Config, error) {
	config := &Config{}
	
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--root":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("required flag --root not provided")
			}
			config.Root = args[i+1]
			i++
		case "--build":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("required flag --build not provided")
			}
			config.BuildCmd = args[i+1]
			i++
		case "--exec":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("required flag --exec not provided")
			}
			config.ExecCmd = args[i+1]
			i++
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		case "--version", "-v":
			printVersion()
			os.Exit(0)
		case "--verbose":
			// Will be handled by logger
		default:
			return nil, fmt.Errorf("unknown flag: %s", args[i])
		}
	}
	
	return config, nil
}

// Validate checks if config paths exist and commands are non-empty
func (c *Config) Validate() error {
	// Check required fields
	if c.Root == "" {
		return fmt.Errorf("required flag --root not provided")
	}
	if c.BuildCmd == "" {
		return fmt.Errorf("required flag --build not provided")
	}
	if c.ExecCmd == "" {
		return fmt.Errorf("required flag --exec not provided")
	}
	
	// Convert to absolute path
	if !filepath.IsAbs(c.Root) {
		abs, err := filepath.Abs(c.Root)
		if err != nil {
			return fmt.Errorf("failed to convert root to absolute path: %w", err)
		}
		c.Root = abs
	}
	
	// Check if directory exists
	info, err := os.Stat(c.Root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("root directory does not exist: %s", c.Root)
		}
		return fmt.Errorf("failed to access root directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("root must be a directory, got file: %s", c.Root)
	}
	
	return nil
}

func printUsage() {
	fmt.Printf(`hotreload - Automatic rebuild and restart for development

USAGE:
  hotreload --root <directory> --build <command> --exec <command>

REQUIRED FLAGS:
  --root <directory>   Project directory to watch for changes
  --build <command>    Command to build the project
  --exec <command>     Command to run the built application

OPTIONAL FLAGS:
  --help, -h           Show this help message
  --version, -v        Show version information
  --verbose            Enable verbose logging

EXAMPLES:
  # Go HTTP server
  hotreload --root ./myproject \
            --build "go build -o ./bin/server ./cmd/server" \
            --exec "./bin/server"

  # Make-based build
  hotreload --root ./myapp \
            --build "make build" \
            --exec "./bin/app"

For more information: https://github.com/yourorg/hotreload
`)
}

func printVersion() {
	fmt.Printf("hotreload version 1.0.0\n")
	fmt.Printf("Go version: %s\n", "go1.22.0")
	fmt.Printf("Built: 2026-03-06T10:00:00Z\n")
}

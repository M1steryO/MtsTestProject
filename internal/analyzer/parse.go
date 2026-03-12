package analyzer

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func Parse(path string) (moduleName, goVersion string, err error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", "", fmt.Errorf("go.mod not found in repository root")
		}
		return "", "", fmt.Errorf("open go.mod: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if idx := strings.Index(line, "//"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "module":
			if moduleName == "" {
				moduleName = fields[1]
			}
		case "go":
			if goVersion == "" {
				goVersion = fields[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("read go.mod: %w", err)
	}

	if moduleName == "" {
		return "", "", fmt.Errorf("module directive not found in go.mod")
	}
	if goVersion == "" {
		goVersion = "unknown"
	}

	return moduleName, goVersion, nil
}

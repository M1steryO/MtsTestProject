package tests

import (
	"github.com/M1steryO/MtsTestProject/internal/analyzer"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseGoMod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		content      string
		wantModule   string
		wantGo       string
		wantErr      bool
		errSubstring string
	}{
		{
			name: "success",
			content: `
module github.com/example/project

go 1.22.0

require github.com/stretchr/testify v1.9.0
`,
			wantModule: "github.com/example/project",
			wantGo:     "1.22.0",
		},
		{
			name: "success with inline comments",
			content: `
module github.com/example/project // module name
go 1.21.5 // go version
`,
			wantModule: "github.com/example/project",
			wantGo:     "1.21.5",
		},
		{
			name: "success with full line comments",
			content: `
// comment
module github.com/example/project

// another comment
go 1.20
`,
			wantModule: "github.com/example/project",
			wantGo:     "1.20",
		},
		{
			name: "missing go version returns unknown",
			content: `
module github.com/example/project
`,
			wantModule: "github.com/example/project",
			wantGo:     "unknown",
		},
		{
			name: "missing module returns error",
			content: `
go 1.22.0
`,
			wantErr:      true,
			errSubstring: "module directive not found",
		},
		{
			name:         "empty file returns error",
			content:      ``,
			wantErr:      true,
			errSubstring: "module directive not found",
		},
		{
			name: "ignore unrelated lines",
			content: `
toolchain go1.22.1
go 1.22.0
require github.com/stretchr/testify v1.9.0
module github.com/example/project
`,
			wantModule: "github.com/example/project",
			wantGo:     "1.22.0",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := writeTempFile(t, tt.content)

			gotModule, gotGo, err := analyzer.Parse(path)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errSubstring != "" && !strings.Contains(err.Error(), tt.errSubstring) {
					t.Fatalf("expected error containing %q, got %q", tt.errSubstring, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotModule != tt.wantModule {
				t.Fatalf("module mismatch: got %q, want %q", gotModule, tt.wantModule)
			}

			if gotGo != tt.wantGo {
				t.Fatalf("go version mismatch: got %q, want %q", gotGo, tt.wantGo)
			}
		})
	}
}

func TestParseGoMod_FileNotFound(t *testing.T) {
	t.Parallel()

	_, _, err := analyzer.Parse(filepath.Join(t.TempDir(), "go.mod"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "go.mod not found in repository root") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	return path
}

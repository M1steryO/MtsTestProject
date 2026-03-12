package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/M1steryO/MtsTestProject/internal/model"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func Find(ctx context.Context, repoDir string) ([]model.Dependency, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-u", "-json", "all")
	cmd.Dir = repoDir
	cmd.Env = append(os.Environ(), "GOWORK=off")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare go list stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("prepare go list stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start go list: %w", err)
	}

	var stderrBytes []byte
	errCh := make(chan error, 1)

	go func() {
		var readErr error
		stderrBytes, readErr = io.ReadAll(stderr)
		errCh <- readErr
	}()

	dec := json.NewDecoder(stdout)
	deps := make([]model.Dependency, 0)

	for {
		var m model.GoListModule
		if err := dec.Decode(&m); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			_ = cmd.Wait()
			<-errCh
			return nil, fmt.Errorf("decode go list output: %w", err)
		}

		if m.Main || m.Update == nil {
			continue
		}

		deps = append(deps, model.Dependency{
			Path:           m.Path,
			CurrentVersion: m.Version,
			LatestVersion:  m.Update.Version,
			Indirect:       m.Indirect,
		})
	}

	if err := <-errCh; err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("read stderr: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("go list timeout or canceled: %w", ctx.Err())
		}

		msg := strings.TrimSpace(string(stderrBytes))
		if msg != "" {
			return nil, fmt.Errorf("go list failed: %w\n%s", err, msg)
		}
		return nil, fmt.Errorf("go list failed: %w", err)
	}

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Path < deps[j].Path
	})

	return deps, nil
}

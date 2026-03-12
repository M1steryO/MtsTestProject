package analyzer

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func Clone(ctx context.Context, repoURL, dst string) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", "--filter=blob:none", "--sparse", repoURL, dst)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

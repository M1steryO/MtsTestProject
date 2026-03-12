package analyzer

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

const defaultTimeout = time.Second * 20

type githubContentResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// Clone подготавливает репозиторий во временной папке:
// для GitHub скачивает go.mod/go.sum по API,
// для остальных URL делает shallow clone через go-git.
func Clone(ctx context.Context, repoURL, dst string) error {
	if isGitHubRepoURL(repoURL) {
		return downloadGitHubFiles(ctx, repoURL, dst)
	}

	return cloneWithGoGit(ctx, repoURL, dst)
}

func cloneWithGoGit(ctx context.Context, repoURL, dst string) error {
	_, err := git.PlainCloneContext(ctx, dst, false, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
	})
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("clone timeout or canceled: %w", ctx.Err())
		}
		return fmt.Errorf("clone repo: %w", err)
	}

	return nil
}

func isGitHubRepoURL(repoURL string) bool {
	u, err := url.Parse(repoURL)
	if err != nil {
		return false
	}

	host := strings.ToLower(u.Host)
	return host == "github.com" || host == "www.github.com"
}

func downloadGitHubFiles(ctx context.Context, repoURL, dst string) error {
	owner, repo, err := parseGitHubRepoURL(repoURL)
	if err != nil {
		return err
	}

	goMod, err := fetchGitHubFile(ctx, owner, repo, "go.mod")
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("go.mod not found in repository root")
		}
		return err
	}

	if err := os.WriteFile(filepath.Join(dst, "go.mod"), goMod, 0o644); err != nil {
		return fmt.Errorf("write go.mod: %w", err)
	}

	// go.sum не обязателен, поэтому просто пропускаем, если файла нет.
	goSum, err := fetchGitHubFile(ctx, owner, repo, "go.sum")
	if err == nil {
		if err := os.WriteFile(filepath.Join(dst, "go.sum"), goSum, 0o644); err != nil {
			return fmt.Errorf("write go.sum: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func parseGitHubRepoURL(repoURL string) (string, string, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("parse repo url: %w", err)
	}

	trimmedPath := strings.Trim(strings.TrimSuffix(u.Path, ".git"), "/")
	parts := strings.Split(trimmedPath, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid GitHub repository URL")
	}

	return parts[0], parts[1], nil
}

func fetchGitHubFile(ctx context.Context, owner, repo, filePath string) ([]byte, error) {
	reqCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	apiURL := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s",
		url.PathEscape(owner),
		url.PathEscape(repo),
		path.Clean(filePath),
	)

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", filePath, err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	// Для приватных репозиториев можно передать токен через env.
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	}

	client := &http.Client{
		Timeout: defaultTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", filePath, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusNotFound {
		return nil, os.ErrNotExist
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %s for %s", resp.Status, filePath)
	}

	var payload githubContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode response for %s: %w", filePath, err)
	}

	if payload.Type != "file" {
		return nil, fmt.Errorf("%s is not a file", filePath)
	}

	if payload.Encoding != "base64" {
		return nil, fmt.Errorf("unsupported encoding %q for %s", payload.Encoding, filePath)
	}

	raw := strings.ReplaceAll(payload.Content, "\n", "")
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode base64 for %s: %w", filePath, err)
	}

	return data, nil
}

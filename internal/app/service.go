package app

import (
	"context"
	"fmt"
	"github.com/M1steryO/MtsTestProject/internal/analyzer"
	"github.com/M1steryO/MtsTestProject/internal/model"
	"os"
	"path/filepath"
	"time"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Run(ctx context.Context, repoURL string) (model.Result, error) {
	tmpDir, err := os.MkdirTemp("", "repo-check-*")
	if err != nil {
		return model.Result{}, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	cloneCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := analyzer.Clone(cloneCtx, repoURL, tmpDir); err != nil {
		return model.Result{}, err
	}

	moduleName, goVersion, err := analyzer.Parse(filepath.Join(tmpDir, "go.mod"))
	if err != nil {
		return model.Result{}, err
	}

	updateCtx, cancel2 := context.WithTimeout(ctx, 60*time.Second)
	defer cancel2()

	deps, err := analyzer.Find(updateCtx, tmpDir)
	if err != nil {
		return model.Result{}, err
	}

	return model.Result{
		Module:    moduleName,
		GoVersion: goVersion,
		Updates:   deps,
	}, nil
}

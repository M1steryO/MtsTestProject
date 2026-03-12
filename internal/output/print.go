package output

import (
	"encoding/json"
	"fmt"
	"github.com/M1steryO/MtsTestProject/internal/model"
	"io"
)

// PrintJSON печатает результат в формате JSON.
func PrintJSON(w io.Writer, res model.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// PrintText печатает результат в текстовом виде.
func PrintText(w io.Writer, res model.Result) error {
	if _, err := fmt.Fprintf(w, "Module: %s\n", res.Module); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Go version: %s\n", res.GoVersion); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Dependencies with available updates: %d\n", len(res.Updates)); err != nil {
		return err
	}

	for _, dep := range res.Updates {
		indirect := ""
		if dep.Indirect {
			indirect = " (indirect)"
		}

		if _, err := fmt.Fprintf(
			w,
			"- %s: %s -> %s%s\n",
			dep.Path,
			dep.CurrentVersion,
			dep.LatestVersion,
			indirect,
		); err != nil {
			return err
		}
	}

	return nil
}

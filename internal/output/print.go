package output

import (
	"encoding/json"
	"fmt"
	"github.com/M1steryO/MtsTestProject/internal/model"
	"io"
)

func PrintJSON(w io.Writer, res model.Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

func PrintText(w io.Writer, res model.Result) {
	fmt.Fprintf(w, "Module: %s\n", res.Module)
	fmt.Fprintf(w, "Go version: %s\n", res.GoVersion)
	fmt.Fprintf(w, "Dependencies with available updates: %d\n", len(res.Updates))

	for _, dep := range res.Updates {
		indirect := ""
		if dep.Indirect {
			indirect = " (indirect)"
		}

		fmt.Fprintf(
			w,
			"- %s: %s -> %s%s\n",
			dep.Path,
			dep.CurrentVersion,
			dep.LatestVersion,
			indirect,
		)
	}
}

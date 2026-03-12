package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/M1steryO/MtsTestProject/internal/app"
	"github.com/M1steryO/MtsTestProject/internal/output"
	"os"
	"path/filepath"
)

func main() {
	jsonOutput := flag.Bool("json", false, "print result as JSON")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "usage: %s [--json] <git-repo-url>\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	service := app.NewService()
	result, err := service.Run(context.Background(), flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	if *jsonOutput {
		if err := output.PrintJSON(os.Stdout, result); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	output.PrintText(os.Stdout, result)
}

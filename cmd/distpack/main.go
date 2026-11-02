package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-saas/erp-agent-skill/internal/distpack"
)

func main() {
	format := flag.String("format", "", "archive format: tar.gz or zip")
	source := flag.String("source", "", "directory containing staged release files")
	output := flag.String("output", "", "archive output path")
	root := flag.String("root", "", "archive root directory")
	binary := flag.String("binary", "", "staged binary file name")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "distpack: positional arguments are not supported")
		os.Exit(2)
	}

	if err := distpack.Write(distpack.Options{
		Format:     *format,
		SourceDir:  *source,
		OutputPath: *output,
		RootName:   *root,
		BinaryName: *binary,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "distpack: %v\n", err)
		os.Exit(1)
	}
}

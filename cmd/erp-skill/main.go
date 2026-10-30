package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	bundle "github.com/go-saas/erp-agent-skill"
	"github.com/go-saas/erp-agent-skill/internal/cli"
	"github.com/go-saas/erp-agent-skill/internal/manifest"
)

func main() {
	value, err := manifest.Load(bytes.NewReader(bundle.ManifestJSON))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app, err := cli.NewDefault(value.Version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := app.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

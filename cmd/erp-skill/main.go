package main

import (
	"bytes"
	"fmt"
	"os"

	bundle "github.com/go-saas/erp-agent-skill"
	"github.com/go-saas/erp-agent-skill/internal/manifest"
)

func main() {
	value, err := manifest.Load(bytes.NewReader(bundle.ManifestJSON))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(os.Args) == 2 && os.Args[1] == "version" {
		fmt.Printf("erp-skill %s\n", value.Version)
		return
	}
	fmt.Fprintln(os.Stderr, "usage: erp-skill version")
	os.Exit(2)
}

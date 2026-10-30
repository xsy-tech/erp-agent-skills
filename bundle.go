package bundle

import _ "embed"

// ManifestJSON is the release manifest embedded in the erp-skill binary.
//
//go:embed manifest.json
var ManifestJSON []byte

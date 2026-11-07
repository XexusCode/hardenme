// Package checks embeds the check definitions (checks/<platform>/*.yaml)
// and their JSON Schema into the binary.
package checks

import "embed"

// FS holds every check definition and schema.json.
//
//go:embed schema.json */*.yaml
var FS embed.FS

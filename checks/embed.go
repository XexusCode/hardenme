// Package checks embeds the check definitions (checks/<platform>/*.yaml),
// the theft stories (checks/stories/<platform>.yaml) and the JSON Schema
// into the binary.
package checks

import "embed"

// FS holds every check definition and schema.json.
//
//go:embed schema.json */*.yaml
var FS embed.FS

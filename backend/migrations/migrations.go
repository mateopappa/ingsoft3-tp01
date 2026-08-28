package migrations

import "embed"

// FS holds all embedded SQL migration scripts.
//
//go:embed *.sql
var FS embed.FS

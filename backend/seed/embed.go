package seed

import "embed"

// Files contains embedded SQL seed scripts.
//go:embed *.sql
var Files embed.FS

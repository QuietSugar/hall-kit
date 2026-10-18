// Package shell embeds the source-able shell assets and shell tools shipped
// with hall-kit.
package shell

import "embed"

// Files holds lib/, source/, bin/ and example.env as installed by
// "hall shell install" under ~/.hall/.
//
//go:embed lib source bin example.env
var Files embed.FS

package web

import "embed"

// FS holds all embedded web templates and static assets.
//
//go:embed templates/* static/css/* static/js/*
var FS embed.FS

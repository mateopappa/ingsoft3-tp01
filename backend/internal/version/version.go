package version

// Commit is the git SHA baked into the binary at build time (ldflags).
// Empty means "local / unknown" — useful for docker compose without build-args.
var Commit = "unknown"

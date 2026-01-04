package version

// Version is the application version, set at build time via ldflags
// Example: go build -ldflags "-X github.com/satetsu888/agentrace/server/internal/version.Version=v1.0.0"
var Version = "dev"

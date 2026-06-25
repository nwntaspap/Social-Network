package gates

import (
	"os"
	"path/filepath"
)

// isFeatureSlice checks if a directory under internalDir is a feature slice.
// It matches if the directory contains a `.feature-slice` marker file,
// OR if it is a directory under internalDir that is not in the skipDirs list.
func isFeatureSlice(internalDir, name string) bool {
	// Marker file takes precedence
	marker := filepath.Join(internalDir, name, ".feature-slice")
	if _, err := os.Stat(marker); err == nil {
		return true
	}

	// Restrict skipDirs check to top-level directories of internalDir
	if skipDirs[name] {
		return false
	}

	// By default, if it's a directory and not in skipDirs, treat it as a feature slice
	dirPath := filepath.Join(internalDir, name)
	info, err := os.Stat(dirPath)
	if err == nil && info.IsDir() {
		return true
	}

	return false
}

// NewDirs lists all the target new codebase directories.
var NewDirs = []string{
	"internal/user", "internal/follow", "internal/topic", "internal/comment",
	"internal/group", "internal/event", "internal/chat", "internal/notification",
	"internal/oauth", "internal/core", "internal/platform", "internal/bootstrap",
	"internal/config", "internal/gates", "cmd/gates", "cmd/server",
}

// NewPkgs lists all the target Go packages for the new codebase.
var NewPkgs = []string{
	"social-network/internal/user", "social-network/internal/follow", "social-network/internal/topic", "social-network/internal/comment",
	"social-network/internal/group", "social-network/internal/event", "social-network/internal/chat", "social-network/internal/notification",
	"social-network/internal/oauth", "social-network/internal/core", "social-network/internal/platform", "social-network/internal/bootstrap",
	"social-network/internal/config", "social-network/internal/gates", "social-network/cmd/gates", "social-network/cmd/server",
}

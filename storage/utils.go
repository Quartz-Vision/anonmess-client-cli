package storage

import (
	"path/filepath"
	"quartzvision/anonmess-client-cli/settings"
)

// Returns the path in context of the app data dir
func DataPath(path string) string {
	return filepath.Join(settings.Config.ProgramDataDir, path)
}

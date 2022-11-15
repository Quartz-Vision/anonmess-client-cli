package storage

import (
	"path/filepath"
	"quartzvision/anonmess-client-cli/settings"
)

// Returns the path in context of the app data dir
func DataPath(paths ...string) string {
	return filepath.Join(
		settings.Config.AppDataDirPath,
		filepath.Join(paths...),
	)
}

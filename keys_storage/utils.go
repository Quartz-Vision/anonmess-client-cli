package keysstorage

import (
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
	"strings"

	"github.com/google/uuid"
)

func keyFileName(prefixes ...string) (name string) {
	return strings.Join(prefixes, "-")
}

// Returns the path in context of the app data dir
func keyPath(packId uuid.UUID, prefixes ...string) string {
	return filestorage.DataPath(settings.Config.KeysStorageDefaultDirName, packId.String(), keyFileName(prefixes...))
}

func safeClose(obj Closable) {
	if obj != nil {
		obj.Close()
	}
}

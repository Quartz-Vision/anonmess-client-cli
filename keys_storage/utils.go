package keysstorage

import (
	"github.com/google/uuid"
	"path/filepath"
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
)

const PACK_PREFIX_IN = "in-"
const PACK_PREFIX_OUT = "out-"
const PACK_PREFIX_ID_KEY = "id-"
const PACK_PREFIX_PAYLOAD_KEY = "data-"

// Returns the path in context of the app data dir
func keyPath(packId uuid.UUID, prefixes ...string) string {
	prefix := filepath.Join(prefixes...)
	return storage.DataPath(settings.Config.KeysStorageDefaultDirName, prefix, packId.String())
}

func safeClose(obj Closable) {
	if obj != nil {
		obj.Close()
	}
}

package keysstorage

import (
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
	"strings"

	"github.com/google/uuid"
)

const PACK_PREFIX_IN = "in"
const PACK_PREFIX_OUT = "out"
const PACK_PREFIX_ID_KEY = "id"
const PACK_PREFIX_PAYLOAD_KEY = "data"

// Returns the path in context of the app data dir
func keyPath(packId uuid.UUID, prefixes ...string) string {
	return filestorage.DataPath(settings.Config.KeysStorageDefaultDirName, packId.String(), strings.Join(prefixes, "-"))
}

func safeClose(obj Closable) {
	if obj != nil {
		obj.Close()
	}
}

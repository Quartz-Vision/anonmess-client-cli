package keysstorage

import (
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
)

// Returns the path in context of the app data dir
func keyPath(keyId KeyId) string {
	return storage.DataPath(settings.Config.KeysStorageDefaultDirName, string(keyId))
}

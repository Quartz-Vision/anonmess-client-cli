package keysstorage

import (
	"os"
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
)

func Init() (err error) {
	return os.MkdirAll(
		storage.DataPath(settings.Config.KeysStorageDefaultDirName),
		os.ModePerm,
	)
}

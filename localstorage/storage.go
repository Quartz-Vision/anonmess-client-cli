package localstorage

import (
	"os"
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
)

func Init() (err error) {
	return os.MkdirAll(
		storage.DataPath(settings.Config.CacheStorageDefaultDirName),
		os.ModePerm,
	)
}

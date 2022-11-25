package keysstorage

import (
	"os"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
)

func Init() (err error) {
	return os.MkdirAll(
		filestorage.DataPath(settings.Config.KeysStorageDefaultDirName),
		os.ModePerm,
	)
}

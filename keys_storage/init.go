package keysstorage

import (
	"os"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() (err error) {
	return os.MkdirAll(
		utils.DataPath(settings.Config.KeysStorageDefaultDirName),
		os.ModePerm,
	)
}

package storage

import (
	"os"
	"quartzvision/anonmess-client-cli/settings"
)

func Init() (err error) {
	return os.MkdirAll(settings.Config.AppDataDirPath, os.ModePerm)
}

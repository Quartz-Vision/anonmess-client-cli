package storage

import (
	"os"
	"quartzvision/anonmess-client-cli/settings"
)

func Init() (err error) {
	if err := os.MkdirAll(settings.Config.ProgramDataDir, os.ModePerm); err != nil {
		return err
	}
	return nil
}

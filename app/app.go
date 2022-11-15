package app

import (
	"quartzvision/anonmess-client-cli/client"
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() error {
	return utils.UntilFirstError([]utils.ErrFn{
		settings.Init,
		storage.Init,
		client.Init,
	})
}

package app

import (
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/storage"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() error {
	return utils.UntilFirstError([]utils.ErrFn{
		settings.Init,
		storage.Init,
	})
}

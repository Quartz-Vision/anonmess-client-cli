package app

import (
	"quartzvision/anonmess-client-cli/cli"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/localstorage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() error {
	return utils.UntilFirstError([]utils.ErrFn{
		settings.Init,
		localstorage.Init,
		keysstorage.Init,
		cli.Init,
	})
}

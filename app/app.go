package app

import (
	"quartzvision/anonmess-client-cli/clientserver"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() error {
	return utils.UntilFirstError(
		keysstorage.Init,
		// cli.Init,
		clientserver.Init,
	)
}

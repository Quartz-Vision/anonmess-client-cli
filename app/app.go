package app

import (
	"quartzvision/anonmess-client-cli/clientserver"
	"quartzvision/anonmess-client-cli/filestorage"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
)

func Init() error {
	return utils.UntilFirstError(
		settings.Init,
		filestorage.InitFileManager,
		keysstorage.Init,
		// cli.Init,
		clientserver.Init,
	)
}

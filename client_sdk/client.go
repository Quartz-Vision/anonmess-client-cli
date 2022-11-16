package clientsdk

import (
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"

	"github.com/google/uuid"
)

func CreateChat() (chatId uuid.UUID, err error) {
	chatId = uuid.New()

	return chatId, keysstorage.ManageKeyPack(chatId)
}

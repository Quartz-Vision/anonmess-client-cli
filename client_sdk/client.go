package clientsdk

import (
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"

	"github.com/google/uuid"
)

type Chat struct {
	Id   uuid.UUID
	Name string
}

var Chats = map[uuid.UUID]*Chat{}

func CreateChat(name string) (chat *Chat, err error) {
	chat = &Chat{
		Id:   uuid.New(),
		Name: name,
	}

	err = keysstorage.ManageKeyPack(chat.Id)
	if err == nil {
		Chats[chat.Id] = chat
	}

	return chat, err
}

func UpdateChatsFromStorage() (err error) {
	return err
}

func ConnectChatFromStorage(chatId uuid.UUID, name string) (err error) {
	if _, ok := Chats[chatId]; ok {
		return nil
	}

	return err
}

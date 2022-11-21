package clientsdk

import (
	anoncastsdk "quartzvision/anonmess-client-cli/anoncast_sdk"
	"quartzvision/anonmess-client-cli/events"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"

	"github.com/google/uuid"
)

type Chat struct {
	Id   uuid.UUID
	Name string
}

type Message struct {
	Chat *Chat
	Text string
}

var Chats = map[uuid.UUID]*Chat{}

func ManageChat(chat *Chat) {
	Chats[chat.Id] = chat

	events.AddListener(chat.Id, anoncastsdk.EVENT_MESSAGE, func(e *events.Event) {
		events.EmitEvent(CLIENT_EVENTS_CHANNEL, &events.Event{
			Type: EVENT_CHAT_MESSAGE,
			Data: &Message{
				Chat: chat,
				Text: e.Data.(*anoncastsdk.Message).Text,
			},
		})
	})
}

func CreateChat(name string) (chat *Chat, err error) {
	chat = &Chat{
		Id:   uuid.New(),
		Name: name,
	}

	err = keysstorage.ManageKeyPack(chat.Id)
	if err == nil {
		ManageChat(chat)
	}

	return chat, err
}

func UpdateChatsFromStorage() (err error) {
	return err
}

func ManageChatFromStorage(chatId uuid.UUID, name string) (chat *Chat, err error) {
	if chat, ok := Chats[chatId]; ok {
		return chat, nil
	}

	if err := keysstorage.ManageKeyPack(chatId); err != nil {
		return nil, err
	}

	chat = &Chat{
		Id:   chatId,
		Name: name,
	}
	ManageChat(chat)
	return chat, nil
}

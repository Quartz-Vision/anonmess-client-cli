package clientsdk

import (
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"

	"github.com/google/uuid"
)

type Chat struct {
	Id   uuid.UUID
	Name string

	client *Client
}

func (c *Client) ManageChat(chat *Chat) {
	c.Chats[chat.Id] = chat
	chat.client = c

	c.anoncastClient.Listen(chat.Id, EVENT_RAW_MESSAGE, c.rawMessagesHandler)
}

func (c *Client) CreateChat(name string) (chat *Chat, err error) {
	chat = &Chat{
		Id:   uuid.New(),
		Name: name,
	}

	err = keysstorage.ManageKeyPack(chat.Id)
	if err == nil {
		c.ManageChat(chat)
	}

	return chat, err
}

func (c *Client) UpdateChatsFromStorage() (err error) {
	// if

	return err
}

func (c *Client) ManageChatFromStorage(chatId uuid.UUID, name string) (chat *Chat, err error) {
	if chat, ok := c.Chats[chatId]; ok {
		return chat, nil
	}

	if err := keysstorage.ManageKeyPack(chatId); err != nil {
		return nil, err
	}

	chat = &Chat{
		Id:   chatId,
		Name: name,
	}
	c.ManageChat(chat)
	return chat, nil
}

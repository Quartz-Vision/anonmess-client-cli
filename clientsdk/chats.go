package clientsdk

import (
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

type Chat struct {
	Id   uuid.UUID
	Name string

	client *Client
}

func (c *Client) ManageChat(chat *Chat) (err error) {
	err = utils.UntilFirstError(
		func() error { return keysstorage.ManageKeyPack(chat.Id) },
		func() error {
			return c.db.Update(func(txn *badger.Txn) error {
				e := badger.NewEntry(chat.Id[:], []byte(chat.Name))
				err := txn.SetEntry(e)
				return err
			})
		},
	)
	if err != nil {
		return err
	}

	c.Chats[chat.Id] = chat
	c.anoncastClient.Listen(chat.Id, EVENT_RAW_MESSAGE, c.rawMessagesHandler)

	return err
}

func (c *Client) CreateChat(name string) (chat *Chat, err error) {
	chat = &Chat{
		Id:   uuid.New(),
		Name: name,
	}

	return chat, c.ManageChat(chat)
}

func (c *Client) UpdateChatsList() (err error) {
	err = c.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			id, _ := uuid.FromBytes(item.Key())

			if _, ok := c.Chats[id]; ok {
				continue
			}

			err := item.Value(func(v []byte) error {
				return c.ManageChat(&Chat{
					Id:   id,
					Name: string(v),
				})
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
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

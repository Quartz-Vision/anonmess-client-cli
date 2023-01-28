package clientsdk

import (
	"path/filepath"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

type Chat struct {
	Id   uuid.UUID `json:"id"`
	Name string    `json:"name"`

	client *Client
}

// lets the chat know it belong to this client
// also adds the chat to the db and makes some preinitializations for it
// like setting up keys, listeners etc
func (c *Client) ManageChat(chat *Chat) (err error) {
	utils.UntilErrorPointer(
		&err,
		func() {
			_, err = keysstorage.ManageKeyPack(chat.Id)
			if err == keysstorage.ErrPackageExists {
				err = nil
			}
		},
		func() {
			err = c.db.Update(func(txn *badger.Txn) error {
				e := badger.NewEntry(chat.Id[:], []byte(chat.Name))
				err := txn.SetEntry(e)
				return err
			})
		},
		func() {
			chat.client = c
			c.Chats[chat.Id] = chat
			c.anoncastClient.Listen(chat.Id, EVENT_RAW_MESSAGE, c.rawMessagesHandler)
		},
	)

	return err
}

func (c *Client) CreateChat(name string) (chat *Chat, err error) {
	chat = &Chat{
		Id:   uuid.New(),
		Name: name,
	}

	err = c.ManageChat(chat)
	if err == nil {
		keys, _ := keysstorage.GetKeyPack(chat.Id)
		err = keys.GenerateKey(settings.Config.KeysStartSizeB)
	}
	return chat, err
}

// Imports keys, assigning them to a new chat with the name
func (c *Client) ImportSharedChat(src string, name string) (chat *Chat, err error) {
	keys, err := keysstorage.ManageSharedKeyPack(src)
	if err != nil {
		return nil, err
	}

	chat = &Chat{
		Id:   keys.PackId,
		Name: name,
	}

	return chat, c.ManageChat(chat)
}

// updates the chats list for the client from the DB
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

// exports the chat keys so that they can be imported
func (ch *Chat) ExportKeysForShare() (err error) {
	if keys, ok := keysstorage.GetKeyPack(ch.Id); ok {
		return keys.ExportShared(filepath.Join(settings.Config.AppDownloadsDirPath, ch.Id.String()))
	}
	return
}

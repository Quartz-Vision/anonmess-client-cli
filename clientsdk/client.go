package clientsdk

import (
	"quartzvision/anonmess-client-cli/anoncastsdk"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
)

type Client struct {
	events.EventManager

	ClientEventsChannel uuid.UUID
	Chats               map[uuid.UUID]*Chat

	anoncastClient *anoncastsdk.Client
	db             *badger.DB
}

func New() *Client {
	db, _ := badger.Open(badger.DefaultOptions(utils.DataPath("db")))

	client := &Client{
		ClientEventsChannel: uuid.New(),
		Chats:               map[uuid.UUID]*Chat{},
		EventManager:        *events.New(),
		anoncastClient:      anoncastsdk.New(),
		db:                  db,
	}

	client.initRawMessageEvent()
	client.anoncastClient.Pipe(
		client.anoncastClient.ClientEventsChannel,
		anoncastsdk.EVENT_ERROR,
		client,
		client.ClientEventsChannel,
		EVENT_ERROR,
	)
	client.anoncastClient.Pipe(
		client.anoncastClient.ClientEventsChannel,
		anoncastsdk.EVENT_CONNECTED,
		client,
		client.ClientEventsChannel,
		EVENT_CONNECTED,
	)
	// client.anoncastClient.ListenClient(anoncastsdk.EVENT_ERROR, client.anoncastErrorsHandler)

	return client
}

// For starting or restarting the connection
func (c *Client) StartConnection() (err error) {
	return c.anoncastClient.Start()
}

// Listen for messages in ClientEventsChannel (a public channel for api)
func (c *Client) ListenClient(etype events.EventType, handler events.EventHandlerFn) {
	c.Listen(c.ClientEventsChannel, etype, handler)
}

func (c *Client) Close() {
	c.anoncastClient.Stop()
	c.db.Close()
}

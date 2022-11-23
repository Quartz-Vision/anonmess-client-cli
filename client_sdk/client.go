package clientsdk

import (
	anoncastsdk "quartzvision/anonmess-client-cli/anoncast_sdk"
	"quartzvision/anonmess-client-cli/events"

	"github.com/google/uuid"
)

type Client struct {
	events.EventManager

	ClientEventsChannel uuid.UUID
	Chats               map[uuid.UUID]*Chat

	anoncastClient *anoncastsdk.Client
}

func New() *Client {
	client := &Client{
		ClientEventsChannel: uuid.New(),
		Chats:               map[uuid.UUID]*Chat{},
		EventManager:        *events.New(),
		anoncastClient:      anoncastsdk.New(),
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

func (c *Client) ListenClient(etype events.EventType, handler events.EventHandlerFn) {
	c.Listen(c.ClientEventsChannel, etype, handler)
}

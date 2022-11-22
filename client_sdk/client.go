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

	return client
}

func (c *Client) StartConnection() (err error) {
	return c.anoncastClient.Start()
}

func (c *Client) AddClientListener(etype events.EventType, handler events.EventHandlerFn) {
	c.AddListener(c.ClientEventsChannel, etype, handler)
}

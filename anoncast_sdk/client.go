package anoncastsdk

import (
	"io"
	"net"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/lists/squeue"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	events.EventManager

	eventsToSend *squeue.SQueue
}

func New() *Client {
	return &Client{
		EventManager: *events.New(),

		eventsToSend: squeue.New(),
	}
}

func (c *Client) SendEvent(channelId uuid.UUID, etype events.EventType, data any) {
	c.eventsToSend.Push(&dataPackage{
		channelId: channelId,
		event: c.Manage(&events.Event{
			Type: etype,
			Data: data,
		}),
		client: c,
	})
}

func (c *Client) Start() (err error) {
	conn, err := net.Dial("tcp", settings.Config.ServerAddr)

	if err != nil {
		return err
	}

	defer conn.Close()

	sizeRawBuf := make([]byte, utils.INT_MAX_SIZE)

	go (func() {
		for {
			for c.eventsToSend.IsEmpty() {
				time.Sleep(time.Millisecond)
			}

			for val, ok := c.eventsToSend.Pop(); ok; val, ok = c.eventsToSend.Pop() {
				pack := val.(*dataPackage)
				buf, _ := pack.MarshalBinary()
				conn.Write(buf)
			}
		}
	})()

	for {
		if _, err := io.ReadFull(conn, sizeRawBuf); err != nil {
			return err
		}

		packageSize, _ := utils.BytesToInt64(sizeRawBuf)
		packageBuf := make([]byte, packageSize+int64(len(sizeRawBuf)))

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(conn, packageBuf[len(sizeRawBuf):]); err != nil {
			return err
		}

		pack := dataPackage{client: c}
		pack.UnmarshalBinary(packageBuf)

		c.EmitEvent(pack.channelId, pack.event.Type, pack.event.Data)
	}
}

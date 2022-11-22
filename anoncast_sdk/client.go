package anoncastsdk

import (
	"io"
	"net"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/lists"
	"quartzvision/anonmess-client-cli/lists/squeue"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	EVENT_WARNING     = events.CreateEventType("warning")
	EVENT_FATAL_ERROR = events.CreateEventType("fatal_error")

	ERROR_NO_CONNECTION       = events.CreateEventType("no_connection")
	ERROR_BROKEN_PACKAGE_SEND = events.CreateEventType("broken_package")
	ERROR_BROKEN_PACKAGE_RECV = events.CreateEventType("broken_package_recv")
)

const (
	MAX_PACKAGE_SIZE_B = 1 << 20
)

type ClientErrorMessage struct {
	code          events.EventType
	details       string
	originalError error
}

type Client struct {
	events.EventManager

	ClientEventsChannel uuid.UUID
	eventsToSend        lists.Bidirectional
	conn                net.Conn
	sendPolling         bool
	mutex               sync.Mutex
}

func New() *Client {
	return &Client{
		EventManager: *events.New(),

		ClientEventsChannel: uuid.New(),
		eventsToSend:        squeue.New(),
		sendPolling:         false,
		mutex:               sync.Mutex{},
	}
}

func (c *Client) AddClientListener(etype events.EventType, handler events.EventHandlerFn) {
	c.AddListener(c.ClientEventsChannel, etype, handler)
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

// func (c)

func (c *Client) pollSendQueue() {
	for {
		for c.eventsToSend.IsEmpty() {
			time.Sleep(time.Millisecond)
		}

		for val, ok := c.eventsToSend.Pop(); ok; val, ok = c.eventsToSend.Pop() {
			pack := val.(*dataPackage)

			if buf, err := pack.MarshalBinary(); err != nil {
				c.EmitEvent(c.ClientEventsChannel, EVENT_WARNING, &ClientErrorMessage{
					code:          ERROR_BROKEN_PACKAGE_SEND,
					details:       "The client tried to send a broken package",
					originalError: err,
				})
				continue
			} else if _, err := c.conn.Write(buf); err != nil {
				c.noConnectionStop(err)
				c.eventsToSend.PushBack(pack)
				return
			}
		}
	}
}

func (c *Client) noConnectionStop(err error) {
	c.Stop()
	c.EmitEvent(c.ClientEventsChannel, EVENT_FATAL_ERROR, &ClientErrorMessage{
		code:          ERROR_NO_CONNECTION,
		details:       "Server connection failed",
		originalError: err,
	})
}

func (c *Client) Stop() {
	c.mutex.Lock()
	c.sendPolling = false

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.mutex.Unlock()
}

func (c *Client) Start() (err error) {
	c.mutex.Lock()
	if c.conn != nil {
		return nil
	}
	c.mutex.Unlock()

	defer c.Stop()

	c.conn, err = net.Dial("tcp", settings.Config.ServerAddr)
	if err != nil {
		c.noConnectionStop(err)
		return err
	}

	go c.pollSendQueue()

	sizeRawBuf := make([]byte, utils.INT_MAX_SIZE)

	for {
		if _, err := io.ReadFull(c.conn, sizeRawBuf); err != nil {
			c.noConnectionStop(err)
			return err
		}

		packageSize, _ := utils.BytesToInt64(sizeRawBuf)
		if packageSize <= 0 || packageSize >= MAX_PACKAGE_SIZE_B {
			c.EmitEvent(c.ClientEventsChannel, EVENT_WARNING, &ClientErrorMessage{
				code:    ERROR_BROKEN_PACKAGE_RECV,
				details: "The client got a broken package",
			})
			continue
		}
		packageBuf := make([]byte, packageSize+int64(len(sizeRawBuf)))

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(c.conn, packageBuf[len(sizeRawBuf):]); err != nil {
			c.noConnectionStop(err)
			return err
		}

		pack := dataPackage{event: c.Manage(&events.Event{})}
		if err := pack.UnmarshalBinary(packageBuf); err != nil {
			c.EmitEvent(c.ClientEventsChannel, EVENT_WARNING, &ClientErrorMessage{
				code:    ERROR_BROKEN_PACKAGE_RECV,
				details: "The client got a broken package",
			})
			continue
		}

		c.EmitEvent(pack.channelId, pack.event.Type, pack.event.Data)
	}
}

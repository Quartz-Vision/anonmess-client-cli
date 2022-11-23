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
	EVENT_ERROR     = events.CreateEventType("client_error")
	EVENT_CONNECTED = events.CreateEventType("client_connected")

	ERROR_FATAL               = events.CreateEventType("fatal_error") // Only when connection fails or being dropped
	ERROR_BROKEN_PACKAGE_SEND = events.CreateEventType("broken_package")
	ERROR_BROKEN_PACKAGE_RECV = events.CreateEventType("broken_package_recv")
)

const (
	MAX_PACKAGE_SIZE_B = 1 << 20
)

type ClientErrorMessage struct {
	Code          events.EventType
	Details       string
	OriginalError error
}

type Client struct {
	events.EventManager

	ClientEventsChannel uuid.UUID
	eventsToSend        lists.Bidirectional
	conn                net.Conn
	mutex               sync.Mutex
	lastError           error
}

func New() *Client {
	return &Client{
		EventManager: *events.New(),

		ClientEventsChannel: uuid.New(),
		eventsToSend:        squeue.New(),
		mutex:               sync.Mutex{},
	}
}

func (c *Client) ListenClient(etype events.EventType, handler events.EventHandlerFn) {
	c.Listen(c.ClientEventsChannel, etype, handler)
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

func (c *Client) pollSendQueue(wg *sync.WaitGroup) {
root:
	for c.conn != nil {
		for c.conn != nil && c.eventsToSend.IsEmpty() {
			time.Sleep(time.Millisecond)
		}

		for c.conn != nil {
			val, ok := c.eventsToSend.Pop()
			if !ok {
				break
			}
			pack := val.(*dataPackage)

			if buf, err := pack.MarshalBinary(); err != nil {
				c.emitError(ERROR_BROKEN_PACKAGE_SEND, "The client tried to send a broken package", err)
				continue
			} else if _, err := c.conn.Write(buf); err != nil {
				c.lastError = err
				c.eventsToSend.PushBack(pack)
				break root
			}
		}
	}

	wg.Done()
}

func (c *Client) emitError(code events.EventType, details string, origin error) {
	c.Emit(c.ClientEventsChannel, EVENT_ERROR, &ClientErrorMessage{
		Code:          code,
		Details:       details,
		OriginalError: origin,
	})
}

func (c *Client) Stop() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

func (c *Client) Start() (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.conn, err = net.Dial("tcp", settings.Config.ServerAddr)
	if err != nil {
		c.emitError(ERROR_FATAL, "Can't connect to the server", err)
		return err
	}
	c.Emit(c.ClientEventsChannel, EVENT_CONNECTED, nil)

	pollingWaitGroup := sync.WaitGroup{}
	pollingWaitGroup.Add(1)
	go c.pollSendQueue(&pollingWaitGroup)

	sizeRawBuf := make([]byte, utils.INT_MAX_SIZE)

	for c.conn != nil {
		if _, err := io.ReadFull(c.conn, sizeRawBuf); err != nil {
			c.lastError = err
			break
		}

		packageSize, _ := utils.BytesToInt64(sizeRawBuf)
		if packageSize <= 0 || packageSize >= MAX_PACKAGE_SIZE_B {
			c.emitError(ERROR_BROKEN_PACKAGE_RECV, "The client got a broken package", err)
			continue
		}
		packageBuf := make([]byte, packageSize+int64(len(sizeRawBuf)))

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(c.conn, packageBuf[len(sizeRawBuf):]); err != nil {
			c.lastError = err
			break
		}

		pack := dataPackage{event: c.Manage(&events.Event{})}
		if err := pack.UnmarshalBinary(packageBuf); err != nil {
			c.emitError(ERROR_BROKEN_PACKAGE_RECV, "The client got a broken package", err)
			continue
		}

		c.Emit(pack.channelId, pack.event.Type, pack.event.Data)
	}

	c.Stop()
	pollingWaitGroup.Wait()

	if c.lastError != nil {
		c.emitError(ERROR_FATAL, "The client's been disconnected", err)
	}
	return c.lastError
}

package anoncastsdk

import (
	"errors"
	"io"
	"net"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
	"sync"

	"github.com/google/uuid"
)

const (
	MAX_PACKAGE_SIZE_B = 1 << 20
)

var ErrConnectionClosed = errors.New("connection already closed")
var ErrConnectionFailed = errors.New("connection failed")
var ErrBrokenPackageSend = errors.New("attempted to send a broken package")
var ErrBrokenPackageRecv = errors.New("received a broken package. Dropping connection")

type ClientErrorMessage struct {
	Code          events.EventType
	Details       string
	OriginalError error
}

type Client struct {
	conn  net.Conn
	mutex sync.Mutex
}

func New() *Client {
	return &Client{
		mutex: sync.Mutex{},
	}
}

// Adds new events to the sending queue. They will be sent to the server
func (c *Client) Write(channelId uuid.UUID, payload []byte) error {
	if c.conn != nil {
		pack := DataPackage{
			ChannelId: channelId,
			Payload:   payload,
		}

		if buf, err := pack.MarshalBinary(); err != nil {
			return ErrBrokenPackageSend
		} else if _, err := c.conn.Write(buf); err != nil {
			return ErrConnectionFailed
		}
	}

	return nil
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
		return ErrConnectionFailed
	}

	return nil
}

func (c *Client) Receive() (pack *DataPackage, err error) {
	sizeRawBuf := make([]byte, utils.INT_MAX_SIZE)

	for c.conn != nil {
		if _, err := io.ReadFull(c.conn, sizeRawBuf); err != nil {
			return nil, ErrConnectionFailed
		}

		// decode package size
		packageSize, _ := utils.BytesToInt64(sizeRawBuf)
		if packageSize <= 0 || packageSize >= MAX_PACKAGE_SIZE_B {
			return nil, ErrBrokenPackageRecv
		}

		// make a buffer with the whole package
		packageBuf := make([]byte, packageSize+int64(len(sizeRawBuf)))
		copy(packageBuf, sizeRawBuf)
		if _, err := io.ReadFull(c.conn, packageBuf[len(sizeRawBuf):]); err != nil {
			return nil, ErrConnectionFailed
		}

		// decode the package
		pack := DataPackage{}
		if err := pack.UnmarshalBinary(packageBuf); err == nil {
			return &pack, nil
		} else if err != ErrKeyPackIdDecodeFailed {
			return nil, ErrBrokenPackageRecv
		}
	}

	return nil, ErrConnectionClosed
}

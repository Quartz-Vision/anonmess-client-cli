package anoncastsdk

import (
	"io"
	"net"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"
	"time"
)

func Init() (err error) {
	initEvents()

	return nil
}

func Start() (err error) {
	conn, err := net.Dial("tcp", settings.Config.ServerAddr)

	if err != nil {
		return err
	}

	defer conn.Close()

	sizeRawBuf := make([]byte, utils.INT_MAX_SIZE)

	go (func() {
		for {
			for eventsToSend.IsEmpty() {
				time.Sleep(time.Millisecond)
			}

			for val, ok := eventsToSend.Pop(); ok; val, ok = eventsToSend.Pop() {
				event := val.(eventPack)
				pack := Package{
					ChannelId: event.channelId,
					Event:     event.event,
				}

				buf, _ := pack.MarshalBinary()

				p := Package{}
				p.UnmarshalBinary(buf)

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

		pack := Package{}
		pack.UnmarshalBinary(packageBuf)

		events.EmitEvent(pack.ChannelId, pack.Event)
	}
}

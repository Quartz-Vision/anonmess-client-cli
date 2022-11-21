package anoncastsdk

import (
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/utils"
)

var EVENT_MESSAGE = events.CreateEventType("message")

type Message struct {
	Text string
}

func initMessageEvent() {
	events.SetEventCoders(
		EVENT_MESSAGE,
		func(e *events.Event) (dst []byte, err error) {
			data := e.Data.(*Message)

			encodedTextLen := utils.Int64ToBytes(int64(len(data.Text)))
			encodedData := make([]byte, len(encodedTextLen)+len(data.Text))

			copy(encodedData, encodedTextLen)
			copy(encodedData[len(encodedTextLen):], []byte(data.Text))

			return encodedData, nil
		},
		func(e *events.Event, src []byte) (err error) {
			textLen, sizeRead := utils.BytesToInt64(src)

			e.Data = &Message{
				Text: string(src[sizeRead : int64(sizeRead)+textLen]),
			}
			return nil
		},
	)
}

package anoncastsdk

type Message struct {
	Text string
}

func initMessageEvent(eventType EventType) {
	EventSerializers[eventType] = EventSerializer{
		Encode: func(e *Event) (dst []byte, err error) {
			data := e.Data.(*Message)

			encodedTextLen := Int64ToBytes(int64(len(data.Text)))
			encodedData := make([]byte, len(encodedTextLen)+len(data.Text))

			copy(encodedData, encodedTextLen)
			copy(encodedData[len(encodedTextLen):], []byte(data.Text))

			return encodedData, nil
		},

		Decode: func(e *Event, src []byte) (err error) {
			textLen, sizeRead := BytesToInt64(src)

			e.Data = &Message{
				Text: string(src[sizeRead : int64(sizeRead)+textLen]),
			}
			return nil
		},
	}
}

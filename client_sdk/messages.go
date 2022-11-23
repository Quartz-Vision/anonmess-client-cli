package clientsdk

import (
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/utils"
)

var (
	EVENT_RAW_MESSAGE  = events.CreateEventType("raw_message")
	EVENT_CHAT_MESSAGE = events.CreateEventType("chat_message")
)

type RawMessage struct {
	Text string
}

// Implements encoding.BinaryMarshaler
func (r *RawMessage) MarshalBinary() (data []byte, err error) {
	textSize := len(r.Text)
	textSizeEnc := utils.Int64ToBytes(int64(textSize))

	data = make([]byte, len(textSizeEnc)+textSize)
	utils.JoinSlices(data, textSizeEnc, []byte(r.Text))

	return data, nil
}

// Implements encoding.BinaryUnmarshaler
func (r *RawMessage) UnmarshalBinary(data []byte) (err error) {
	textSize, textSizeLen := utils.BytesToInt64(data[:8])
	r.Text = string(data[textSizeLen : int64(textSizeLen)+textSize])
	return nil
}

type ChatMessage struct {
	Chat *Chat
	Text string
}

type MessageHandlerFn func(msg *ChatMessage)

// Wraps handlers for EVENT_CHAT_MESSAGE
func (c *Client) WrapMessageHandler(f MessageHandlerFn) events.EventHandlerFn {
	return func(e *events.Event) {
		f(e.Data.(*ChatMessage))
	}
}

// Creates Event coders for EVENT_RAW_MESSAGE
func (c *Client) initRawMessageEvent() {
	c.anoncastClient.SetEventCoders(
		EVENT_RAW_MESSAGE,
		func(e *events.Event) (dst []byte, err error) {
			return e.Data.(*RawMessage).MarshalBinary()
		},
		func(e *events.Event, src []byte) (err error) {
			msg := &RawMessage{}
			err = msg.UnmarshalBinary(src)
			e.Data = msg
			return err
		},
	)
}

// Sends messages to the main client channel, changing them to ChatMessage
func (c *Client) rawMessagesHandler(e *events.Event) {
	c.Emit(c.ClientEventsChannel, EVENT_CHAT_MESSAGE, &ChatMessage{
		Chat: c.Chats[e.ChannelId],
		Text: e.Data.(*RawMessage).Text,
	})
}

func (ch *Chat) SendMessage(text string) {
	ch.client.anoncastClient.SendEvent(ch.Id, EVENT_RAW_MESSAGE, &RawMessage{
		Text: text,
	})
}

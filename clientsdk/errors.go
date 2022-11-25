package clientsdk

import (
	anoncastsdk "quartzvision/anonmess-client-cli/anoncast_sdk"
	"quartzvision/anonmess-client-cli/events"
)

var (
	EVENT_ERROR = events.CreateEventType("client_error")
)

type ErrorHandlerFn func(msg *anoncastsdk.ClientErrorMessage)

// Wraps handlers for EVENT_ERROR
func (c *Client) WrapErrorHandler(f ErrorHandlerFn) events.EventHandlerFn {
	return func(e *events.Event) {
		f(e.Data.(*anoncastsdk.ClientErrorMessage))
	}
}

// func (c *Client) anoncastErrorsHandler(e *events.Event) {
// 	c.Emit(c.ClientEventsChannel, EVENT_CHAT_MESSAGE, &ChatMessage{
// 		Chat: c.Chats[e.ChannelId],
// 		Text: e.Data.(*RawMessage).Text,
// 	})
// }

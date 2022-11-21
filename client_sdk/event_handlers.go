package clientsdk

import (
	"quartzvision/anonmess-client-cli/events"

	"github.com/google/uuid"
)

type MessageHandlerFn func(msg *Message)

var (
	CLIENT_EVENTS_CHANNEL = uuid.New()
	EVENT_CHAT_MESSAGE    = events.CreateEventType("chat_message")
)

// Wraps handlers for EVENT_CHAT_MESSAGE
func WrapMessageHandler(f MessageHandlerFn) events.EventHandlerFn {
	return func(e *events.Event) {
		f(e.Data.(*Message))
	}
}

func AddClientListener(etype events.EventType, handler events.EventHandlerFn) {
	events.AddListener(CLIENT_EVENTS_CHANNEL, etype, handler)
}

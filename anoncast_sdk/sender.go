package anoncastsdk

import (
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/lists/squeue"

	"github.com/google/uuid"
)

type eventPack struct {
	channelId uuid.UUID
	event     *events.Event
}

var eventsToSend = squeue.New()

func SendEvent(channelId uuid.UUID, e *events.Event) {
	eventsToSend.Push(eventPack{
		channelId: channelId,
		event:     e,
	})
}

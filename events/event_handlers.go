package events

import (
	"github.com/google/uuid"
)

type EventHandlerFn func(e *Event)

type handlersMap map[EventType]EventHandlerFn

var eventHandlers map[uuid.UUID]*handlersMap = map[uuid.UUID]*handlersMap{}

func AddListener(channelId uuid.UUID, etype EventType, handler EventHandlerFn) {
	channelHandlers, ok := eventHandlers[channelId]
	if !ok {
		channelHandlers = &handlersMap{}
		eventHandlers[channelId] = channelHandlers
	}

	(*channelHandlers)[etype] = handler
}

func EmitEvent(channelId uuid.UUID, e *Event) {
	if channelHandlers, ok := eventHandlers[channelId]; ok {
		if handler, ok := (*channelHandlers)[e.Type]; ok {
			handler(e)
		}
	}
}

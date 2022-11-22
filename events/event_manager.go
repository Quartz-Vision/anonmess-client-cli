package events

import (
	"hash/crc64"

	"github.com/google/uuid"
)

type EventType uint64

var eventTypeCrcTable = crc64.MakeTable(crc64.ISO)

func CreateEventType(name string) EventType {
	return EventType(crc64.Checksum([]byte(name), eventTypeCrcTable))
}

type EventMarshalFn func(e *Event) (dst []byte, err error)
type EventUnmarshalFn func(e *Event, src []byte) (err error)
type eventSerializer struct {
	encode EventMarshalFn
	decode EventUnmarshalFn
}
type EventHandlerFn func(e *Event)
type handlersMap map[EventType]EventHandlerFn

type EventManager struct {
	eventSerializers map[EventType]*eventSerializer
	eventHandlers    map[uuid.UUID]*handlersMap
}

func New() *EventManager {
	return &EventManager{
		eventSerializers: map[EventType]*eventSerializer{},
		eventHandlers:    map[uuid.UUID]*handlersMap{},
	}
}

// Set encode/decode functions for the event type
func (m *EventManager) SetEventCoders(etype EventType, encoder EventMarshalFn, decoder EventUnmarshalFn) {
	m.eventSerializers[etype] = &eventSerializer{
		encode: encoder,
		decode: decoder,
	}
}

func (m *EventManager) AddListener(channelId uuid.UUID, etype EventType, handler EventHandlerFn) {
	channelHandlers, ok := m.eventHandlers[channelId]
	if !ok {
		channelHandlers = &handlersMap{}
		m.eventHandlers[channelId] = channelHandlers
	}

	(*channelHandlers)[etype] = handler
}

func (m *EventManager) EmitEvent(channelId uuid.UUID, etype EventType, data any) {
	if channelHandlers, ok := m.eventHandlers[channelId]; ok {
		if handler, ok := (*channelHandlers)[etype]; ok {
			handler(m.Manage(&Event{
				Type:      etype,
				ChannelId: channelId,
				Data:      data,
			}))
		}
	}
}

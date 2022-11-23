package events

import (
	"hash/crc64"
	"sync"

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
type handlersMap map[EventType]*map[uuid.UUID]EventHandlerFn

type IEventManaget interface {
	Listen(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID)
	Unlisten(channelId uuid.UUID, etype EventType, listenerId uuid.UUID)
	// Once(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID)
	Emit(channelId uuid.UUID, etype EventType, data any)
	Pipe(
		srcChannel uuid.UUID,
		srcEtype EventType,
		dstManager IEventManaget,
		dstChannel uuid.UUID,
		dstEtype EventType,
	)
}
type EventManager struct {
	IEventManaget

	eventSerializers map[EventType]*eventSerializer
	eventHandlers    map[uuid.UUID]*handlersMap
	mutex            sync.Mutex
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

func (m *EventManager) Listen(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID) {
	listenerId = uuid.New()

	m.mutex.Lock()
	channelHandlers, ok := m.eventHandlers[channelId]
	if !ok {
		channelHandlers = &handlersMap{}
		m.eventHandlers[channelId] = channelHandlers
	}

	eventHandlers, ok := (*channelHandlers)[etype]
	if !ok {
		eventHandlers = &map[uuid.UUID]EventHandlerFn{}
		(*channelHandlers)[etype] = eventHandlers
	}
	m.mutex.Unlock()

	(*eventHandlers)[listenerId] = handler
	return listenerId
}

func (m *EventManager) Unlisten(channelId uuid.UUID, etype EventType, listenerId uuid.UUID) {
	if channelHandlers, ok := m.eventHandlers[channelId]; ok {
		if eventHandlers, ok := (*channelHandlers)[etype]; !ok {
			delete(*eventHandlers, listenerId)
		}
	}
}

// func (m *EventManager) Once(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID) {
// 	m.mutex.Lock()
// 	listenerId := uuid.New()
// 	channelHandlers, ok := m.eventHandlers[channelId]
// 	if !ok {
// 		channelHandlers = &handlersMap{}
// 		m.eventHandlers[channelId] = channelHandlers
// 	}

// 	(*channelHandlers)[etype][listenerId] = handler
// 	m.mutex.Unlock()
// }

func (m *EventManager) Emit(channelId uuid.UUID, etype EventType, data any) {
	if channelHandlers, ok := m.eventHandlers[channelId]; ok {
		if eventHandlers, ok := (*channelHandlers)[etype]; ok {
			for _, handler := range *eventHandlers {
				go handler(m.Manage(&Event{
					Type:      etype,
					ChannelId: channelId,
					Data:      data,
				}))
			}
		}
	}
}

func (m *EventManager) Pipe(
	srcChannel uuid.UUID,
	srcEtype EventType,
	dstManager IEventManaget,
	dstChannel uuid.UUID,
	dstEtype EventType,
) {
	m.Listen(srcChannel, srcEtype, func(e *Event) {
		dstManager.Emit(dstChannel, dstEtype, e.Data)
	})
}

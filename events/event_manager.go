package events

import (
	"hash/crc64"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map/v2"
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
type handlersCollection map[uuid.UUID]EventHandlerFn
type eventHandlers map[EventType]*handlersCollection

type IEventManager interface {
	Listen(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID)
	// Once(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID)
	Unlisten(channelId uuid.UUID, etype EventType, listenerId uuid.UUID)
	Emit(channelId uuid.UUID, etype EventType, data any)
	Pipe(
		srcChannel uuid.UUID,
		srcEtype EventType,
		dstManager IEventManager,
		dstChannel uuid.UUID,
		dstEtype EventType,
	) (listenerId uuid.UUID)
}
type EventManager struct {
	IEventManager

	eventSerializers map[EventType]*eventSerializer
	eventHandlers    cmap.ConcurrentMap[uuid.UUID, *eventHandlers]
}

func New() *EventManager {
	return &EventManager{
		eventSerializers: map[EventType]*eventSerializer{},
		eventHandlers:    NewUUIDCmap[*eventHandlers](),
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

	go m.eventHandlers.Upsert(channelId, nil, func(exist bool, valueInMap, _ *eventHandlers) *eventHandlers {
		if !exist {
			valueInMap = &eventHandlers{}
		}
		handlers, ok := (*valueInMap)[etype]
		if !ok {
			handlers = &handlersCollection{}
			(*valueInMap)[etype] = handlers
		}
		(*handlers)[listenerId] = handler

		return valueInMap
	})

	return listenerId
}

func (m *EventManager) Unlisten(channelId uuid.UUID, etype EventType, listenerId uuid.UUID) {
	go m.eventHandlers.RemoveCb(channelId, func(key uuid.UUID, v *eventHandlers, exists bool) bool {
		if exists {
			if handlers, ok := (*v)[etype]; !ok {
				delete(*handlers, listenerId)

				if len(*handlers) == 0 {
					delete(*v, etype)
				}
			}
		}
		return len(*v) == 0
	})
}

// func (m *EventManager) Once(channelId uuid.UUID, etype EventType, handler EventHandlerFn) (listenerId uuid.UUID) {
// 	listenerId = uuid.New()
// 	executed := atomic.Bool{}

// 	m.mutex.Lock()
// 	channelHandlers, ok := m.eventHandlers[channelId]
// 	if !ok {
// 		channelHandlers = &handlersMap{}
// 		m.eventHandlers[channelId] = channelHandlers
// 	}

// 	eventHandlers, ok := (*channelHandlers)[etype]
// 	if !ok {
// 		eventHandlers = &map[uuid.UUID]EventHandlerFn{}
// 		(*channelHandlers)[etype] = eventHandlers
// 	}

// 	(*eventHandlers)[listenerId] = func(e *Event) {
// 		if !executed.Swap(true) {
// 			m.Unlisten(channelId, etype, listenerId)
// 			handler(e)
// 		}
// 	}
// 	m.mutex.Unlock()
// 	return listenerId
// }

func (m *EventManager) Emit(channelId uuid.UUID, etype EventType, data any) {
	safeHandlers := []EventHandlerFn{}

	m.eventHandlers.RemoveCb(channelId, func(key uuid.UUID, v *eventHandlers, exists bool) bool {
		if exists {
			if handlers, ok := (*v)[etype]; ok {
				safeHandlers = make([]EventHandlerFn, 0, len(*handlers))
				for _, handler := range *handlers {
					safeHandlers = append(safeHandlers, handler)
				}
			}
		}
		return false
	})

	for i := range safeHandlers {
		go safeHandlers[i](&Event{
			ChannelId: channelId,
			Type:      etype,
			Data:      data,
			manager:   m,
		})
	}
}

func (m *EventManager) Pipe(
	srcChannel uuid.UUID,
	srcEtype EventType,
	dstManager IEventManager,
	dstChannel uuid.UUID,
	dstEtype EventType,
) (listenerId uuid.UUID) {
	return m.Listen(srcChannel, srcEtype, func(e *Event) {
		dstManager.Emit(dstChannel, dstEtype, e.Data)
	})
}

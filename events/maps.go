package events

import (
	"unsafe"

	"github.com/google/uuid"
	cmap "github.com/orcaman/concurrent-map/v2"
)

func NewUUIDCmap[V any]() cmap.ConcurrentMap[uuid.UUID, V] {
	return cmap.NewWithCustomShardingFunction[uuid.UUID, V](UUID32)
}

func UUID32(key uuid.UUID) uint32 {
	return *(*uint32)(unsafe.Pointer(&key[0]))
}

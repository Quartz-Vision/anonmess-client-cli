package storage

import (
	"errors"
)

var ErrStorageClosed = errors.New("operation is impossible, storage is closed")

type Storage interface {
	Close()
	Open() (err error)
	ReadChunk(destBuf []byte, start int64) (err error)
	Append(data []byte) (pos int64, err error)
	Len() (length int64, err error)
}

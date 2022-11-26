package filestorage

import (
	"errors"
)

var ErrStorageClosed = errors.New("operation is impossible, storage is closed")

type File interface {
	ReadAt(b []byte, offset int64) (nRead int64, err error)
	Seek(offset int64, whence int) (ret int64, err error)
	Write(b []byte) (nWritten int64, err error)
	WriteAt(b []byte, offset int64) (nWritten int64, err error)
	Append(data []byte) (pos int64, err error)
	Size() (length int64, err error)
	Close()
}

package filestorage

import (
	"os"
	"path"
)

type RawFsStorage struct {
	FilePath string
	Opened   bool
	file     *File
}

func NewRawFsStorage(filePath string) (storage *RawFsStorage) {
	return &RawFsStorage{
		FilePath: filePath,
	}
}

func (obj *RawFsStorage) Open() (err error) {
	if obj.Opened {
		return nil
	}

	if err := os.MkdirAll(path.Dir(obj.FilePath), os.ModePerm); err != nil {
		return err
	}
	obj.file = OpenFile(obj.FilePath, os.O_RDWR|os.O_CREATE, 0o600)
	obj.Opened = true

	return nil
}

func (obj *RawFsStorage) Close() {
	if obj.Opened {
		obj.file.Close()
		obj.Opened = false
	}
}

func (obj *RawFsStorage) ReadChunk(destBuf []byte, start int64) (err error) {
	if !obj.Opened {
		return ErrStorageClosed
	}

	_, err = obj.file.ReadAt(destBuf, start)
	return err
}

func (obj *RawFsStorage) Append(data []byte) (pos int64, err error) {
	if !obj.Opened {
		return 0, ErrStorageClosed
	}

	writePos, err := obj.file.Seek(0, os.SEEK_END)
	if err != nil {
		return 0, err
	}

	_, err = obj.file.Write(data)

	return writePos, err
}

func (obj *RawFsStorage) Len() (length int64, err error) {
	return obj.file.Seek(0, os.SEEK_END)
}

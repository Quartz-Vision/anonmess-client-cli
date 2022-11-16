package keysstorage

import (
	"io"
	"quartzvision/anonmess-client-cli/crypto/random"
	storage "quartzvision/anonmess-client-cli/file_storage"
	"quartzvision/anonmess-client-cli/settings"
)

const POS_REWIND = -1

type KeyBuffer struct {
	storage        storage.Storage
	currentPostion int64
	buf            []byte
	KeyLength      int64
}

func NewKeyBuffer(path string) (obj *KeyBuffer, err error) {
	obj = &KeyBuffer{
		storage:        storage.NewRawFsStorage(path),
		currentPostion: POS_REWIND,
		buf:            make([]byte, settings.Config.KeysBufferSizeB),
	}

	if err := obj.storage.Open(); err != nil {
		return nil, err
	}
	err = obj.UpdateLengthFromFile()
	return obj, err
}

func (obj *KeyBuffer) GetKeySlice(pos int64, length int64) (data []byte, err error) {
	relativePos := pos - obj.currentPostion
	endPos := pos + length

	if endPos > obj.KeyLength {
		return nil, ErrOutOfBound
	}

	if obj.currentPostion != POS_REWIND && relativePos >= 0 && length <= (settings.Config.KeysBufferSizeB-relativePos) {
		return obj.buf[relativePos : relativePos+length], nil
	}

	err = obj.storage.ReadChunk(obj.buf, pos)
	if err == io.EOF {
		err = nil
	}

	return obj.buf[:length], err
}

func (obj *KeyBuffer) AppendKey(data []byte) (pos int64, err error) {
	pos, err = obj.storage.Append(data)

	if err != nil {
		return pos, err
	}

	if (obj.currentPostion + settings.Config.KeysBufferSizeB) > obj.KeyLength {
		// reset the buffer if its current value overlaps with appended data
		obj.currentPostion = POS_REWIND
	}
	obj.KeyLength += int64(len(data))

	return pos, err
}

func (obj *KeyBuffer) UpdateLengthFromFile() (err error) {
	obj.KeyLength, err = obj.storage.Len()
	return err
}

func (obj *KeyBuffer) GenerateKey(length int64) (err error) {
	if rest := length % settings.Config.KeysBufferSizeB; rest != 0 {
		if key, err := random.GenerateRandomBytes(rest); err != nil {
			return err
		} else {
			obj.AppendKey(key)
		}
	}

	length = length / settings.Config.KeysBufferSizeB

	for i := int64(0); i < length; i++ {
		if key, err := random.GenerateRandomBytes(settings.Config.KeysBufferSizeB); err != nil {
			return err
		} else {
			obj.AppendKey(key)
		}
	}

	return nil
}

func (obj *KeyBuffer) Close() {
	obj.storage.Close()
}

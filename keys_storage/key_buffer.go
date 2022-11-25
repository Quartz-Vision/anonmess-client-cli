package keysstorage

import (
	"io"
	"quartzvision/anonmess-client-cli/crypto/random"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
)

const POS_REWIND = -1

type KeyBuffer struct {
	storage        filestorage.Storage
	currentPostion int64
	buf            []byte
	KeyLength      int64
}

func NewKeyBuffer(path string) (b *KeyBuffer, err error) {
	b = &KeyBuffer{
		storage:        filestorage.NewRawFsStorage(path),
		currentPostion: POS_REWIND,
		buf:            make([]byte, settings.Config.KeysBufferSizeB),
	}

	if err := b.storage.Open(); err != nil {
		return nil, err
	}
	err = b.UpdateLengthFromFile()
	return b, err
}

func (b *KeyBuffer) GetKeySlice(pos int64, length int64) (data []byte, err error) {
	relativePos := pos - b.currentPostion
	endPos := pos + length

	if endPos > b.KeyLength {
		return nil, ErrOutOfBound
	}

	if b.currentPostion != POS_REWIND && relativePos >= 0 && length <= (settings.Config.KeysBufferSizeB-relativePos) {
		return b.buf[relativePos : relativePos+length], nil
	}

	err = b.storage.ReadChunk(b.buf, pos)
	if err == io.EOF {
		err = nil
	}

	return b.buf[:length], err
}

func (b *KeyBuffer) AppendKey(data []byte) (pos int64, err error) {
	pos, err = b.storage.Append(data)

	if err != nil {
		return pos, err
	}

	if (b.currentPostion + settings.Config.KeysBufferSizeB) > b.KeyLength {
		// reset the buffer if its current value overlaps with appended data
		b.currentPostion = POS_REWIND
	}
	b.KeyLength += int64(len(data))

	return pos, err
}

func (b *KeyBuffer) UpdateLengthFromFile() (err error) {
	b.KeyLength, err = b.storage.Len()
	return err
}

func (b *KeyBuffer) GenerateKey(length int64) (err error) {
	if rest := length % settings.Config.KeysBufferSizeB; rest != 0 {
		if key, err := random.GenerateRandomBytes(rest); err != nil {
			return err
		} else {
			b.AppendKey(key)
		}
	}

	length = length / settings.Config.KeysBufferSizeB

	for i := int64(0); i < length; i++ {
		if key, err := random.GenerateRandomBytes(settings.Config.KeysBufferSizeB); err != nil {
			return err
		} else {
			b.AppendKey(key)
		}
	}

	return nil
}

func (b *KeyBuffer) Close() {
	b.storage.Close()
}

package keysstorage

import (
	"os"
	"quartzvision/anonmess-client-cli/crypto/random"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
)

const POS_REWIND = -1

type KeyBuffer struct {
	*filestorage.BufferedFile
}

func NewKeyBuffer(path string) (b *KeyBuffer, err error) {
	file, err := filestorage.NewFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	bf, err := filestorage.NewBufferedFile(file, settings.Config.KeysBufferSizeB)

	b = &KeyBuffer{
		BufferedFile: bf,
	}
	return b, err
}

func (b *KeyBuffer) GenerateKey(length int64) (err error) {
	if rest := length % b.BufferSize; rest != 0 {
		if key, err := random.GenerateRandomBytes(rest); err != nil {
			return err
		} else if _, err := b.Append(key); err != nil {
			return err
		}
	}

	length = length / b.BufferSize

	for i := int64(0); i < length; i++ {
		if key, err := random.GenerateRandomBytes(b.BufferSize); err != nil {
			return err
		} else if _, err := b.Append(key); err != nil {
			return err
		}
	}

	return nil
}

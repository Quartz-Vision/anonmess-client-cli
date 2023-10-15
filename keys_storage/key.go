package keysstorage

import (
	"os"
	"path/filepath"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/Quartz-Vision/gocrypt/random"
	"github.com/Quartz-Vision/gofile"

	"github.com/google/uuid"
)

const defaultPermMode = 0o600

type KeyType int

const (
	KeyPayload KeyType = iota
	KeyId
)

func (k KeyType) String() string {
	switch k {
	case KeyPayload:
		return "payload"
	case KeyId:
		return "id"
	default:
		return "unknown"
	}
}

type KeyDirection int

const (
	KeyIn KeyDirection = iota
	KeyOut
)

const keyPosPrefix = "pos"

func (k KeyDirection) String() string {
	switch k {
	case KeyIn:
		return "in"
	case KeyOut:
		return "out"
	default:
		return "unknown"
	}
}

type Key struct {
	*gofile.BufferedFile
	KeyType      KeyType
	KeyDirection KeyDirection
	Pos          *KeyPosition
}

func MakeKeyName(keyType KeyType, keyDirection KeyDirection) string {
	return keyType.String() + "-" + keyDirection.String()
}

func MakeKeyPath(packId uuid.UUID, keyType KeyType, keyDirection KeyDirection) string {
	return filepath.Join(
		settings.Config.AppDataDirPath,
		settings.Config.KeysStorageDefaultDirName,
		packId.String(),
		MakeKeyName(keyType, keyDirection),
	)
}

func NewKey(packId uuid.UUID, keyType KeyType, keyDirection KeyDirection) (key *Key, err error) {
	var keyPosition *KeyPosition
	path := MakeKeyPath(packId, keyType, keyDirection)
	os.MkdirAll(filepath.Dir(path), defaultPermMode)

	file, err := gofile.NewFile(path, defaultPermMode)
	if err != nil {
		return nil, err
	}

	bf, err := gofile.NewBufferedFile(file, settings.Config.KeysBufferSizeB)
	if err != nil {
		return nil, err
	}

	if keyDirection == KeyOut {
		keyPosition, err = NewKeyPosition(path+keyPosPrefix, defaultPermMode)
		if err != nil {
			file.Close()
			return nil, err
		}
	}

	return &Key{
		BufferedFile: bf,
		KeyType:      keyType,
		KeyDirection: keyDirection,
		Pos:          keyPosition,
	}, nil
}

func (k *Key) ExportShared(dest string) (err error) {
	var file gofile.File
	var direction KeyDirection

	if k.KeyDirection == KeyOut {
		direction = KeyIn
	} else {
		direction = KeyOut
	}

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = gofile.NewFile(filepath.Join(dest, MakeKeyName(k.KeyType, direction)), defaultPermMode)
		},
		func() { err = file.Trunc() },
		func() { err = k.PipeTo(file, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}

func (k *Key) ImportShared(src string) (err error) {
	var file gofile.File

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = gofile.NewFile(filepath.Join(src, MakeKeyName(k.KeyType, k.KeyDirection)), defaultPermMode)
		},
		func() { err = file.PipeTo(k, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}

// Generates new key parts
func (k *Key) GenerateKey(length int64) (err error) {
	if rest := length % k.BufferSize; rest != 0 {
		if key, err := random.GenerateRandomBytes(rest); err != nil {
			return err
		} else if _, err := k.Append(key); err != nil {
			return err
		}
	}

	length = length / k.BufferSize

	for i := int64(0); i < length; i++ {
		if key, err := random.GenerateRandomBytes(k.BufferSize); err != nil {
			return err
		} else if _, err := k.Append(key); err != nil {
			return err
		}
	}

	return nil
}

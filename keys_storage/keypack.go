package keysstorage

import (
	"path/filepath"
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/settings"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type Key struct {
	*KeyBuffer
	prefix            string
	packPrefix        string
	packSharingPrefix string
}

func NewKey(packId uuid.UUID, keyPrefix string, packPrefix string, packSharingPrefix string) (key *Key, err error) {
	kb, err := NewKeyBuffer(keyPath(packId, packPrefix, keyPrefix))
	if err != nil {
		return nil, err
	}

	return &Key{
		KeyBuffer:         kb,
		prefix:            keyPrefix,
		packPrefix:        packPrefix,
		packSharingPrefix: packSharingPrefix,
	}, nil
}

func (k *Key) ExportShared(dest string) (err error) {
	var file filestorage.File

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = filestorage.NewFile(filepath.Join(dest, keyFileName(k.packSharingPrefix, k.prefix)), 0o600)
		},
		func() { err = file.Trunc() },
		func() { err = k.PipeTo(file, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}
func (k *Key) ImportShared(src string) (err error) {
	var file filestorage.File

	return utils.UntilErrorPointer(
		&err,
		func() {
			file, err = filestorage.NewFile(filepath.Join(src, keyFileName(k.packPrefix, k.prefix)), 0o600)
		},
		func() { err = file.PipeTo(k, settings.Config.KeysBufferSizeB) },
		func() { file.Close() },
	)
}

const (
	PACK_ID_KEY = iota
	PACK_PAYLOAD_KEY
	PACK_BASE_LEN
)
const (
	PACK_IN = iota * PACK_BASE_LEN
	PACK_OUT
	PACK_IO_LEN = iota
	PACK_LEN    = PACK_IO_LEN * PACK_BASE_LEN
)

var keyPrefixes = [PACK_LEN][2]string{
	PACK_IN + PACK_ID_KEY:       {PACK_PREFIX_IN, PACK_PREFIX_ID_KEY},
	PACK_IN + PACK_PAYLOAD_KEY:  {PACK_PREFIX_IN, PACK_PREFIX_PAYLOAD_KEY},
	PACK_OUT + PACK_ID_KEY:      {PACK_PREFIX_OUT, PACK_PREFIX_ID_KEY},
	PACK_OUT + PACK_PAYLOAD_KEY: {PACK_PREFIX_OUT, PACK_PREFIX_PAYLOAD_KEY},
}

type keymapping [PACK_LEN]*Key

type KeyPack struct {
	PackId uuid.UUID
	Keys   keymapping
}

func newKeyPack(packId uuid.UUID) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{
		PackId: packId,
		Keys:   keymapping{},
	}

	for i := range keyPack.Keys {
		keyPack.Keys[i], err = NewKey(packId, keyPrefixes[i][1], keyPrefixes[i][0], keyPrefixes[(i+PACK_BASE_LEN)%PACK_LEN][0])
		if err != nil {
			return nil, err
		}
	}

	return keyPack, err
}

func getSharedPackId(src string) (packId uuid.UUID, err error) {
	packId = uuid.UUID{}
	var packageFile filestorage.File

	utils.UntilErrorPointer(
		&err,
		func() { packageFile, err = filestorage.NewFile(filepath.Join(src, "_package_"), 0o600) },
		func() { _, err = packageFile.ReadAt(packId[:], 0) },
		func() { packageFile.Close() },
	)

	return packId, err
}

func importSharedKeyPack(packId uuid.UUID, src string) (keyPack *KeyPack, err error) {
	utils.UntilErrorPointer(
		&err,
		func() { keyPack, err = newKeyPack(packId) },
		func() { err = keyPack.ImportShared(src) },
	)

	return keyPack, err
}

func (p *KeyPack) GenerateKey(keySize int64) (err error) {
	for i := range p.Keys {
		err = p.Keys[i].GenerateKey(keySize)
		if err != nil {
			return err
		}
	}
	return
}

func (p *KeyPack) ExportShared(dest string) (err error) {
	var packageFile filestorage.File

	return utils.UntilErrorPointer(
		&err,
		func() { packageFile, err = filestorage.NewFile(filepath.Join(dest, "_package_"), 0o600) },
		func() { err = packageFile.Trunc() },
		func() { _, err = packageFile.Write(p.PackId[:]) },
		func() {
			packageFile.Close()

			for i := range p.Keys {
				err = p.Keys[i].ExportShared(dest)
				if err != nil {
					return
				}
			}
		},
	)
}

func (p *KeyPack) ImportShared(src string) (err error) {
	for i := range p.Keys {
		err = p.Keys[i].ImportShared(src)
		if err != nil {
			return err
		}
	}
	return
}

func (p *KeyPack) Close() {
	for i := range p.Keys {
		safeClose(p.Keys[i])
	}
}

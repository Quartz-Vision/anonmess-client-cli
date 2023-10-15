package keysstorage

import (
	"path/filepath"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/Quartz-Vision/gofile"

	"github.com/google/uuid"
)

// KeyPack contains all the keys needed for a chat - id key, payload key etc. Also their positions
// It can be easily extended, you just need to add new constants before `PACK_BASE_LEN`
// and then add new keys to the `keyPrefixes``

type KeyPack struct {
	PackId     uuid.UUID
	IdIn       *Key
	IdOut      *Key
	PayloadIn  *Key
	PayloadOut *Key
	keys       [4]*Key
}

func newKeyPack(packId uuid.UUID) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{
		PackId: packId,
	}

	if utils.UntilErrorPointer(
		&err,
		func() {
			keyPack.IdIn, err = NewKey(packId, KeyId, KeyIn)
			keyPack.keys[0] = keyPack.IdIn
		},
		func() {
			keyPack.IdOut, err = NewKey(packId, KeyId, KeyOut)
			keyPack.keys[1] = keyPack.IdOut
		},
		func() {
			keyPack.PayloadIn, err = NewKey(packId, KeyPayload, KeyIn)
			keyPack.keys[2] = keyPack.PayloadIn
		},
		func() {
			keyPack.PayloadOut, err = NewKey(packId, KeyPayload, KeyOut)
			keyPack.keys[3] = keyPack.PayloadOut
		},
	) != nil {
		keyPack.Close()
		keyPack = nil
	}

	return keyPack, err
}

// Helps to get chatId of a shared pack
func getSharedPackId(src string) (packId uuid.UUID, err error) {
	packId = uuid.UUID{}
	var packageFile gofile.File

	utils.UntilErrorPointer(
		&err,
		func() { packageFile, err = gofile.NewFile(filepath.Join(src, "_package_"), 0o600) },
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

// generates new key part of the same size for all of the keys
func (p *KeyPack) GenerateKey(keySize int64) (err error) {
	for i := range p.keys {
		err = p.keys[i].GenerateKey(keySize)
		if err != nil {
			return err
		}
	}
	return
}

func (p *KeyPack) ExportShared(dest string) (err error) {
	var packageFile gofile.File

	return utils.UntilErrorPointer(
		&err,
		func() { packageFile, err = gofile.NewFile(filepath.Join(dest, "_package_"), 0o600) },
		func() { err = packageFile.Trunc() },
		func() { _, err = packageFile.Write(p.PackId[:]) },
		func() {
			packageFile.Close()

			for i := range p.keys {
				err = p.keys[i].ExportShared(dest)
				if err != nil {
					return
				}
			}
		},
	)
}

func (p *KeyPack) ImportShared(src string) (err error) {
	for i := range p.keys {
		err = p.keys[i].ImportShared(src)
		if err != nil {
			return err
		}
	}
	return
}

func (p *KeyPack) Close() {
	for key := range p.keys {
		safeClose(p.keys[key])
	}
}

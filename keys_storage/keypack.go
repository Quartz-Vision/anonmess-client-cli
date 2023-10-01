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

const (
	PACK_PREFIX_IN          = "in"
	PACK_PREFIX_OUT         = "out"
	PACK_PREFIX_ID_KEY      = "id"
	PACK_PREFIX_POS         = "pos"
	PACK_PREFIX_PAYLOAD_KEY = "data"
)
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
type posmapping [PACK_BASE_LEN]*KeyPosition

type KeyPack struct {
	PackId      uuid.UUID
	Keys        keymapping
	KeyPostions posmapping
}

func newKeyPack(packId uuid.UUID) (keyPack *KeyPack, err error) {
	keyPack = &KeyPack{
		PackId:      packId,
		Keys:        keymapping{},
		KeyPostions: posmapping{},
	}

	for i := range keyPack.Keys {
		keyPack.Keys[i], err = NewKey(packId, keyPrefixes[i][1], keyPrefixes[i][0], keyPrefixes[(i+PACK_BASE_LEN)%PACK_LEN][0])
		if err != nil {
			return nil, err
		}
	}
	for i := range keyPack.KeyPostions {
		keyPack.KeyPostions[i], err = NewKeyPosition(packId, keyPrefixes[i][1], keyPrefixes[i][0])
		if err != nil {
			return nil, err
		}
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
	for i := range p.Keys {
		err = p.Keys[i].GenerateKey(keySize)
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

package keysstorage

import (
	"quartzvision/anonmess-client-cli/filestorage"
	"quartzvision/anonmess-client-cli/utils"

	"github.com/google/uuid"
)

type KeyPosition struct {
	file filestorage.File
}

func NewKeyPosition(packId uuid.UUID, keyPrefix string, packPrefix string) (pos *KeyPosition, err error) {
	var file filestorage.File

	return pos, utils.UntilErrorPointer(
		&err,
		func() {
			file, err = filestorage.NewFile(keyPath(packId, packPrefix, keyPrefix, PACK_PREFIX_POS), 0o600)
		},
		func() { _, err = file.Write(utils.Int64ToBytes(0)) },
		func() {
			pos = &KeyPosition{
				file: file,
			}
		},
	)
}

func (p *KeyPosition) Take(dataSize int64) (pos int64, err error) {
	return pos, p.file.TReadWrite(func(txn filestorage.Editable) (err error) {
		data := make([]byte, utils.INT_MAX_SIZE)
		_, err = txn.ReadAt(data, 0)
		if err != nil {
			return err
		}
		pos, _ = utils.BytesToInt64(data)
		nextPos := pos + dataSize
		_, err = txn.WriteAt(utils.Int64ToBytes(nextPos), 0)
		return err
	})
}

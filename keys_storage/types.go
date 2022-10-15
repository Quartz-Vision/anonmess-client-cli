package keysstorage

import "errors"

type KeyId string

var ErrOutOfBound = errors.New("requested key slice is out of the key's bounds")

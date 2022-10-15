package asymmetric

import "errors"

type AsymmetricKey struct {
	Private []byte
	Public  []byte
}

var ErrEncodeFailed = errors.New("data encoding failed")
var ErrDecodeFailed = errors.New("data decoding failed")

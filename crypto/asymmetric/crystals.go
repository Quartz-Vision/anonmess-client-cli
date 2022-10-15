package asymmetric

import (
	"quartzvision/anonmess-client-cli/crypto/random"

	kyber "github.com/kudelskisecurity/crystals-go/crystals-kyber"
)

func GenerateKeys() (key AsymmetricKey, err error) {
	seed, err := random.GenerateRandomBytes(64)
	if err != nil {
		return key, err
	}

	pk, sk := kyber.NewKyber1024().PKEKeyGen(seed)

	return AsymmetricKey{
		Private: sk,
		Public:  pk,
	}, nil
}

func Encode(data []byte, key []byte) (encData []byte, err error) {
	seed, err := random.GenerateRandomBytes(32)
	if err != nil {
		return encData, err
	}

	encData = kyber.NewKyber1024().Encrypt(key, data, seed)
	if encData == nil {
		return encData, ErrEncodeFailed
	}

	return encData, nil
}

func Decode(encData []byte, key []byte) (data []byte, err error) {
	if err != nil {
		return encData, err
	}

	data = kyber.NewKyber1024().Decrypt(key, encData)
	if data == nil {
		return data, ErrDecodeFailed
	}

	return data, nil
}

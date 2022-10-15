package random

import "crypto/rand"


func GenerateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, ErrGeneratingFailed
	}

	return b, nil
}

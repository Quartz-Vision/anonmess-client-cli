package main_test

import (
	"fmt"
	"quartzvision/anonmess-client-cli/app"
	"quartzvision/anonmess-client-cli/crypto/random"
	quartzSymmetric "quartzvision/anonmess-client-cli/crypto/symmetric"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"testing"
)

func BenchmarkSymmetricEncDec(b *testing.B) {
	key, err := random.GenerateRandomBytes(1073741824)
	if err != nil {
		return
	}

	data, err := random.GenerateRandomBytes(1073741824)
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := quartzSymmetric.Encode(data, key)
		if err != nil {
			return
		}

		err = quartzSymmetric.Decode(data, key)
		if err != nil {
			return
		}
	}
}

func BenchmarkReadKey(b *testing.B) {
	if err := app.Init(); err != nil {
		fmt.Println(err.Error())
		return
	}

	buf, err := keysstorage.NewKeyBuffer("test")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := int64(0); i < 200; i++ {
			_, err := buf.GetKeySlice((i*10)<<10, 1024)
			if err != nil {
				fmt.Println("FUCK NO KEY")
				return
			}
		}
	}
}

package main_test

import (
	"fmt"
	"quartzvision/anonmess-client-cli/app"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
	"testing"

	"github.com/Quartz-Vision/gocrypt/random"
	"github.com/Quartz-Vision/gocrypt/symmetric"
)

func BenchmarkSymmetricEncDec(b *testing.B) {
	key, err := random.GenerateRandomBytes(1073741824) // 1G
	if err != nil {
		return
	}

	data, err := random.GenerateRandomBytes(1073741824) // 1G
	if err != nil {
		return
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := symmetric.Encode(data, key)
		if err != nil {
			return
		}

		err = symmetric.Decode(data, key)
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
			k := make([]byte, (i*10)<<10)
			_, err := buf.ReadAt(k, 1024)
			if err != nil {
				fmt.Println("NO KEY")
				return
			}
		}
	}
}

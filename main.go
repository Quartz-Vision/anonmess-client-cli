package main

import (
	"fmt"
	"quartzvision/anonmess-client-cli/app"
	"quartzvision/anonmess-client-cli/crypto/symmetric"
	keysstorage "quartzvision/anonmess-client-cli/keys_storage"
)

func main() {
	if err := app.Init(); err != nil {
		fmt.Printf("Program initialization failed: %s", err.Error())
		return
	}

	data := "test"
	data_bytes := []byte(data)

	// // encData, err := quartzAsymmetric.Encode([]byte(data), keys.Public)
	// // if err != nil {
	// // 	fmt.Println("FUCK NO ENC DATA")
	// // 	return
	// // }

	// // fmt.Printf("\nENC DATA: \nlength = %v (it was %v)\n", len(encData), len(data))

	// // decData, err := quartzAsymmetric.Decode(encData, keys.Private)
	// // if err != nil {
	// // 	fmt.Println("FUCK NO DEC DATA")
	// // 	return
	// // }

	// // fmt.Printf("\n\n%v\n\n", string(decData))

	// key, err := random.GenerateRandomBytes(1024<<10 + 1024)
	// if err != nil {
	// 	fmt.Println("FUCK NO KEYS")
	// 	return
	// }

	/////////////////////////////////
	keysstorage.ManageKey("test-short")
	keysstorage.ManageKey("test")

	// key, err := keysstorage.Keys["test"].GetKeySlice(0, 3)
	// fmt.Printf("KEY %v %v\n", key, err)

	// // buf.AppendKey(key)

	idBytes := []byte(data_bytes)
	key, _ := keysstorage.Keys["test"].GetKeySlice(0, int64(len(idBytes)))

	// fmt.Printf("%v\n", key)

	symmetric.Encode(idBytes, key)

	id, ok := keysstorage.DecodeKeyID(1, idBytes)
	fmt.Printf("%v %v\n", id, ok)
	/////////

	// buf.GetKeySlice(0, 10)
	// buf.GetKeySlice(1024<<10, 5)
	// buf.GetKeySlice(0, 2048)

	// fmt.Printf("LEN %v\n", buf.KeyLength)

	// _, err = buf.GetKeySlice((1024<<10)+2048, 2048)
	// fmt.Println(err.Error())

	// buf.AppendKey([]byte{2, 2, 8})
	/////////////////////////////

	////////////////////////////////
	// s := storage.RawFsStorage{
	// 	FilePath: storage.DataPath("key-test"),
	// }
	// s.Open()
	// l, _ := s.Len()
	// fmt.Printf("LEN %v\n", l)

	// c := make([]byte, 3)
	// s.ReadChunk(c, 0)

	// // s.Append([]byte{1, 3, 3, 7})

	// fmt.Printf("CNT %v\n", c)
	////////////////////////////////

	// key, err := buf.GetKeySlice(0, int64(len(data_bytes)))
	// if err != nil {
	// 	fmt.Println("FUCK NO KEY")
	// 	return
	// }

	// // // data := "test kek lol gg some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash test kek lol gg some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash test kek lol gg some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash test kek lol gg some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash some trash"
	// // data := "abc"

	// err = quartzSymmetric.Encode(data_bytes, key)
	// if err != nil {
	// 	fmt.Println("FUCK NO ENC DATA")
	// 	return
	// }

	// // fmt.Printf("\nENC DATA: \nlength = %v (it was %v)\n", len(data_bytes), len(data))

	// err = quartzSymmetric.Decode(data_bytes, key)
	// if err != nil {
	// 	fmt.Println("FUCK NO DEC DATA")
	// 	return
	// }

	// fmt.Printf("\n%v\n\n", string(data_bytes))

	// fmt.Println("OK")
}

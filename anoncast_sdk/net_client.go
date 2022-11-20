package anoncastsdk

import (
	"fmt"
	"io"
	"net"
	"quartzvision/anonmess-client-cli/settings"
	"time"
)

func Init() (err error) {
	initEvents()

	return nil
}

func Start() (err error) {
	conn, err := net.Dial("tcp", settings.Config.ServerAddr)

	if err != nil {
		return err
	}

	defer conn.Close()

	sizeRawBuf := make([]byte, INT64_SIZE)

	go (func() {
		for {
			for EventsToSend.IsEmpty() {
				time.Sleep(time.Millisecond)
			}

			a := 0
			start := time.Now()
			for val, ok := EventsToSend.Pop(); ok; val, ok = EventsToSend.Pop() {
				a++
				pack := Package{
					Event: val.(*Event),
				}

				buf, _ := pack.MarshalBinary()
				conn.Write(buf)
			}
			end := time.Now()

			fmt.Printf("\nT/send: %v (%s), c: %v\n", (end.Sub(start)).Milliseconds(), "ms", a)
		}
	})()

	a := 0
	for {
		if _, err := io.ReadFull(conn, sizeRawBuf); err != nil {
			return err
		}

		packageSize, _ := BytesToInt64(sizeRawBuf)
		packageBuf := make([]byte, packageSize+int64(len(sizeRawBuf)))

		copy(packageBuf, sizeRawBuf)

		if _, err := io.ReadFull(conn, packageBuf[len(sizeRawBuf):]); err != nil {
			return err
		}

		pack := Package{}
		pack.UnmarshalBinary(packageBuf)

		// EmitEvent(pack.Event)
		a++
		if a >= 999999 {
			EmitEvent(pack.Event)
		}
	}
}

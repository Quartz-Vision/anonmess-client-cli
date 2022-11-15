package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

func Init() (err error) {
	conn, err := net.Dial("tcp", "0.0.0.0:8081")

	if err != nil {
		return err
	}

	defer conn.Close()

	str := "keke"
	buf := make([]byte, len(str)+4)

	copy(buf[4:], str)

	binary.LittleEndian.PutUint32(buf[:4], uint32(len(str)))

	conn.Write(buf)

	for {
		rbuf := make([]byte, 4)

		io.ReadFull(conn, rbuf)

		fmt.Println(string(rbuf))
	}
}

package netclient

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"quartzvision/anonmess-client-cli/settings"
)

func Start() (err error) {
	conn, err := net.Dial("tcp", settings.Config.ServerAddr)

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

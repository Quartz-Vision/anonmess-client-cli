package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Init() (err error) {
	// conn, err := net.Dial("tcp", "0.0.0.0:8081")

	// if err != nil {
	// 	return err
	// }

	// defer conn.Close()

	// str := "keke"
	// buf := make([]byte, len(str)+4)

	// copy(buf[4:], str)

	// binary.LittleEndian.PutUint32(buf[:4], uint32(len(str)))

	// conn.Write(buf)

	// for {
	// 	rbuf := make([]byte, 4)

	// 	io.ReadFull(conn, rbuf)

	// 	fmt.Println(string(rbuf))
	// }
	scanner := bufio.NewScanner(os.Stdin)

	input := ""
	args := []string{}

root:
	for {
		fmt.Print("> ")
		scanner.Scan()
		input = scanner.Text()

		if len(input) > 1 && input[0] == '/' {
			args = strings.Fields(input)

			switch args[0] {
			case "/chat":
				if len(args) < 2 {
					fmt.Println("Creating chat...")
				} else {
					fmt.Printf("Connecting to the chat with id %s\n", args[1])
				}
			case "/exit":
				break root
			}
		}
	}

	return nil
}

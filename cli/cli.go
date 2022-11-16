package cli

import (
	"bufio"
	"fmt"
	"os"
	clientsdk "quartzvision/anonmess-client-cli/client_sdk"

	"github.com/google/shlex"
)

func chat(args ...string) bool {
	const help = (`
/chat list - prints list of chats
/chat create - creates a chat, returns its ID
/chat connect <id> - add a new chat to the cache, adopts its keys	
`)

	if len(args) == 0 {
		fmt.Print(help)
		return false
	}

	switch args[0] {
	case "create":
		chatId, err := clientsdk.CreateChat()
		if err != nil {
			fmt.Printf("Error creating a new chat: %s\n", err)
		} else {
			fmt.Printf("The new chat's id is: %s\n", chatId.String())
		}
	}

	return false
}

func exit(args ...string) bool {
	return true
}

func Init() (err error) {
	var commands = map[string]func(...string) bool{
		"chat": chat,
		"exit": exit,
	}

	scanner := bufio.NewScanner(os.Stdin)
	input := ""

	for {
		fmt.Print("> ")
		scanner.Scan()
		input = scanner.Text()

		if len(input) > 1 && input[0] == '/' && input[1] != '/' {
			args, err := shlex.Split(input[1:])
			if err != nil {
				fmt.Printf("Command parsing failed: %s", err.Error())
				continue
			}

			if fn, ok := commands[args[0]]; !ok {
				fmt.Printf("Command not found: %s", args[0])
			} else if fn(args[1:]...) {
				break
			}
		}
	}

	return nil
}

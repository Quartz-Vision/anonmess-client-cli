package cli

import (
	"bufio"
	"fmt"
	"os"
	anoncastsdk "quartzvision/anonmess-client-cli/anoncast_sdk"
	clientsdk "quartzvision/anonmess-client-cli/client_sdk"
	"quartzvision/anonmess-client-cli/settings"
	"time"

	"github.com/google/shlex"
	"github.com/google/uuid"
)

var currentChat *clientsdk.Chat = nil

func chat(args ...string) bool {
	const help = (`
/chat list - prints list of chats
/chat create <name> - creates a chat, returns its ID
/chat connect <id> <name> - add a new chat to the cache and name it
/chat switch <id> - enter the chat by id
`)

	if len(args) == 0 {
		fmt.Print(help)
		return false
	}

	switch args[0] {
	case "create":
		chat, err := clientsdk.CreateChat(args[1])
		if err != nil {
			fmt.Printf("Error creating a new chat: %s\n", err)
		} else {
			fmt.Printf("The new chat's id is: %s\n", chat.Id.String())
		}
	case "list":
		fmt.Println("Your chats:")
		for id := range clientsdk.Chats {
			fmt.Printf(" - %s   [%s]\n", clientsdk.Chats[id].Name, id.String())
		}
	case "switch":
		if id, err := uuid.Parse(args[1]); err != nil {
			fmt.Printf("Error parsing uuid: %s\n", err)
		} else {
			if chat, ok := clientsdk.Chats[id]; ok {
				currentChat = chat
			} else {
				currentChat = nil
				fmt.Printf("Chat %s doesn't exist\n", args[1])
			}
		}
	case "test":
		start := time.Now()
		c := 1000000
		for i := 0; i < c; i++ {
			anoncastsdk.SendEvent(&anoncastsdk.Event{
				Type: anoncastsdk.EVENT_MESSAGE,
				Data: &anoncastsdk.Message{
					Text: "test",
				},
			})
		}
		end := time.Now()

		fmt.Printf("\nT/append: %v (%s), c: %v\n", (end.Sub(start)).Milliseconds(), "ms", c)
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

	fmt.Printf("\n%s\nApp data dir: %s\n=========================\n\n", settings.Config.AppName, settings.Config.AppDataDirPath)

	anoncastsdk.EventHandlers[anoncastsdk.EVENT_MESSAGE] = func(e *anoncastsdk.Event) {
		fmt.Printf("\n>>> %v\n", e.Data.(*anoncastsdk.Message).Text)
	}

	go anoncastsdk.Start()

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
		} else if len(input) > 0 {
			anoncastsdk.SendEvent(&anoncastsdk.Event{
				Type: anoncastsdk.EVENT_MESSAGE,
				Data: &anoncastsdk.Message{
					Text: input,
				},
			})
		}
	}

	return nil
}

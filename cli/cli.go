package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	anoncastsdk "quartzvision/anonmess-client-cli/anoncast_sdk"
	"quartzvision/anonmess-client-cli/clientsdk"
	"quartzvision/anonmess-client-cli/events"
	"quartzvision/anonmess-client-cli/settings"
	"time"

	"github.com/google/shlex"
	"github.com/google/uuid"
)

var currentChat *clientsdk.Chat = nil

func chat(client *clientsdk.Client, args ...string) bool {
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
		chat, err := client.CreateChat(args[1])
		if err != nil {
			fmt.Printf("Error creating a new chat: %s\n", err)
		} else {
			fmt.Printf("The new chats id is: %s\n", chat.Id.String())
		}
	case "list":
		fmt.Println("Your chats:")
		for id := range client.Chats {
			fmt.Printf(" - %s   [%s]\n", client.Chats[id].Name, id.String())
		}
	case "switch":
		if id, err := uuid.Parse(args[1]); err != nil {
			fmt.Printf("Error parsing uuid: %s\n", err)
		} else {
			if chat, ok := client.Chats[id]; ok {
				currentChat = chat
			} else {
				currentChat = nil
				fmt.Printf("Chat %s doesn't exist\n", args[1])
			}
		}
	case "share":
		if id, err := uuid.Parse(args[1]); err != nil {
			fmt.Printf("Error parsing uuid: %s\n", err)
		} else {
			if chat, ok := client.Chats[id]; ok {
				_ = chat.ExportKeysForShare()
			} else {
				fmt.Printf("Chat %s doesn't exist\n", args[1])
			}
		}
	case "connect":
		if chat, err := client.ImportSharedChat(args[1], args[2]); err != nil {
			fmt.Printf("Error importing chat: %v\n", err.Error())
		} else {
			fmt.Printf("The new chats id is: %s\n", chat.Id.String())
		}
	}

	return false
}

func exit(c *clientsdk.Client, args ...string) bool {
	return true
}

func Init() (err error) {
	client := clientsdk.New()

	var commands = map[string]func(*clientsdk.Client, ...string) bool{
		"chat": chat,
		"exit": exit,
	}

	scanner := bufio.NewScanner(os.Stdin)
	input := ""

	fmt.Printf("\n%s\nApp data dir: %s\n=========================\n\n", settings.Config.AppName, settings.Config.AppDataDirPath)

	client.ListenClient(clientsdk.EVENT_CHAT_MESSAGE, client.WrapMessageHandler(func(msg *clientsdk.ChatMessage) {
		fmt.Printf("\n[%s] >>> %v\n> ", msg.Chat.Name, msg.Text)
	}))
	client.ListenClient(clientsdk.EVENT_ERROR, client.WrapErrorHandler(func(err *anoncastsdk.ClientErrorMessage) {
		fmt.Printf("\n{|ERROR|} >>> %v, %v\n> ", err.Details, err.OriginalError)
		if err.Code == anoncastsdk.ERROR_FATAL {
			fmt.Println("Trying to connect in 5s")
			time.Sleep(5 * time.Second)
			go client.StartConnection()
		}
	}))
	client.ListenClient(clientsdk.EVENT_CONNECTED, func(e *events.Event) {
		fmt.Printf("\n{|SERVER CONNECTED|} \n> ")
	})

	go client.StartConnection()

	client.UpdateChatsList()

	proceed := true

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		client.Close()
		os.Exit(0)
	}()

	for proceed {
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
			} else if fn(client, args[1:]...) {
				break
			}
		} else if len(input) > 0 && currentChat != nil {
			currentChat.SendMessage(input)
		}
	}

	client.Close()

	return nil
}

package main

import (
	"fmt"
	"quartzvision/anonmess-client-cli/app"
)

func main() {
	if err := app.Init(); err != nil {
		fmt.Printf("Program initialization failed: %s", err.Error())
		return
	}
}

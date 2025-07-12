package main

import (
	"fmt"
	"gator/internal/config"
	"os"
)

func main() {
	configObj, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	currentState := state{&configObj}

	cmds := commands{
		function: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)

	if len(os.Args) < 2 {
		fmt.Println("Error: No commands provided")
		os.Exit(1)
	}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	if err := cmds.run(&currentState, cmd); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	} else {
		fmt.Println("Command executed successfully")
	}
}

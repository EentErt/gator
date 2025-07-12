package main

import (
	"fmt"
	"gator/internal/config"
)

func main() {
	config, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	if err := config.SetUser("nate"); err != nil {
		fmt.Println("Error setting user:", err)
	} else {
		fmt.Println("Current user set to:", config.CurrentUserName)
	}

	config, err = config.Read()
	if err != nil {
		fmt.Println("Error reading config after setting user:", err)
	} else {
		fmt.Println("Current user after update:", config.CurrentUserName)
	}
}

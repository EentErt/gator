package main

import (
	"fmt"
	"gator/internal/config"
)

func main() {
	configObj, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	if err := configObj.SetUser("nate"); err != nil {
		fmt.Println("Error setting user:", err)
	} else {
		fmt.Println("Current user set to:", configObj.CurrentUserName)
	}

	configObj, err = config.Read()
	if err != nil {
		fmt.Println("Error reading config after setting user:", err)
	} else {
		fmt.Println("Current user after update:", configObj.CurrentUserName)
	}
	fmt.Println("Database URL:", configObj.DbURL)
}

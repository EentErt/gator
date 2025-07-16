package main

import (
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"

	_ "github.com/lib/pq"
)

var State = state{}

func main() {
	// Read the config file and translate it into a Config struct
	configObj, err := config.Read()
	if err != nil {
		fmt.Println("Error reading config:", err)
	}

	// Get the database URL
	db, err := sql.Open("postgres", configObj.DbURL)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
	}

	dbQueries := database.New(db)

	// Set the current state
	State = state{
		db:  dbQueries,
		cfg: &configObj,
	}

	cmds := commands{
		function: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	if len(os.Args) < 2 {
		fmt.Println("Error: No commands provided")
		os.Exit(1)
	}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	if err := cmds.run(&State, cmd); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	} else {
		fmt.Println("Command executed successfully")
	}
}

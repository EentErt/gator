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

	// register the commands
	cmds.register("login", handlerLogin)                             // log in as existing user
	cmds.register("register", handlerRegister)                       // register a new user
	cmds.register("reset", handlerReset)                             // reset the database
	cmds.register("users", handlerUsers)                             // get a list of users
	cmds.register("agg", handlerAgg)                                 // scrape feeds at an interval
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))     // add a feed
	cmds.register("feeds", handlerFeeds)                             // get a list of feeds
	cmds.register("follow", middlewareLoggedIn(handlerFollow))       // follow a feed
	cmds.register("following", middlewareLoggedIn(handlerFollowing)) // get a list of followed feeds
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))   // unfollow a feed
	cmds.register("browse", middlewareLoggedIn(handlerBrowse))       // browse a number of posts

	// If length of args is less than 2, then no command was given
	if len(os.Args) < 2 {
		fmt.Println("Error: No commands provided")
		os.Exit(1)
	}

	cmd := command{name: os.Args[1], args: os.Args[2:]}

	// run the given command
	if err := cmds.run(&State, cmd); err != nil {
		fmt.Println("Error executing command:", err)
		os.Exit(1)
	} else {
		fmt.Println("Command executed successfully")
	}
}

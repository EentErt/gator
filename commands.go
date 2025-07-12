package main

import (
	"fmt"
	"gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmd map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("Login command requires a username argument")
	}

	username := cmd.args[0]
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Println("Username set to:", username)
	return nil
}

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
	function map[string]func(*state, command) error
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

func (c *commands) run(s *state, cmd command) error {
	if f, ok := c.function[cmd.name]; ok {
		if err := f(s, cmd); err != nil {
			return fmt.Errorf("Error executing command:", cmd.name)
		}
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.function[name] = f
}

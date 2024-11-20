package main

import _ "github.com/lib/pq"
import (
	"fmt"
	"github.com/interyx/internal/config"
	"os"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	names map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.names[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.names[cmd.name]
	if !ok {
		return fmt.Errorf("Command %s not found", cmd.name)
	}
	return handler(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Login requires at least one argument\n")
	}
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("An error occurred: %v\n", err)
	}
	fmt.Printf("%s was logged in successfully!\n", cmd.args[0])
	return nil
}

func main() {
	fmt.Println("Welcome to THE GATOR ZONE! CHOMP CHOMP CHOMP")
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
		fmt.Println("Exiting...")
		os.Exit(1)
	}
	thisState := state{
		cfg: &cfg,
	}
	myCommands := commands{}
	myCommands.names = make(map[string]func(*state, command) error, 5)
	myCommands.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Not enough arguments provided")
		os.Exit(1)
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}
	err = myCommands.run(&thisState, cmd)
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
		os.Exit(1)
	}
}

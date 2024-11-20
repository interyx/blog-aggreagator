package main

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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Login requires at least one argument")
	}
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("An error occurred: %v", err)
	}
	fmt.Printf("%s was logged in successfully!", cmd.args[0])
	return nil
}

func main() {
	fmt.Println("Welcome to THE GATOR ZONE! CHOMP CHOMP CHOMP")
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
		fmt.Println("Exiting...")
		os.Exit(0)
	}
	fmt.Println("Setting user to INTERYX...")
	err = cfg.SetUser("interyx")
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
	}
	fmt.Println("INTERYX set successfully!")
	cfg, err = config.Read()
	if err != nil {
		fmt.Printf("An error has occurred %v\n", err)
	}
	fmt.Println("Printing config...")
	fmt.Println(cfg)
}

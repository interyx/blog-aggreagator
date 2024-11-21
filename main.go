package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/interyx/internal/config"
	"github.com/interyx/internal/database"
	_ "github.com/lib/pq"
	"os"
	"time"
)

type state struct {
	db  *database.Queries
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
	ctx := context.Background()
	if len(cmd.args) == 0 {
		return fmt.Errorf("Login requires at least one argument\n")
	}
	username := cmd.args[0]
	user, err := s.db.GetUser(ctx, username)
	if err != nil {
		return fmt.Errorf("No user with that name found")
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("An error occurred: %v\n", err)
	}
	fmt.Printf("%s was logged in successfully!\n", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("Registration requires a name argument.")
	}
	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	ctx := context.Background()
	newUser, err := s.db.CreateUser(ctx, params)
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(newUser.Name)
	if err != nil {
		return err
	}
	fmt.Printf("User %s was created!\nParams %v\n", newUser.Name, params)
	return nil
}

func handlerReset(s *state, cmd command) error {
  ctx := context.Background()
  err := s.db.DeleteAllUsers(ctx)
  if err != nil {
    return err
  }
  fmt.Println("All users have been deleted")
  return nil;
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
		fmt.Println("Exiting...")
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Welcome to THE GATOR ZONE! CHOMP CHOMP CHOMP")
	cfg, err := config.Read()
	handleError(err)
	db, err := sql.Open("postgres", cfg.Db_url)
	handleError(err)
	thisState := state{
		cfg: &cfg,
		db:  database.New(db),
	}
	myCommands := commands{}
	myCommands.names = make(map[string]func(*state, command) error, 5)
	myCommands.register("login", handlerLogin)
	myCommands.register("register", handlerRegister)
  myCommands.register("reset", handlerReset)
	args := os.Args
	if len(args) < 2 {
		handleError(fmt.Errorf("Not enough arguments provided"))
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}
	err = myCommands.run(&thisState, cmd)
	handleError(err)
}

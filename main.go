package main

import (
	"database/sql"
	"fmt"
	"github.com/interyx/internal/config"
	"github.com/interyx/internal/database"
	_ "github.com/lib/pq"
	"os"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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

func main() {
	cfg, err := config.Read()
	handleError(err)
	db, err := sql.Open("postgres", cfg.Db_url)
	handleError(err)
	thisState := state{
		cfg: &cfg,
		db:  database.New(db),
	}
	cmds := commands{}
	cmds.names = make(map[string]func(*state, command) error, 5)
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	args := os.Args
	if len(args) < 2 {
		handleError(fmt.Errorf("Not enough arguments provided"))
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}
	err = cmds.run(&thisState, cmd)
	handleError(err)
}

package main

import (
	"context"
	// "encoding/xml"
	// "fmt"
	// "github.com/google/uuid"
	"github.com/interyx/gator/internal/database"
	_ "github.com/lib/pq"
	// "html"
	// "io"
	// "net/http"
	// "net/url"
	// "os"
	// "time"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.User)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

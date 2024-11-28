package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/interyx/internal/database"
	"github.com/lib/pq"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, err
	}
	req.Header.Add("User-Agent", "gator")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return &RSSFeed{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &RSSFeed{}, err
	}
	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		return &RSSFeed{}, err
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil
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

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Wrong number of arguments.\nUSAGE: addfeed \"<feed name>\" <url>")
	}

	_, err := url.ParseRequestURI(cmd.args[1])
	if err != nil {
		return fmt.Errorf("Incorrectly formed URL\nUSAGE: addfeed \"<feed name>\" <url>")
	}
	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		UserID:    user.ID,
		Url:       cmd.args[1],
	}
	res, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return fmt.Errorf("An error occurred while creating the feed.  Please try again.")
	}
	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    res.ID,
	}
	_, err = s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", res)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.DeleteAllUsers(ctx)
	if err != nil {
		return err
	}
	fmt.Println("All users have been deleted")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetAllUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		fmt.Printf("* %s", user.Name)
		if user.Name == s.cfg.User {
			fmt.Printf(" (current)")
		}
		fmt.Printf("\n")
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetAllFeeds(context.Background())
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", feeds)
	return nil
}

func handlerAgg(s *state, cmd command) error {
	var time_between_reqs time.Duration
	var err error
	if len(cmd.args) > 1 {
		return fmt.Errorf("Too many arguments.  This function takes a single argument with a duration like \"1m or \"1h\nUsage agg <duration>")
	}
	if len(cmd.args) == 0 {
		time_between_reqs, err = time.ParseDuration("1m")
	} else {
		time_between_reqs, err = time.ParseDuration(cmd.args[0])
	}
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v\n", time_between_reqs)
	ticker := time.NewTicker(time_between_reqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func interpretTime(s string) sql.NullTime {
	timeFormats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC3339,
		time.RFC3339Nano,
	}
	var parsedTime sql.NullTime
	for _, format := range timeFormats {
		if parseAttempt, err := time.Parse(format, s); err == nil {
			parsedTime = sql.NullTime{
				Time:  parseAttempt,
				Valid: true,
			}
			return parsedTime
		}
	}
	return parsedTime
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	handleError(err)
	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	handleError(err)
	err = s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	handleError(err)
	fmt.Printf("Scanning %s...\n", feed.Channel.Title)
	for _, item := range feed.Channel.Item {
		descPub := interpretTime(item.PubDate)
		if !descPub.Valid {
			fmt.Printf("Publish time %v could not be read\n", item.PubDate)
		}
		descString := sql.NullString{
			String: item.Description,
			Valid:  true,
		}
		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: descString,
			PublishedAt: descPub,
			FeedID:      nextFeed.ID,
		}
		_, err := s.db.CreatePost(context.Background(), params)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code.Name() == "unique_violation" {
					continue
				}
			} else {
				fmt.Printf("An error has occurred: %v\n", err)
			}
		}
	}
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("This function requires a feed URL.\nUsage: follow <url>")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	res, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", res)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	follows, err := s.db.GetFeedFollowsForUser(ctx, user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%s's Feeds:\n", user.Name)
	for _, feed := range follows {
		fmt.Printf("* %s\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Wrong number of arguments.\nUsage: unfollow <url>")
	}
	params := database.DeleteFeedFollowParams{
		Name: user.Name,
		Url:  cmd.args[0],
	}
	err := s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	return nil
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("An error has occurred: %v\n", err)
		fmt.Println("Exiting...")
		os.Exit(1)
	}
}

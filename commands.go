package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"time"

	"github.com/google/uuid"
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
	function map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("login command requires a username argument")
	}

	//check if user exists in database
	if _, err := s.db.GetUserByName(context.Background(), cmd.args[0]); err != nil {
		return err
	}

	username := cmd.args[0]
	if err := s.cfg.SetUser(username); err != nil {
		return err
	}

	fmt.Println("Username set to:", username)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	if f, ok := c.function[cmd.name]; !ok {
		return fmt.Errorf("function %s does not exist", cmd.name)
	} else {
		if err := f(s, cmd); err != nil {
			return err
		}
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.function[name] = f
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("register command requires a name")
	}

	userId := uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	timeNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	userParams := database.CreateUserParams{
		ID:        userId.UUID,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      cmd.args[0],
	}

	s.db.CreateUser(context.Background(), userParams)

	_, err := s.db.GetUser(context.Background(), userId.UUID)
	if err != nil {
		fmt.Println("User already exists:", err)
		os.Exit(1)
	}

	s.cfg.SetUser(cmd.args[0])
	fmt.Println("User registered successfully:")
	fmt.Println("User ID:", uuid.UUID.String(userId.UUID))
	fmt.Println("Created at:", userParams.CreatedAt.Time.String())
	fmt.Println("Updated at:", userParams.UpdatedAt.Time.String())
	fmt.Println("User Name:", userParams.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.Generate(context.Background()); err != nil {
		return err
	}

	if err := s.db.ResetFeed(context.Background()); err != nil {
		return err
	}

	fmt.Println("Users table reset")
	os.Exit(0)
	return nil
}

func handlerUsers(s *state, cmd command) error {
	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range userList {
		if user == s.cfg.CurrentUserName {
			fmt.Println("*", user, "(current)")
		} else {
			fmt.Println("*", user)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	rss, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(rss.Channel.Title)
	fmt.Println(rss.Channel.Description)
	fmt.Println(rss.Channel.Link)
	for _, item := range rss.Channel.Item {
		fmt.Println(item.Title)
		fmt.Println(item.PubDate)
		fmt.Println(item.Description)
		fmt.Println(item.Link)
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed requires a name and url")
	}

	timeNow := getNullTimeNow()

	feedName := sql.NullString{
		String: cmd.args[0],
		Valid:  true,
	}

	feedUrl := sql.NullString{
		String: cmd.args[1],
		Valid:  true,
	}

	userID := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	if err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      feedName,
		Url:       feedUrl,
		UserID:    userID,
	}); err != nil {
		return err
	}

	// add the feed_follow entry for the
	if err := handlerFollow(s, command{name: "follow", args: []string{cmd.args[1]}}, user); err != nil {
		return err
	}

	if err := printFeed(s, feedName.String); err != nil {
		return err
	}
	return nil
}

func printFeed(s *state, nameString string) error {
	feed, err := s.db.GetFeed(context.Background(), sql.NullString{String: nameString, Valid: true})
	if err != nil {
		return err
	}

	fmt.Println("ID:", feed.ID)
	fmt.Println("Created At:", feed.CreatedAt.Time.String())
	fmt.Println("Updated At:", feed.UpdatedAt.Time.String())
	fmt.Println("Name:", feed.Name.String)
	fmt.Println("URL:", feed.Url.String)
	fmt.Println("User ID:", feed.UserID.UUID.String())
	return nil

}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		poster, err := s.db.GetUser(context.Background(), feed.UserID.UUID)
		if err != nil {
			return err
		}
		fmt.Println("Feed:", feed.Name.String)
		fmt.Println("URL:", feed.Url.String)
		fmt.Println("Posted By:", poster.Name)
		fmt.Println("")
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("follow command requires a url")
	}

	// Get the requested feed
	feed, err := s.db.GetFeedByUrl(context.Background(), sql.NullString{String: cmd.args[0], Valid: true})
	if err != nil {
		return err
	}

	// Get the current time
	timeNow := getNullTimeNow()

	// Set up the Parameter struct for the feed_follow creation
	Params := database.CreateFeedFollowParams{
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		UserID:    uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID:    sql.NullInt32{Int32: feed.ID, Valid: true},
	}

	newFeedFollow, err := s.db.CreateFeedFollow(context.Background(), Params)
	if err != nil {
		return err
	}

	fmt.Println("New Follow successful:")
	fmt.Println("Feed:", newFeedFollow.FeedName.String)
	fmt.Println("User:", newFeedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("User %s Following:\n", user.Name)
	for _, follow := range follows {
		fmt.Println(follow.FeedName.String)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	// get feed from url
	feed, err := s.db.GetFeedByUrl(context.Background(), sql.NullString{String: cmd.args[0], Valid: true})
	if err != nil {
		return err
	}

	// setup params for delete feed follow function
	Params := database.DeleteFeedFollowParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		FeedID: sql.NullInt32{Int32: feed.ID, Valid: true},
	}

	if err := s.db.DeleteFeedFollow(context.Background(), Params); err != nil {
		return err
	}
	return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	timeNow := getNullTimeNow()

	params := database.MarkFeedFetchedParams{
		UpdatedAt:     timeNow,
		LastFetchedAt: timeNow,
		ID:            feed.ID,
	}

	if err := s.db.MarkFeedFetched(context.Background(), params); err != nil {
		return err
	}

	RSS, err := fetchFeed(context.Background(), feed.Url.String)
	if err != nil {
		return err
	}

}

func getNullTimeNow() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"os"
	"strconv"
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

// login as the given user name
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

// run the given command
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

// register a command to the program
func (c *commands) register(name string, f func(*state, command) error) {
	c.function[name] = f
}

// register a user
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("register command requires a name")
	}

	// set the user name for the sql insertion
	userId := uuid.NullUUID{
		UUID:  uuid.New(),
		Valid: true,
	}

	// set the current time for sql insertion
	timeNow := sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	// set up the parameter struct for sql insertion
	userParams := database.CreateUserParams{
		ID:        userId.UUID,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      cmd.args[0],
	}

	// perform the sql insertion
	s.db.CreateUser(context.Background(), userParams)

	// get the user to verify addition to table
	_, err := s.db.GetUser(context.Background(), userId.UUID)
	if err != nil {
		fmt.Println("User already exists:", err)
		os.Exit(1)
	}

	// print the user's attributes
	s.cfg.SetUser(cmd.args[0])
	fmt.Println("User registered successfully:")
	fmt.Println("User ID:", uuid.UUID.String(userId.UUID))
	fmt.Println("Created at:", userParams.CreatedAt.Time.String())
	fmt.Println("Updated at:", userParams.UpdatedAt.Time.String())
	fmt.Println("User Name:", userParams.Name)
	return nil
}

// reset the database
func handlerReset(s *state, cmd command) error {
	// reset users table
	if err := s.db.Generate(context.Background()); err != nil {
		return err
	}

	// reset feed table
	if err := s.db.ResetFeed(context.Background()); err != nil {
		return err
	}

	fmt.Println("Users table reset")
	os.Exit(0)
	return nil
}

// get a list of users
func handlerUsers(s *state, cmd command) error {
	// get the list
	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	// print the list, highlight the current user
	for _, user := range userList {
		if user == s.cfg.CurrentUserName {
			fmt.Println("*", user, "(current)")
		} else {
			fmt.Println("*", user)
		}
	}
	return nil
}

// Aggregate posts from feeds
func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("agg command requires a time duration")
	}

	// get the wait time (argument)
	waitTime, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	// create a ticker to wait for the given duration
	ticker := time.NewTicker(waitTime)
	for ; ; <-ticker.C {
		// scrape the feeds
		scrapeFeeds(s)
	}
}

// add a feed to the feed table
func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed requires a name and url")
	}

	timeNow := getNullTimeNow()

	// set name for sql insertion
	feedName := sql.NullString{
		String: cmd.args[0],
		Valid:  true,
	}

	// set url for sql insertion
	feedUrl := sql.NullString{
		String: cmd.args[1],
		Valid:  true,
	}

	// set user_id for sql insertion
	userID := uuid.NullUUID{
		UUID:  user.ID,
		Valid: true,
	}

	// create the feed
	if err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      feedName,
		Url:       feedUrl,
		UserID:    userID,
	}); err != nil {
		return err
	}

	// add the feed_follow entry for the feed and current user
	if err := handlerFollow(s, command{name: "follow", args: []string{cmd.args[1]}}, user); err != nil {
		return err
	}

	if err := printFeed(s, feedName.String); err != nil {
		return err
	}
	return nil
}

// print the values of a feed
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

// print a list of feeds
func handlerFeeds(s *state, cmd command) error {
	// get the list
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	// print the list
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

// follow a given feed
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

	// create the feed_follow entry
	newFeedFollow, err := s.db.CreateFeedFollow(context.Background(), Params)
	if err != nil {
		return err
	}

	fmt.Println("New Follow successful:")
	fmt.Println("Feed:", newFeedFollow.FeedName.String)
	fmt.Println("User:", newFeedFollow.UserName)
	return nil
}

// get a list of the feeds that the user is following
func handlerFollowing(s *state, cmd command, user database.User) error {
	// get the list
	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return err
	}

	// print the list
	fmt.Printf("User %s Following:\n", user.Name)
	for _, follow := range follows {
		fmt.Println(follow.FeedName.String)
	}
	return nil
}

// unfollow a given feed
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

	// delete the feed_follow entry
	if err := s.db.DeleteFeedFollow(context.Background(), Params); err != nil {
		return err
	}
	return nil
}

// check if a user is logged in then run the function with that user as a parameter
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUserByName(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

// scrape the oldest feed
func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	// get current time
	timeNow := getNullTimeNow()

	// set parameters for marking the feed as fetched
	params := database.MarkFeedFetchedParams{
		UpdatedAt:     timeNow,
		LastFetchedAt: timeNow,
		ID:            feed.ID,
	}

	// mark the feed as fetched
	if err := s.db.MarkFeedFetched(context.Background(), params); err != nil {
		return err
	}

	// fetch the feed
	RSS, err := fetchFeed(context.Background(), feed.Url.String)
	if err != nil {
		return err
	}

	// print the feed
	fmt.Println(RSS.Channel.Title)
	for _, item := range RSS.Channel.Item {
		// attempt to parse the publish time of the feed
		publishedAt, err := parseTime(item.PubDate)
		if err != nil {
			// fmt.Println("failed to parse time")
			return err
		}

		// set a description value for sql insertion
		desc := sql.NullString{}
		if item.Description != "" {
			desc = sql.NullString{String: item.Description, Valid: true}
		}

		// set up parameters for creating a post in the posts table
		params := database.CreatePostParams{
			CreatedAt:   timeNow,
			UpdatedAt:   timeNow,
			Title:       item.Title,
			Url:         item.Link,
			Description: desc,
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: true},
			FeedID:      feed.ID,
		}

		// add the post to the posts table
		if err := s.db.CreatePost(context.Background(), params); err != nil {
			return err
		}
		fmt.Println(" -", item.Title)
	}
	return nil
}

// get a nulltime struct for the current time for sql insertion
func getNullTimeNow() sql.NullTime {
	return sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}
}

// parse a time stamp string
func parseTime(timeStamp string) (time.Time, error) {
	// list of layouts to try parsing
	layouts := []string{time.ANSIC, time.UnixDate, time.RubyDate, time.RFC822, time.RFC822Z,
		time.RFC850, time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano}

	// try parsing
	for _, layout := range layouts {
		output, err := time.Parse(layout, timeStamp)
		if err == nil {
			return output, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time stamp")
}

// print a number of posts
func handlerBrowse(s *state, cmd command, user database.User) error {
	// default value to print is 2
	limitParam := int32(2)

	// set the number to print if a number is given
	if len(cmd.args) > 0 {
		limit, err := strconv.ParseInt(cmd.args[0], 10, 32)
		if err != nil {
			return err
		}
		limitParam = int32(limit)
	}

	// set up parameters for sql request
	params := database.GetPostsForUserParams{
		UserID: uuid.NullUUID{UUID: user.ID, Valid: true},
		Limit:  limitParam,
	}

	// get the posts
	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return err
	}

	// print the posts
	printPosts(posts)
	return nil
}

// print a slice of posts
func printPosts(posts []database.GetPostsForUserRow) {
	for i, post := range posts {
		fmt.Println("-- Post", i+1)
		fmt.Println(post.Title)
		fmt.Println(post.PublishedAt.Time)
		fmt.Println(post.Description.String)
	}
}

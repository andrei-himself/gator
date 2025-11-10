package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"os"
	"time"
	"errors"
	"context"
	"database/sql"
	"github.com/andrei-himself/gator/internal/config"
	"github.com/andrei-himself/gator/internal/database"
	"github.com/andrei-himself/gator/internal/rss"
	"github.com/google/uuid"
)

type state struct {
	cfg *config.Config
	db *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	m map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	com, ok := c.m[cmd.name]
	if ok != true {
		return fmt.Errorf(fmt.Sprintf("command '%v' not found", cmd.name))
	}
	return com(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.m[name] = f
} 

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command expects username as an argument")
	}
	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Fprintln(os.Stderr, "username doesn't exist in the database")
		os.Exit(1)
	} else if err != nil {
		return err
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	} 
	fmt.Printf("User '%s' has been set!\n", name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command expects username as an argument")
	}
	name := cmd.args[0]
	data := database.CreateUserParams{
		ID : uuid.New(),
		CreatedAt : time.Now(),
		UpdatedAt : time.Now(),
		Name : name,
	}
	_, err := s.db.CreateUser(context.Background(), data)
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User '%s' has been regstered!\n", name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	} 
	err = s.db.DeleteFeeds(context.Background())
	if err != nil {
		return err
	} 
	err = s.db.DeleteFeedFollows(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Users, feeds, and feed follows deleted successfully!")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	var users []database.User
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	current := s.cfg.CurrentUserName
	for _, v := range users {
		if v.Name == current {
			fmt.Println("*", current, "(current)")
			continue
		}
		fmt.Println("*", v.Name)
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("agg command expects time between requests as an argument")
	}
	
	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %v\n", duration)
	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	} 
	return nil
}

func handlerAddfeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed command expects name and url as arguments")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	feed := database.CreateFeedParams{
		ID : uuid.New(),
		CreatedAt : time.Now(),
		UpdatedAt : time.Now(),
		Name : name,
		Url : url,
		UserID : user.ID,
	}

	createdFeed, err := s.db.CreateFeed(context.Background(), feed)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", createdFeed)

	followCmd := command{
		name : "follow",
		args : []string{
			fmt.Sprintf("%s", createdFeed.Url),
		},
	}
	err = handlerFollow(s, followCmd, user)
	if err != nil {
		return err
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, v := range feeds {
		fmt.Printf("%+v\n", v)
		user, err := s.db.GetUserByID(context.Background(), v.UserID)
		if err != nil {
			return err
		}
		fmt.Println(user.Name)
	}
	
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("follow command expects url as an argument")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	feedfollow := database.CreateFeedFollowParams{
		ID : uuid.New(),
		CreatedAt : time.Now(),
		UpdatedAt : time.Now(),
		UserID : user.ID,
		FeedID : feed.ID,
	}

	_, err = s.db.CreateFeedFollow(context.Background(), feedfollow)
	if err != nil {
		return err
	}
	fmt.Println(user.Name, "now follows feed", feed.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	feedFollowsForUser, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}
	fmt.Printf("%+v", feedFollowsForUser)
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("unfollow command expects feed url as an argument")
	}
	url := cmd.args[0]
	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return err
	}

	params := database.DeleteFeedFollowParams{
		UserID : user.ID,
		FeedID : feed.ID,
	}

	return s.db.DeleteFeedFollow(context.Background(), params)
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	
	return func(s *state,cmd command) error{
		username := s.cfg.CurrentUserName
		user, err := s.db.GetUser(context.Background(), username)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func scrapeFeeds(s *state) error {
	nextFeedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	markParams := database.MarkFeedFetchedParams{
		ID : nextFeedToFetch.ID,
		LastFetchedAt : sql.NullTime{Time: time.Now(), Valid: true},
	}
	err = s.db.MarkFeedFetched(context.Background(), markParams)
	if err != nil {
		return err
	}
	
	feed, err := rss.FetchFeed(context.Background(), nextFeedToFetch.Url)
	if err != nil {
		return err
	}

	for _, v := range feed.Channel.Item {
		fmt.Printf("%+v\n", v.Title)
		
	}
	return nil
}

func main() {
	var s state
	conf, err := config.Read()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	} 
	s.cfg = &conf

	db, err := sql.Open("postgres", s.cfg.DBURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	dbQueries := database.New(db)
	s.db = dbQueries

	var commands commands
	commands.m = map[string]func(*state, command) error{}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", middlewareLoggedIn(handlerAddfeed))
	commands.register("feeds", handlerFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	args := os.Args
	if len(args) < 2 {
		err := fmt.Errorf("Not enough arguments")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	name := args[1]
	cmdArgs := args[2:]
	cmd := command{
		name: name,
		args: cmdArgs,
	}
	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
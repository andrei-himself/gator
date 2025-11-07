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
	fmt.Println("Users and feeds deleted successfully!")
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
	feed, err := rss.FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", *feed)
	return nil
}

func handlerAddfeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed command expects name and url as arguments")
	}
	name := cmd.args[0]
	url := cmd.args[1]

	username := s.cfg.CurrentUserName
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}

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
	fmt.Printf("%+v", createdFeed)
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

func handlerFollow(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("follow command expects url as an argument")
	}
	url := cmd.args[0]

	username := s.cfg.CurrentUserName
	
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
	commands.register("addfeed", handlerAddfeed)
	commands.register("feeds", handlerFeeds)
	commands.register("follow", handlerFollow)
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
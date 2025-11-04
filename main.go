package main

import _ "github.com/lib/pq"

import (
	"fmt"
	"os"
	"database/sql"
	"github.com/andrei-himself/gator/internal/config"
	"github.com/andrei-himself/gator/internal/database"
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
	err := s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	} 
	fmt.Println("User has been set!")
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
package main

import (
	"fmt"
	"github.com/andrei-himself/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Errorf("%v\n", err)
	} 
	err = cfg.SetUser("andrei")
	if err != nil {
		fmt.Errorf("%v\n", err)
	}
	cfg, err = config.Read()
	if err != nil {
		fmt.Errorf("%v\n", err)
	} 
	fmt.Printf("%+v\n", cfg)
}
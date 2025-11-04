package config

import (
	"os"
	"path/filepath"
	"io"
	"encoding/json"
)

type Config struct {
	DBURL string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	file, err := os.Open(filepath.Join(home, ".gatorconfig.json"))
	if err != nil {
		return Config{}, err
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	var config Config
	if err = json.Unmarshal(bytes, &config); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUserName = username
	jsonData, err := json.Marshal(*c)
	if err != nil {
		return err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(home, ".gatorconfig.json"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}
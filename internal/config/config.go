package config

import (
	"encoding/json"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}

	// Read the config file from the home directory
	var config Config
	configText, err := os.ReadFile(homeDir + configFileName)
	if err != nil {
		return Config{}, err
	}

	// put the config into a new Config struct
	if err := json.Unmarshal(configText, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(user string) error {
	c.CurrentUserName = user
	if err := write(c); err != nil {
		return err
	}
	return nil
}

func write(c *Config) error {
	// Get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// convert the config to JSON
	configText, err := json.Marshal(c)
	if err != nil {
		return err
	}

	// write the config to file
	if err := os.WriteFile(homeDir+configFileName, configText, 0644); err != nil {
		return err
	}
	return nil
}

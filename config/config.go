package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type Config struct {
	Token  string `json:"token"`
	UserID string `json:"user_id"`
	Host   string `json:"host"`
}

const (
	FieldToken  string = "Token"
	FieldUserID string = "UserID"
	FieldHost   string = "Host"
)

// GetOrCreate retrieves a *Config from reading or creating config on filesystem.
func GetOrCreate() *Config {
	var c Config

	// Check if config file exists
	path := getPath()
	file := filepath.Join(path, "config.json")
	_, err := os.Stat(file)
	if err != nil {
		// Missing file -> create folder
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			// Failed to create config folder
			log.Fatalf("Unable to create config folder: %v", err)
		}
		c.save() // create config.json with empty values
		return &c
	}

	// Config file exists -> read and unmarshal
	var data []byte
	data, err = os.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read existing config file: %v", err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		log.Fatalf("Unable to unmarshal existing config file: %v", err)
	}

	return &c
}

func (c *Config) SetField(field string, value interface{}) {
	s := reflect.ValueOf(c).Elem()
	f := s.FieldByName(field)

	if !f.IsValid() {
		log.Fatalf("No such field: %s", field)
	}

	if !f.CanSet() {
		log.Fatalf("Cannot set %s field value", field)
	}

	v := reflect.ValueOf(value)
	fieldType := f.Type()
	valueType := v.Type()

	if fieldType != valueType {
		log.Fatalf("Provided value type did not match obj field type")
	}

	// Set config and save to local file.
	f.Set(v)
	c.save()
}

func (c *Config) save() {
	bytes, err := json.Marshal(c)
	if err != nil {
		log.Fatalf("Unable to marshal config: %v", err)
	}

	file := filepath.Join(getPath(), "config.json")
	err = os.WriteFile(file, bytes, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to save config: %v", err)
	}
}

// getPath returns the absolute path of Sharify's
// config directory based on the user's operating system.
//
// Windows - %AppData%\.sharifydesktop
// Mac - $HOME/Library/Application Support/sharify-desktop
// Linux - $XDG_CONFIG_HOME/sharify-desktop or $HOME/.config/sharify-desktop
func getPath() string {
	parent, err := os.UserConfigDir()
	if parent == "" {
		log.Fatalf("Unable to get config path: %v", err)
	}

	var child string
	switch runtime.GOOS {
	case "windows":
		child = ".sharifydesktop"
	default:
		child = "sharify-desktop"
	}

	return filepath.Join(parent, child)
}

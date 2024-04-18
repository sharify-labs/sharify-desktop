package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

type Config struct {
	Token string `json:"token"`
	Host  string `json:"host"`
}

func NewConfig() *Config {
	var c Config

	path := getPath()
	file := filepath.Join(path, "config.json")
	if _, err := os.Stat(file); err != nil {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			log.Fatalf("Unable to create config folder: %v", err)
		}
		c.save() // create config.json with empty values
		return &c
	}

	// Config file exists -> read and unmarshal
	data, err := os.ReadFile(file)
	if err != nil {
		log.Fatalf("Unable to read existing config file: %v", err)
	}

	if err = json.Unmarshal(data, &c); err != nil {
		log.Fatalf("Unable to unmarshal existing config file: %v", err)
	}
	return &c
}

func (c *Config) SetField(name string, value interface{}) {
	field := reflect.ValueOf(c).Elem().FieldByName(name)

	if !field.IsValid() {
		log.Fatalf("No such field: %s", name)
	}

	if !field.CanSet() {
		log.Fatalf("Cannot set %s field value", name)
	}

	v := reflect.ValueOf(value)
	fieldType := field.Type()
	valueType := v.Type()

	if fieldType != valueType {
		log.Fatalf("Provided value type did not match obj field type")
	}

	// Set config and save to local file.
	field.Set(v)
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
// Linux - $XDG_CONFIG_HOME/sharify-desktop or $HOME/.config/sharify-desktop.
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

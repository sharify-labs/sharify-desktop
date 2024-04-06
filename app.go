package main

import (
	"bytes"
	"fmt"
	"fyne.io/systray/example/icon"
	"github.com/ncruces/zenity"
	"github.com/sharify-labs/sharify-go"
	"golang.design/x/clipboard"
	"io/fs"
)

type App struct {
	api    *sharify.API
	config *Config
}

func NewApp() *App {
	config := NewConfig()
	api := sharify.New(sharify.AuthToken(config.Token), sharify.UserAgent("sharify-desktop/"+Version))
	return &App{
		config: config,
		api:    api,
	}
}

func (a *App) Icon() []byte {
	iconBytes, err := fs.ReadFile(assets, "assets/sharify-desktop-icon.png")
	if err != nil {
		fmt.Printf("unable to load icon: %v", err)
		return icon.Data
	}
	return iconBytes
}

func (_ *App) DisplayNotification(message string) {
	_ = zenity.Notify(
		message,
		zenity.Title("Sharify"),
		zenity.Icon(zenity.InfoIcon),
	)
}

func (a *App) PromptSettings() {
	var field string
	var value string
	var err error
	field, err = zenity.List(
		"Select a setting to change:",
		[]string{
			string(FieldToken),
			string(FieldHost),
		},
		zenity.Title("Settings"),
	)
	if err != nil {
		// Cancelled
		return
	}

	switch ConfigField(field) {
	case FieldHost:
		var availableHosts []string
		availableHosts, err = a.api.ListHosts()
		if err != nil {
			_ = zenity.Error(err.Error(), zenity.Title("Error"), zenity.Icon(zenity.ErrorIcon))
			return
		}
		value, err = zenity.List(
			"Select a host:",
			availableHosts,
			zenity.Title("Hosts"),
		)
		if err != nil {
			// Cancelled
			return
		}
	case FieldToken:
		value, err = zenity.Entry(
			"Enter your API token:",
			zenity.Title("Token"),
		)
		if err != nil {
			// Cancelled
			return
		}
		a.api.SetToken(value)
	}

	a.config.SetField(field, value)

	// Display success message
	_ = zenity.Notify(
		fmt.Sprintf("Successfully updated %s!", field),
		zenity.Title("Success"),
		zenity.Icon(zenity.InfoIcon),
	)
}

func (a *App) ShortenLink() {
	var result *sharify.UploadDetails
	var data []byte
	var err error

	// Attempt to read url from clipboard
	if data = clipboard.Read(clipboard.FmtText); data != nil {
		if a.config.Token != "" {
			result, err = a.api.ShortenLink(string(data), sharify.SetDomain(a.config.Host))
		} else {
			result, err = a.api.ShortenLink(string(data))
		}
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to shorten url: %v", err))
			a.DisplayNotification(fmt.Sprintf("failed to shorten url: %v", err))
			return
		} else {
			clipboard.Write(clipboard.FmtText, []byte(result.URL))
			a.DisplayNotification(MessageUploadSuccess)
			return
		}
	}

	// Clipboard read failed
	a.DisplayNotification(MessageErrReadingClipboard)
}

func (a *App) UploadClipboard() {
	var result *sharify.UploadDetails
	var data []byte
	var err error

	// Attempt to read image
	if data = clipboard.Read(clipboard.FmtImage); data != nil {
		if a.config.Token != "" {
			result, err = a.api.UploadFile(bytes.NewReader(data), sharify.SetDomain(a.config.Host))
		} else {
			result, err = a.api.UploadFile(bytes.NewReader(data))
		}
		if err != nil {
			a.DisplayNotification(fmt.Sprintf("failed to upload image: %v", err))
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(result.URL))
		a.DisplayNotification(MessageUploadSuccess)
		return
	}

	// Not image -> Attempt to read text
	if data = clipboard.Read(clipboard.FmtText); data != nil {
		if a.config.Token != "" {
			result, err = a.api.UploadPaste(string(data), sharify.SetDomain(a.config.Host))
		} else {
			result, err = a.api.UploadPaste(string(data))
		}
		if err != nil {
			a.DisplayNotification(fmt.Sprintf("failed to upload text: %v", err))
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(result.URL))
		a.DisplayNotification(MessageUploadSuccess)
		return
	}

	// Clipboard read failed
	a.DisplayNotification(MessageErrReadingClipboard)
}

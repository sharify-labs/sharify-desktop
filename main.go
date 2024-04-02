package main

import (
	"bytes"
	"embed"
	"fmt"
	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/ncruces/zenity"
	"github.com/sharify-labs/sharify-desktop/config"
	"github.com/sharify-labs/sharify-go"
	"golang.design/x/clipboard"
	"io/fs"
)

//go:embed assets/*
var assets embed.FS
var Version string

const (
	UploadSuccessMessage   string = "URL copied to clipboard."
	ErrReadingMessage      string = "Failed to read clipboard. Please try again."
	QuitTooltip            string = "Quit Sharify"
	UploadClipboardTooltip string = "Upload your clipboard"
	ShortenURLTooltip      string = "Shorten a URL in your clipboard"
	ChangeSettingsTooltip  string = "Change settings"
)

func main() {
	systray.Run(onReady, nil)
}

func exitApp() {
	systray.Quit()
}

func loadIcon() []byte {
	iconBytes, err := fs.ReadFile(assets, "assets/sharify-desktop-icon.png")
	if err != nil {
		fmt.Printf("unable to load icon: %v", err)
		return icon.Data
	}
	return iconBytes
}

var api *sharify.API

func onReady() {
	systray.SetIcon(loadIcon())
	systray.SetTitle("Sharify")
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	mQuit := systray.AddMenuItem("Quit", QuitTooltip)
	mUpload := systray.AddMenuItem("Upload Clipboard", UploadClipboardTooltip)
	mShorten := systray.AddMenuItem("Shorten URL", ShortenURLTooltip)
	mSettings := systray.AddMenuItem("Settings", ChangeSettingsTooltip)
	go func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				exitApp()
			case <-mUpload.ClickedCh:
				uploadClipboard()
			case <-mShorten.ClickedCh:
				shortenURL()
			case <-mSettings.ClickedCh:
				promptSettingsList()
			}
		}
	}()
}

func promptSettingsList() {
	var (
		field string
		value string
		err   error
	)
	field, err = zenity.List(
		"Select a setting to change:",
		[]string{
			config.FieldToken,
			config.FieldHost,
		},
		zenity.Title("Settings"),
	)
	if err != nil {
		// Cancelled
		return
	}

	switch field {
	case config.FieldHost:
		var availableHosts []string
		availableHosts, err = getAvailableHosts()
		if err != nil {
			// Unable to get list of available hosts
			_ = zenity.Error(err.Error(), zenity.Title("Error"), zenity.Icon(zenity.ErrorIcon))
			return
		}
		// Display selection
		value, err = zenity.List(
			"Select a host:",
			availableHosts,
			zenity.Title("Hosts"),
		)
		if err != nil {
			// Cancelled
			return
		}
	default:
		value, err = zenity.Entry(
			"Enter your "+field,
			zenity.Title("Update "+field),
		)
		if err != nil {
			// Cancelled
			return
		}
	}

	c := config.GetOrCreate()
	c.SetField(field, value)

	// Display success message
	_ = zenity.Notify(
		fmt.Sprintf("Successfully updated %s!", field),
		zenity.Title("Success"),
		zenity.Icon(zenity.InfoIcon),
	)
}

func getAvailableHosts() ([]string, error) {
	c := config.GetOrCreate()
	api = sharify.New(sharify.AuthToken(c.Token), sharify.UserAgent("sharify-desktop/"+Version))
	result, err := api.ListHosts()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available hosts: %v", err)
	}
	return result, nil
}

func displayNotification(message string) {
	_ = zenity.Notify(
		message,
		zenity.Title("Sharify"),
		zenity.Icon(zenity.InfoIcon),
	)
}

func uploadClipboard() {
	var data []byte

	// Attempt to read image
	if data = clipboard.Read(clipboard.FmtImage); data != nil {
		api = sharify.New(sharify.AuthToken(config.GetOrCreate().Token), sharify.UserAgent("sharify-desktop/"+Version))
		result, err := api.UploadFile(bytes.NewReader(data), sharify.SetDomain(config.GetOrCreate().Host))
		if err != nil {
			displayNotification(fmt.Sprintf("failed to upload image: %v", err))
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(result.URL))
		displayNotification(UploadSuccessMessage)
		return
	}

	// Not image -> Attempt to read text
	if data = clipboard.Read(clipboard.FmtText); data != nil {
		resultURL, err := uploadText(data)
		if err != nil {
			displayNotification(fmt.Sprintf("failed to upload text: %v", err))
			return
		} else {
			clipboard.Write(clipboard.FmtText, []byte(resultURL))
			displayNotification(UploadSuccessMessage)
			return
		}
	}

	// Clipboard read failed
	displayNotification(ErrReadingMessage)
}

func shortenURL() {
	var data []byte

	// Attempt to read url from clipboard
	if data = clipboard.Read(clipboard.FmtText); data != nil {
		api = sharify.New(sharify.AuthToken(config.GetOrCreate().Token), sharify.UserAgent("sharify-desktop/"+Version))
		result, err := api.ShortenLink(string(data), sharify.SetDomain(config.GetOrCreate().Host))
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to shorten url: %v", err))
			displayNotification(fmt.Sprintf("failed to shorten url: %v", err))
			return
		} else {
			clipboard.Write(clipboard.FmtText, []byte(result.URL))
			displayNotification(UploadSuccessMessage)
			return
		}
	}

	// Clipboard read failed
	displayNotification(ErrReadingMessage)
}

func uploadText(data []byte) (string, error) {
	api = sharify.New(sharify.AuthToken(config.GetOrCreate().Token), sharify.UserAgent("sharify-desktop/"+Version))
	result, err := api.CreatePaste(string(data), sharify.SetDomain(config.GetOrCreate().Host))
	if err != nil {
		return "", fmt.Errorf("failed to upload paste: %v", err)
	}
	return result.URL, nil
}

package main

import (
	"embed"
	"fmt"
	"fyne.io/systray"
	"golang.design/x/clipboard"
)

//go:embed assets/*
var assets embed.FS
var Version string

func main() {
	fmt.Println("Initializing...")
	systray.Run(func() {
		fmt.Println("Checking clipboard compatibility...")
		err := clipboard.Init()
		if err != nil {
			panic(err)
		}
		fmt.Println("Starting App...")
		app := NewApp()

		systray.SetIcon(app.Icon())
		systray.SetTitle("Sharify")

		go func() {
			mQuit := systray.AddMenuItem("Quit", TooltipQuit)
			mUpload := systray.AddMenuItem("Upload Clipboard", TooltipUploadClipboard)
			mShorten := systray.AddMenuItem("Shorten URL", TooltipShortenURL)
			mSettings := systray.AddMenuItem("Settings", TooltipChangeSettings)
			for {
				select {
				case <-mQuit.ClickedCh:
					systray.Quit()
				case <-mUpload.ClickedCh:
					app.UploadClipboard()
				case <-mShorten.ClickedCh:
					app.ShortenLink()
				case <-mSettings.ClickedCh:
					app.PromptSettings()
				}
			}
		}()
		fmt.Println("App is running.")
	}, nil)
}

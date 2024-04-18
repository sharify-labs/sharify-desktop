package main

import (
	"embed"
	"log"

	"fyne.io/systray"
	"golang.design/x/clipboard"
)

//go:embed assets/*
var assets embed.FS
var Version string

func main() {
	log.Println("Initializing...")
	systray.Run(func() {
		log.Println("Checking clipboard compatibility...")
		err := clipboard.Init()
		if err != nil {
			panic(err)
		}
		log.Println("Starting App...")
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
		log.Println("App is running.")
	}, nil)
}

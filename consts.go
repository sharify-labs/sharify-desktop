package main

const (
	MessageUploadSuccess       string = "URL copied to clipboard."
	MessageErrReadingClipboard string = "Failed to read clipboard. Please try again."
	TooltipQuit                string = "Quit Sharify"
	TooltipUploadClipboard     string = "Upload your clipboard"
	TooltipShortenURL          string = "Shorten a URL in your clipboard"
	TooltipChangeSettings      string = "Change settings"
)

type ConfigField string

const (
	FieldToken ConfigField = "Token"
	FieldHost  ConfigField = "Host"
)

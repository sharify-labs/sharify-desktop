package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/ncruces/zenity"
	"github.com/sharify-labs/sharify-desktop/config"
	"golang.design/x/clipboard"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
)

//go:embed assets/*
var assets embed.FS

const (
	UploadSuccessMessage   string = "URL copied to clipboard."
	ErrReadingMessage      string = "Failed to read clipboard. Please try again."
	QuitTooltip            string = "Quit Sharify"
	UploadClipboardTooltip string = "Upload your clipboard"
	ShortenURLTooltip      string = "Shorten a URL in your clipboard"
	ChangeSettingsTooltip  string = "Change settings"
)

const (
	PasteURL  string = "https://paste.crystaldev.co"
	ZephyrURL string = "https://xericl.dev/api"
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
			config.FieldUserID,
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
	// Prepare the GET request
	var requestBody bytes.Buffer

	req, err := http.NewRequest("GET", ZephyrURL+"/hosts", &requestBody)
	if err != nil {
		log.Printf("Failed to create POST request: %v", err)
		return nil, err
	}
	req.Header.Add("X-Upload-Token", c.Token)
	req.Header.Add("X-Upload-User", c.UserID)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send GET request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get available hosts, status code: %d", resp.StatusCode)
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read available hosts response: %v", err)
	}

	var result []string
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal hosts response: %v", err)
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
		resultURL, err := uploadImage(data)
		if err != nil {
			displayNotification(fmt.Sprintf("failed to upload image: %v", err))
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(resultURL))
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
		resultURL, err := createRedirect(data)
		if err != nil {
			fmt.Println(fmt.Sprintf("failed to shorten url: %v", err))
			displayNotification(fmt.Sprintf("failed to shorten url: %v", err))
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

func uploadImage(data []byte) (string, error) {
	// Prepare the POST request
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Create a form file field 'file'
	fileWriter, err := multipartWriter.CreateFormFile("file", "image.png")
	if err != nil {
		log.Printf("Failed to create form file: %v", err)
		return "", err
	}
	// Write the image data to the form file
	if _, err = fileWriter.Write(data); err != nil {
		log.Printf("Failed to write image data to form file: %v", err)
		return "", err
	}
	// Important to close the multipart writer to set the terminating boundary
	if err = multipartWriter.Close(); err != nil {
		log.Printf("Failed to close multipart writer: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", ZephyrURL+"/upload", &requestBody)
	if err != nil {
		log.Printf("Failed to create POST request: %v", err)
		return "", err
	}
	c := config.GetOrCreate()
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Add("X-Upload-Token", c.Token)
	req.Header.Add("X-Upload-User", c.UserID)
	req.Header.Add("X-Upload-Host", c.Host)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send POST request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload image, status code: %d", resp.StatusCode)
	}

	var u []byte
	u, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read upload response body: %v", err)
	}

	return string(u), nil
}

func createRedirect(data []byte) (string, error) {
	// Prepare the POST request
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	// Create a form file field 'file'
	formField, err := multipartWriter.CreateFormField("long_url")
	if err != nil {
		log.Printf("Failed to create form file: %v", err)
		return "", err
	}
	// Write the image data to the form file
	if _, err = formField.Write(data); err != nil {
		log.Printf("Failed to write url data to form file: %v", err)
		return "", err
	}
	// Important to close the multipart writer to set the terminating boundary
	if err = multipartWriter.Close(); err != nil {
		log.Printf("Failed to close multipart writer: %v", err)
		return "", err
	}

	req, err := http.NewRequest("POST", ZephyrURL+"/redirect", &requestBody)
	if err != nil {
		log.Printf("Failed to create POST request: %v", err)
		return "", err
	}
	c := config.GetOrCreate()
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req.Header.Add("X-Upload-Token", c.Token)
	req.Header.Add("X-Upload-User", c.UserID)
	req.Header.Add("X-Upload-Host", c.Host)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send POST request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create redirect: %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body, error: %d", err)
	}

	return string(respBody), nil
}

type PastebinResponse struct {
	Key string `json:"key"`
}

func uploadText(data []byte) (string, error) {
	// Prepare the POST request
	reqBody := bytes.NewBufferString(string(data))
	resp, err := http.Post(PasteURL+"/documents", "text/plain", reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to make post request, error: %d", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to upload text, status code: %d", resp.StatusCode)
	}

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body, error: %d", err)
	}

	var result PastebinResponse
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body, error: %d", err)
	}
	log.Println("Key: " + result.Key)
	return PasteURL + "/" + result.Key, nil
}

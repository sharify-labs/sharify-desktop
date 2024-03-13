package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/ncruces/zenity"
	"golang.design/x/clipboard"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"zephyr-desktop/config"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("Zephyr")
	systray.AddMenuItem("Quit", "Quit the whole app")
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	mUpload := systray.AddMenuItem("Upload Clipboard", "Upload the image in your clipboard")
	mSettings := systray.AddMenuItem("Settings", "Modify settings")
	go func() {
		for {
			select {
			case <-mUpload.ClickedCh:
				uploadClipboard()
			case <-mSettings.ClickedCh:
				promptSettingsList()
			}
		}
	}()
	//mQuit := systray.AddMenuItem("Quit", "Quit the whole app")

	//Sets the icon of a menu item. Only available on Mac and Windows.
	//mQuit.SetIcon(icon.Data)
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
	_ = zenity.Info(
		fmt.Sprintf("Successfully updated %s!", field),
		zenity.Title("Success"),
		zenity.Icon(zenity.InfoIcon),
	)
}

func getAvailableHosts() ([]string, error) {
	c := config.GetOrCreate()
	// Prepare the GET request
	var requestBody bytes.Buffer

	req, err := http.NewRequest("GET", "https://xericl.dev/api/hosts", &requestBody)
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

func uploadClipboard() {
	var data []byte

	// Attempt to read image
	data = clipboard.Read(clipboard.FmtImage)
	if data != nil {
		resultURL, err := uploadImage(data)
		if err != nil {
			log.Printf("failed to upload image: %v", err)
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(resultURL))
		return
	}

	// Not image -> Attempt to read text
	data = clipboard.Read(clipboard.FmtText)
	if data != nil {
		resultURL, err := uploadText(data)
		if err != nil {
			log.Printf("failed to upload text: %v", err)
			return
		}
		clipboard.Write(clipboard.FmtText, []byte(resultURL))
	}

	// Clipboard read failed
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

	req, err := http.NewRequest("POST", "https://xericl.dev/upload", &requestBody)
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

type PastebinResponse struct {
	Key string `json:"key"`
}

const PasteURL string = "https://paste.crystaldev.co"

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

func onExit() {
	// clean up here
}

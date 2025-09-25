# Sharify-Desktop
Simple cross-platform desktop app for uploading content to Sharify.

Sharify-Desktop works alongside other Sharify components:
- **[Canvas](https://github.com/sharify-labs/canvas)**: Serves uploaded content to end users
- **[Spine](https://github.com/sharify-labs/spine)**: Web panel backend for user/content management
- **[Zephyr](https://github.com/sharify-labs/zephyr)**: Core upload API and storage service
- **[Sharify-Go](https://github.com/sharify-labs/sharify-go)**: Go SDK for programmatic uploads

_Note: This project is not actively maintained and should not be used in a production environment._

## What it does

Sharify-Desktop runs as a system tray application that uploads content to Sharify with a single click.<br>
It supports images, text snippets, and URLs.

## Key Features
- System tray integration with context menu
- One-click upload with instant URL replacement in clipboard
- Cross-platform packaging for Windows, macOS, and Linux (RPM)
- Desktop notifications for upload status

## How it works

```
Copy Content → System Tray Click → Read Clipboard → Upload to Sharify → Replace Clipboard with URL → Desktop Notification
```

The interesting parts:
- Cross-platform clipboard access with [golang-design/clipboard](https://github.com/golang-design/clipboard)
- Cross-platform desktop notifications with [ncruces/zenity](https://github.com/ncruces/zenity)
- Cross-platform system tray with [fyne.io/systray](https://github.com/fyne-io/systray)
- Communicates with the Sharify API using [sharify-go](https://github.com/sharify-labs/sharify-go)
- Config stored in OS-appropriate locations (AppData, Library, .config)

## Usage
Right-click the Sharify icon in your system tray:
- **Upload Clipboard** - Detects and uploads whatever is in your clipboard
- **Shorten URL** - Treats clipboard content as a URL to shorten
- **Settings** - Configure API token and preferred domain
- **Quit** - Exit the application

## Configuration

#### Settings Dialog
- **API Token**: Your Sharify upload token (format: `sfy_abc123_def456...`)
- **Host Domain**: Your preferred upload domain

#### Config File Locations
- **Windows**: `%AppData%\.sharifydesktop\config.json`
- **macOS**: `~/Library/Application Support/sharify-desktop/config.json`
- **Linux**: `~/.config/sharify-desktop/config.json`

## Development Setup
```bash
make build
make run
```
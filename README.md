# rapper

  <p align="left">
  <img alt="GitHub License" src="https://img.shields.io/github/license/anibaldeboni/rapper?logo=gnu">
  <a href="https://github.com/anibaldeboni/rapper/actions/workflows/master.yml" rel="nofollow">
    <img src="https://img.shields.io/github/actions/workflow/status/anibaldeboni/rapper/master.yml?branch=master&logo=Github" alt="Build" />
  </a>
  <img alt="GitHub language count" src="https://img.shields.io/github/languages/count/anibaldeboni/rapper?logo=go">
  <img alt="GitHub code size in bytes" src="https://img.shields.io/github/languages/code-size/anibaldeboni/rapper">
  <img href="https://github.com/anibaldeboni/rapper/releases/latest" alt="GitHub Release" src="https://img.shields.io/github/v/release/anibaldeboni/rapper?logo=semanticrelease">
  </p>

Rapper is a powerful, configurable CLI tool to perform multiple HTTP requests based on CSV files. It features an interactive Terminal User Interface (TUI), profile management, dynamic worker pools, and real-time metrics monitoring.

## Features

### 🎯 Multi-View TUI Interface
- **Files View**: Browse and select CSV files for processing
- **Logs View**: Real-time processing logs with scroll support
- **Settings View**: Edit configuration with live preview
- **Workers View**: Monitor and control worker pool dynamically

### 📋 Profile Management
- Support for multiple configuration profiles (dev, staging, production, etc.)
- Quick profile switching with `Ctrl+P`
- Visual profile selector with active profile indicator
- Each profile stored as separate YAML file

### ⚙️ Configuration Editor
- In-app configuration editing
- Form fields for URL template, request body, and headers
- Tab navigation between fields
- Save changes with `Ctrl+S`
- Real-time validation and unsaved changes indicator

### 👷 Dynamic Worker Pool
- Adjust worker count in real-time with arrow keys or +/-
- Visual slider for worker count (1 to CPU count)
- Instant feedback without restarting

### 📊 Real-Time Metrics
- Processing status indicator
- Total requests, success/error counts
- Lines processed from CSV
- Throughput (requests per second)
- Elapsed time during processing
- Active workers count
- Auto-refresh every 500ms

### 🎨 Visual Polish
- Toast notifications for important actions
- Color-coded metrics (green for success, red for errors)
- Smooth animations and transitions
- Enhanced modal dialogs
- Responsive layout

## Installing

We provide pre-compiled binaries for Linux and MacOS (amd64 and arm64). The latest release could be found [here](https://github.com/anibaldeboni/github.com/anibaldeboni/rapper/releases/latest). After downloading a suitable binary for your system and architecture follow the commands:

```shell
chmod +x rapper-linux-amd64
mv ./rapper-linux-amd64 ~/.local/bin/rapper
```

The instructions above move the binary to `~./local/bin` with the name `rapper` if you have another folder mapped in `$PATH` move the app to the pertinent location.

## Configuration

Prior to running `rapper` you must set a `config.yml` structure is as follow:

```yaml
token: "JWT of your application"
path:
  method: PUT # HTTP method you wish to be used in requests (currently supports PUT and POST)
  template: https://api.myapp.io/users/{{.id}} # the variables are replaced by the corresponding csv values
payload: # a json template to be filled with variables extracted from the CSV
  template: |
    {
      "address": {
        "street": "{{.street_name}}", # the variables are replaced by the corresponding csv values
        "number": "{{.house_number}}",
        "city": "{{.city}}"
      }
    }
csv:
  fields: # The fields you want to use from the CSV, if none will use all
    - id
    - street_number
    - house_number
    - city
  separator: "," # the separator used in the CSV, if not specified will use comma
```

Have in mind that when a request fails all variables selected in `csv` field will be used to form the error message, so select all variables you need to form the url and payload and any other that is relevant to identify problems when an error occur

## Keyboard Shortcuts

### Global Navigation
- `Ctrl+F`: Switch to Files view
- `Ctrl+L`: Switch to Logs view
- `Ctrl+T`: Switch to Settings view
- `Ctrl+W`: Switch to Workers view
- `Esc`: Go back / Cancel operation
- `Ctrl+C` / `q`: Quit application
- `?`: Toggle help

### Settings View
- `Tab` / `Shift+Tab`: Navigate between form fields
- `Ctrl+S`: Save configuration
- `Ctrl+P`: Open profile selector
- Arrow keys in form: Edit text
- `↑` / `↓`: Navigate profile list (when profile selector is open)
- `Enter`: Select profile (when profile selector is open)

### Workers View
- `←` / `→` or `h` / `l`: Decrease/increase worker count
- `-` / `+`: Decrease/increase worker count

### Files & Logs View
- `↑` / `↓`: Navigate file list / Scroll logs
- `Enter`: Select file for processing
- `Esc`: Cancel processing (Files view only)

## Usage

All options are available via `rapper -h`

You may run `rapper` directly in a directory containing a `config.yml` or `config.yaml` and CSV files to process. Or setting the options:

```shell
  -config string
    	path to directory containing a config file (default current working dir)
  -dir string
    	path to directory containing the CSV files (default current working dir)
  -output string
    	path to output file, including the file name
  -workers int
    	number of request workers (max: 5) (default 1)
```

A little demo of the app execution:
![rapper usage recording](./assets/rapper.gif)

# Development

In the project root directory you will find a `Makefile` with all available commands.

### Building

```shell
make build
```

After that, just copy the binary to a location mapped in the `$PATH` environment variable

### Lint

Make sure you have `golangci-lint` installed. More instruction on how to install could be found [here](https://golangci-lint.run/usage/install/)

```shell
make lint
```

### Tsting

For test assertions we use [testify](https://github.com/stretchr/testify) and [gomock](https://go.uber.org/mock) for test mocks generation.

```shell
make test
```

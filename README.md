# rapper

Rapper is a configurable cli tool to perform multiple HTTP requests based on a CSV file containing data.

## Installing

We provide pre-compiled binaries for Linux and MacOS (amd64 and arm64). The latest release could be found [here](https://github.com/anibaldeboni/github.com/anibaldeboni/rapper/releases/latest)
After downloading the suitable binary to your system and architecture follow the commands:

```
chmod +x rapper-linux-amd64
mv ./rapper-linux-amd64 ~/.local/bin/rapper
```

The instructions above move the binary to `~./local/bin` with the name `rapper` if you have another folder mapped in `$PATH` move the app to the pertinent location.

## Configuration

Prior to running `rapper` you must set a `config.yml` file inside the directory containing the CSV files you want to send data. The `config.yml` structure is as follow:

```
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
  fields: # The fields you want to use from the CSV
    - id
    - street_number
    - house_number
    - city
  separator: "," # the separator used in the CSV, if not specified will use comma
```

Have in mind that when a request fails all variables selected in `csv` field will be used to form the error message, so select all variables you need to form the url and payload and any other that is relevant to identify problems when an error occur

## Usage

You may run `rapper` directly in a directory containing a `config.yml` and CSV files to process

```
cd ~/folder-with-csv
rapper
```

Or poiting the app to the proper directory

```
rapper ~/some-folder
```

Then you may follow the instructions in your screen.
![rapper usage recording](./assets/rapper.gif)

# Development

In the project root directory you will find a `Makefile` with all available commands.

### Building

```
make build
```

After that, just copy the binary to a location mapped in the `$PATH` environment variable

### Lint

Make sure you have `golangci-lint` installed. More instruction on how to install could be found [here](https://golangci-lint.run/usage/install/)

```
make lint
```

### Tsting

For test assertions we use [testify](https://github.com/stretchr/testify) and [vektra/mockery](https://github.com/vektra/mockery) for test mocks generation.

```
make test
```

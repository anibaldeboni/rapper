# rapper

Rapper is a configurable cli tool to perform multiple HTTP requests based on a CSV file containing data.

## Installing

We provide pre-compiled binaries for Linux and MacOS (amd64 and arm64). The latest release could be found [here](https://github.com/anibaldeboni/rapper/releases/latest)
After downloading the apropriate binary to your system and architecture follow the commands:

```
chmod +x rapper-linux-amd64
mv ./rapper-linux-amd64 ~/.local/bin
```

The instructions above move the binary to `~./local/bin` if you have another folder mapped in `$PATH` move the app to the proper location.

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
csv: # The column names you want to use from the CSV
  - id
  - street_name
  - house_number
  - city
```

## Usage

```
cd ~/folder-with-csv
rapper
```

Then you may follow the instructions in your screen.

## Building

In the project root directory you will find a `build.sh` script, just run in your terminal:

```
./build.sh
```

After that, just copy the binary to a location mapped in the `$PATH` environment variable

# rapper

Rapper is a configurable cli tool to perform multiple HTTP requests based on a CSV file containing data.

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
csv: # The fields you want to use from the CSV
  - id
  - street_number
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

set -e
APP=${1:-"rapper"}
APP_NAME="github.com/anibaldeboni/rapper/internal/ui.AppName=$APP"
APP_VERSION="github.com/anibaldeboni/rapper/internal/ui.AppVersion=$(git rev-parse --short HEAD)"
BUILD_OUTPUT="./build/$APP"

go get -v ./...
go build -o $BUILD_OUTPUT -ldflags "-s -w -X '$APP_NAME' -X '$APP_VERSION'" ./main.go
echo "\nBuild output: \033[1m$BUILD_OUTPUT\033[0m\n"

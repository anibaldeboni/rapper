set -e
APP=${1:-"rapper"}
BUILD_OUTPUT="./build/$APP"

go get -v ./...
# Build with VCS stamping enabled (default in Go 1.18+)
# Version info is automatically embedded via debug.ReadBuildInfo()
go build -buildvcs=true -o $BUILD_OUTPUT -ldflags "-s -w" ./main.go
echo "\nBuild output: \033[1m$BUILD_OUTPUT\033[0m\n"


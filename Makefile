APP_NAME=weather-checker

default: run

build:
	go build main.go

run:
	go run $(APP_NAME)

clean:
	rm -f $(APP_NAME)
	go clean

test:
	go test -v ./...

NAME=bing-bong

.PHONY: clean

all: clean build

dev:
	go run main.go

build:
	GOARCH=amd64 GOOS=linux go build -trimpath -ldflags '-w -s' -o build/$@/$(NAME)

clean:
	rm -rf ./build
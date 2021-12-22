NAME=bing-bong

.PHONY: clean

all: clean release

release: build
	zip -q -r build/release.zip build/

dev:
	go run main.go

build:
	GOARCH=amd64 GOOS=linux go build -trimpath -ldflags '-w -s' -o build/$(NAME)
	cp config.yml build/

clean:
	rm -rf ./build
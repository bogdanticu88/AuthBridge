.PHONY: build run test clean

build:
	go build -o authbridge ./main.go

run:
	go run ./main.go start

test:
	go test ./... -v

clean:
	rm -f authbridge

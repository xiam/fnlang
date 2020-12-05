clean:
	rm -rf ./bin

build: clean
	go build -o ./bin/fn ./cmd/fn

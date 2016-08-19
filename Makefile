build:	vendor
	go build

test:	vendor build
	glide nv | xargs -n 1 go test -cover

vendor:
	glide install

.PHONY: build test


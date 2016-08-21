TAGS = irc

all: bin/glenda

bin/%: vendor
	go build -tags $(TAGS) -o $@ ./cmd/$*

test:	vendor bin/glenda
	glide nv | xargs -n 1 go test -cover

vendor:
	glide install

clean:
	rm -f bin/*

.PHONY: all test clean


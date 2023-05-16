PROJECT=go-gnudip
BINARY=$(PROJECT)
ENVFILE=.env

-include $(ENVFILE)
export

all:
	go build -o $(BINARY) .

run:
	go run .

watch:
	find . -name '*.go' -o -name $(ENVFILE) | entr -r make run

clean:
	rm -f $(BINARY)

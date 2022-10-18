.PHONY: bmp

include version.mk

all:
	@go build -ldflags "all=$(GO_LDFLAGS)" github.com/matm/bmp/cmd/bmp

clean:
	rm -f bmp

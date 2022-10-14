.PHONY: bmp

all:
	go build github.com/matm/bmp/cmd/bmp

clean:
	rm -f bmp

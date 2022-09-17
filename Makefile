.PHONY: bsp

all:
	go build github.com/matm/bsp/cmd/bsp

clean:
	rm -f bsp

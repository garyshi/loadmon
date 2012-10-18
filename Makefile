TARGETS=loadmon

all: $(TARGETS)

clean:
	rm -f $(TARGETS)

loadmon: $(wildcard src/*.go)
	go build -o $@ $^

TARGETS=loadmon

all: $(TARGETS)

clean:
	rm -f $(TARGETS)

loadmon:
	go build -o $@ src/*.go

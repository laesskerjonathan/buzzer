SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
BINARY=buzzer-server

build: $(BINARY)

$(BINARY): $(SOURCES)
	go build -o ${BINARY}

clean:
	rm -f ${BINARY}

.PHONY: build clean

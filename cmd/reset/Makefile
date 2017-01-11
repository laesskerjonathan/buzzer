SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
BINARY=reset

build: $(BINARY)

$(BINARY): $(SOURCES)
	go build -o ${BINARY}

clean:
	rm -f ${BINARY}

.PHONY: build clean

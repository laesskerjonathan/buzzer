SRCDIR=.
BINDIR=/home/pi/bin

SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
BINARY=pticker

build: $(BINARY)

$(BINARY): $(SOURCES)
	go build -o ${BINARY}

install: clean build
	install -m 755 ${BINARY} ${BINDIR}
	sudo  install -m 644 ${BINARY}.service /etc/systemd/system
	sudo systemctl daemon-reload && sudo systemctl enable $(BINARY).service

clean:
	rm -f ${BINARY}

.PHONY: build install clean

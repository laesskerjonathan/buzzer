SRCDIR=.
BINDIR=/home/pi/bin

SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
BINARY=pbuzzer

build: $(BINARY)

$(BINARY): $(SOURCES)
	go build -o ${BINARY}

install: clean build
	install -m 755 ${BINARY} ${BINDIR}
	install start-${BINARY}.sh ${BINDIR}
	sed -i '/start-${BINARY}/d' /home/pi/.config/lxsession/LXDE-pi/autostart
	echo "@${BINDIR}/start-${BINARY}.sh" >> /home/pi/.config/lxsession/LXDE-pi/autostart

clean:
	rm -f ${BINARY}

.PHONY: install build clean

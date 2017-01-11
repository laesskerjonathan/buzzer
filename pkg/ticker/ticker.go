package ticker

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/tarm/serial"
)

/*
	Start Packet:
	0x01 0x5A 0x30 0x30 0x02 0x41 0x41 0x1B 0x20  0x62  0x20  0x41 0x42 0x43  0x04
	|-------------------------------------------| |---| |---| |-------------| |---|
	|                     A                     | | B | | C | |      D      | | E |

	A: Packet Header
	B: Modifier Symbol in ASCII
		a: Rotate
		b: Fixed
		c: Flash
		e: Roll-Up
		f: Roll-Down
		g: Roll-Left
		h: Roll-Right
		i: Wipe-Up
		j: Wipe-Down
	C: Data Header
	D: Data in ASCII encoding (example "ABC")
	E: EOT


	Stop Packet:
	0x01 0x5A 0x30 0x30 0x02 0x41 0x41 0x1B 0x20  0x61  0x20  0x04
	|-------------------------------------------| |---| |---| |---|
	|                     A                     | | B | | C | | E |

	The stop packet is equivalent to the start packet with modifier 'a' and without data
*/

const (
	packetHeader = "\x01\x5a\x30\x30\x02\x41\x41\x1b\x20"
	stopModifier = "\x61"
	dataHeader   = "\x20"
	eot          = "\x04"
)

//
type Ticker struct {
	devName string
	port    *serial.Port
	effect  string
}

//
func NewTicker(name string) (*Ticker, error) {
	t := &Ticker{devName: name, effect: "\x61"}
	if err := t.open(); err != nil {
		return nil, err
	}
	return t, nil
}

func (t *Ticker) open() error {
	c := &serial.Config{Name: t.devName, Baud: 9600}
	port, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	t.port = port
	return nil
}

//
func (t *Ticker) Effect(name string) error {
	switch name {
	case "rotate":
		t.effect = "\x61"
	case "fixed":
		t.effect = "\x62"
	case "flash":
		t.effect = "\x63"
	case "rollUp":
		t.effect = "\x65"
	case "rollDown":
		t.effect = "\x66"
	case "rollLeft":
		t.effect = "\x67"
	case "rollRight":
		t.effect = "\x68"
	case "wipeUp":
		t.effect = "\x69"
	case "wipeDown":
		t.effect = "\x6a"
	default:
		return errors.New(fmt.Sprintf("no such effect: %s", name))
	}
	return nil
}

//
func (t *Ticker) Start(text string) error {
	var buf bytes.Buffer
	buf.WriteString(packetHeader)
	buf.WriteString("a")
	buf.WriteString(dataHeader)
	buf.WriteString(text)
	buf.WriteString(eot)
	_, err := t.port.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

//
func (t *Ticker) Stop() error {
	var buf bytes.Buffer
	buf.WriteString(packetHeader)
	buf.WriteString(stopModifier)
	buf.WriteString(dataHeader)
	buf.WriteString(eot)
	_, err := t.port.Write(buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}

//
func (t *Ticker) Update(data fmt.Stringer) error {
	text := data.String()
	if err := t.Stop(); err != nil {
		return err
	}
	if err := t.Start(text); err != nil {
		return err
	}
	return nil
}

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gvalkov/golang-evdev"
)

const (
	KEY_0     = 82
	KEY_1     = 79
	KEY_2     = 80
	KEY_3     = 81
	KEY_4     = 75
	KEY_5     = 76
	KEY_6     = 77
	KEY_7     = 71
	KEY_8     = 72
	KEY_9     = 73
	KEY_ENTER = 96
)

//
type Keypad struct {
	name string
	dev  *evdev.InputDevice
	stop bool
	code chan string
}

//
func NewKeypad(name string) (*Keypad, error) {
	k := &Keypad{
		name: name,
		code: make(chan string),
	}
	devs, err := evdev.ListInputDevicePaths("/dev/input/event*")
	if err != nil {
		return nil, err
	}
	for _, d := range devs {
		dev, err := evdev.Open(d)
		if err != nil {
			return nil, err
		}
		if strings.Contains(dev.Name, k.name) {
			k.dev = dev
			break
		}
	}
	if k.dev == nil {
		return nil, errors.New("no appropriate device found")
	}
	return k, nil
}

//
func (k *Keypad) Start() <-chan string {
	go func() {
		var c string
		ticker := time.NewTicker(time.Millisecond * 50)
		for !k.stop {
			ev, err := k.dev.ReadOne()
			if err != nil {
				log.Fatal(err)
			}
			// proceed only with key events
			if ev.Type != evdev.EV_KEY {
				continue
			}
			kev := evdev.NewKeyEvent(ev)
			// proceed only with key down events
			if kev.State != evdev.KeyDown {
				continue
			}
			// evaluate scan code
			switch kev.Scancode {
			case KEY_ENTER:
				k.code <- c
				c = ""
			case KEY_1:
				c = fmt.Sprintf("%s1", c)
			case KEY_2:
				c = fmt.Sprintf("%s2", c)
			case KEY_3:
				c = fmt.Sprintf("%s3", c)
			case KEY_4:
				c = fmt.Sprintf("%s4", c)
			case KEY_5:
				c = fmt.Sprintf("%s5", c)
			case KEY_6:
				c = fmt.Sprintf("%s6", c)
			case KEY_7:
				c = fmt.Sprintf("%s7", c)
			case KEY_8:
				c = fmt.Sprintf("%s8", c)
			case KEY_9:
				c = fmt.Sprintf("%s9", c)
			case KEY_0:
				c = fmt.Sprintf("%s0", c)
			}
			<-ticker.C
		}
	}()
	return k.code
}

//
func (k *Keypad) Stop() {
	k.stop = true
	// TODO: wait
	return
}

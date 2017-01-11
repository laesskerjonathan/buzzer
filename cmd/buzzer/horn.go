package main

import (
	"time"

	"github.com/luismesas/goPi/piface"
)

const (
	HornRelay  = 1
	HornButton = 3
)

//
type Horn struct {
	pfd     *piface.PiFaceDigital
	unwatch bool
}

//
func NewHorn(pfd *piface.PiFaceDigital) *Horn {
	return &Horn{
		pfd: pfd,
	}
}

//
func (h *Horn) On() {
	h.pfd.Relays[HornRelay].AllOn()
}

//
func (h *Horn) Off() {
	h.pfd.Relays[HornRelay].AllOff()
}

//
func (h *Horn) WatchButton() {
	go func() {
		ticker := time.NewTicker(time.Millisecond * 50)
		for !h.unwatch {
			if h.pfd.Switches[HornButton].Value() == byte(0) {
				h.Off()
			}
			<-ticker.C
		}
	}()
}

//
func (h *Horn) UnwatchButton() {
	h.unwatch = true
}

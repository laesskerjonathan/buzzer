package main

import (
	"time"

	"github.com/luismesas/goPi/piface"
)

const (
	LightRelay  = 0
	LightButton = 2
)

//
type Light struct {
	pfd     *piface.PiFaceDigital
	unwatch bool
}

//
func NewLight(pfd *piface.PiFaceDigital) *Light {
	return &Light{
		pfd: pfd,
	}
}

//
func (l *Light) On() {
	l.pfd.Relays[LightRelay].AllOn()
}

//
func (l *Light) Off() {
	l.pfd.Relays[LightRelay].AllOff()
}

//
func (l *Light) WatchButton() {
	go func() {
		ticker := time.NewTicker(time.Millisecond * 50)
		for !l.unwatch {
			if l.pfd.Switches[LightButton].Value() == byte(0) {
				l.Off()
			}
			<-ticker.C
		}
	}()
}

//
func (l *Light) UnwatchButton() {
	l.unwatch = true
}

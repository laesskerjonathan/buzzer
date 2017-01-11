package main

import (
	"time"

	"github.com/luismesas/goPi/piface"
)

//
type Buzzer struct {
	pfd  *piface.PiFaceDigital
	stop bool
}

//
func NewBuzzer(pfd *piface.PiFaceDigital) *Buzzer {
	return &Buzzer{
		pfd: pfd,
	}
}

//
func (b *Buzzer) Watch() {
	// is somebody sitting on the buzzer?
	ticker := time.NewTicker(time.Millisecond * 50)
	for b.pfd.Switches[0].Value() == byte(0) && !b.stop {
		<-ticker.C
	}
	// wait for pressing the buzzer
	for b.pfd.Switches[0].Value() != byte(0) && !b.stop {
		<-ticker.C
	}
}

//
func (b *Buzzer) Unwatch() {
	b.stop = true
}

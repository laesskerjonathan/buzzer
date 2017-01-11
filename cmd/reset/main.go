package main

import (
	"fmt"

	"github.com/luismesas/goPi/piface"
	"github.com/luismesas/goPi/spi"
)

func main() {

	// creates a new pifacedigital instance
	pfd := piface.NewPiFaceDigital(spi.DEFAULT_HARDWARE_ADDR, spi.DEFAULT_BUS, spi.DEFAULT_CHIP)

	// initializes pifacedigital board
	err := pfd.InitBoard()
	if err != nil {
		fmt.Printf("Error on init board: %s", err)
		return
	}

	for i := 0; i < 8; i++ {
		pfd.Leds[i].AllOff()
	}
	pfd.Relays[0].AllOff()
	pfd.Relays[1].AllOff()
}

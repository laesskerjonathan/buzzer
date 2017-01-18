package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/marcsauter/buzzer/pkg/pitch"

	"github.com/luismesas/goPi/piface"
	"github.com/luismesas/goPi/spi"
)

func main() {
	pin := os.Getenv("BUZZER_PIN")
	if len(pin) == 0 {
		pin = "1982"
		log.Fatal("BUZZER_PIN missing using default")
	}
	device := os.Getenv("BUZZER_KEYPAD_DEVICE")
	if len(device) == 0 {
		log.Fatal("BUZZER_KEYPAD_DEVICE missing or not valid")
	}
	url, err := url.Parse(os.Getenv("BUZZER_PITCH_URL"))
	if err != nil || !url.IsAbs() {
		log.Fatal("BUZZER_PITCH_URL missing or not valid")
	}
	interval, err := strconv.Atoi(os.Getenv("BUZZER_PITCH_CHECK_INTERVAL"))
	if err != nil {
		log.Fatal("BUZZER_PITCH_CHECK_INTERVAL missing or not valid")
	}

	// creates a new pifacedigital instance
	pfd := piface.NewPiFaceDigital(spi.DEFAULT_HARDWARE_ADDR, spi.DEFAULT_BUS, spi.DEFAULT_CHIP)

	// initializes pifacedigital board
	err = pfd.InitBoard()
	if err != nil {
		fmt.Printf("Error on init board: %s", err)
		return
	}
	b := NewBuzzer(pfd)
	h := NewHorn(pfd)
	h.WatchButton()
	l := NewLight(pfd)
	l.WatchButton()
	k, err := NewKeypad(device)
	if err != nil {
		log.Fatal(err)
	}
	//
	s := NewScreen()
	s.Init("buzzer", "Pitch Info", "Pitch Info")
	s.Main()
	s.StartTicker()
	//
	p := pitch.NewPitch(url)
	p.StartCheckNext(interval, s)
	//
	code := k.Start()
	cancel := make(chan os.Signal, 1)
	signal.Notify(cancel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	for {
		select {
		case c := <-code:
			if c == pin {
				s.Keypad(fmt.Sprintf(DefaultKeypadText, "ERROR: invalid PIN"))
			} else {
				s.Keypad(fmt.Sprintf("PIN valid - Please press the Buzzer to release the Pitch ...\n"))
				b.Watch()
				l.On() // light on
				h.On() // horn on
			}
		case <-cancel:
			p.StopCheckNext()
			b.Unwatch()
			l.UnwatchButton()
			l.Off()
			h.UnwatchButton()
			h.Off()
			k.Stop()
			s.StopCountdown()
			s.Destroy()
			log.Fatalln("signal received - exiting")
		}
	}
}
